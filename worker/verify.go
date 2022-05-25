package worker

import (
	"errors"
	"sync/atomic"
	"time"

	"github.com/weijun-sh/rsyslog/cmd/utils"
	"github.com/weijun-sh/rsyslog/common"
	"github.com/weijun-sh/rsyslog/mongodb"
	"github.com/weijun-sh/rsyslog/params"
	"github.com/weijun-sh/rsyslog/router"
	"github.com/weijun-sh/rsyslog/tokens"
	mapset "github.com/deckarep/golang-set"
)

var (
	verifySwapCh      = make(chan *mongodb.MgoSwap, 10)
	maxVerifyRoutines = int64(10)
	curVerifyRoutines = int64(0)

	cachedVerifyingSwaps    = mapset.NewSet()
	maxCachedVerifyingSwaps = 100
)

// StartVerifyJob verify job
func StartVerifyJob() {
	logWorker("verify", "start router swap verify job")

	go startVerifyProducer()

	mongodb.MgoWaitGroup.Add(1)
	go startVerifyConsumer()
}

func startVerifyProducer() {
	for {
		septime := getSepTimeInFind(maxVerifyLifetime)
		res, err := mongodb.FindRouterSwapsWithStatus(mongodb.TxNotStable, septime)
		if err != nil {
			logWorkerError("verify", "find router swap error", err)
		}
		if len(res) > 0 {
			logWorker("verify", "find router swap to verify", "count", len(res))
		}
		for _, swap := range res {
			if utils.IsCleanuping() {
				logWorker("verify", "stop router swap verify job")
				return
			}
			if cachedVerifyingSwaps.Contains(swap.Key) {
				logWorkerTrace("verify", "ignore cached verifying swap before dispatch", "key", swap.Key)
				continue
			}
			logWorker("verify", "dispatch swap for verify", "fromChainID", swap.FromChainID, "toChainID", swap.ToChainID, "txid", swap.TxID, "logIndex", swap.LogIndex)
			verifySwapCh <- swap // produce
		}
		restInJob(restIntervalInVerifyJob)
	}
}

func startVerifyConsumer() {
	defer mongodb.MgoWaitGroup.Done()
	for {
		select {
		case <-utils.CleanupChan:
			logWorker("verify", "stop verify swap job")
			return
		case swap := <-verifySwapCh: // consume
			// loop and check, break if free worker exist
			for {
				if atomic.LoadInt64(&curVerifyRoutines) < maxVerifyRoutines {
					break
				}
				time.Sleep(1 * time.Second)
			}

			atomic.AddInt64(&curVerifyRoutines, 1)
			go func() {
				_ = processRouterSwapVerify(swap)
			}()
		}
	}
}

func isBlacked(swap *mongodb.MgoSwap) bool {
	return params.IsChainIDInBlackList(swap.FromChainID) ||
		params.IsChainIDInBlackList(swap.ToChainID) ||
		params.IsTokenIDInBlackList(swap.GetTokenID()) ||
		params.IsAccountInBlackList(swap.From) ||
		params.IsAccountInBlackList(swap.Bind) ||
		params.IsAccountInBlackList(swap.TxTo)
}

//nolint:funlen,gocyclo // ok
func processRouterSwapVerify(swap *mongodb.MgoSwap) (err error) {
	defer atomic.AddInt64(&curVerifyRoutines, -1)

	if router.IsChainIDPaused(swap.FromChainID) || router.IsChainIDPaused(swap.ToChainID) {
		return nil
	}

	fromChainID := swap.FromChainID
	txid := swap.TxID
	logIndex := swap.LogIndex

	if cachedVerifyingSwaps.Contains(swap.Key) {
		logWorkerTrace("verify", "ignore cached verifying swap before dispatch", "key", swap.Key)
		return nil
	}
	if cachedVerifyingSwaps.Cardinality() >= maxCachedVerifyingSwaps {
		cachedVerifyingSwaps.Pop()
	}
	cachedVerifyingSwaps.Add(swap.Key)
	isProcessed := true
	defer func() {
		if !isProcessed {
			cachedVerifyingSwaps.Remove(swap.Key)
		}
	}()

	var dbErr error
	if isBlacked(swap) {
		err = tokens.ErrSwapInBlacklist
		dbErr = mongodb.UpdateRouterSwapStatus(fromChainID, txid, logIndex, mongodb.SwapInBlacklist, now(), err.Error())
		if dbErr != nil {
			logWorkerError("verify", "verify router swap db error", dbErr, "fromChainID", fromChainID, "toChainID", swap.ToChainID, "txid", txid, "logIndex", logIndex)
		}
		return err
	}

	bridge := router.GetBridgeByChainID(fromChainID)
	if bridge == nil {
		return tokens.ErrNoBridgeForChainID
	}

	logWorker("verify", "process swap verify", "fromChainID", fromChainID, "toChainID", swap.ToChainID, "txid", swap.TxID, "logIndex", swap.LogIndex)

	verifyArgs := &tokens.VerifyArgs{
		SwapType:      tokens.SwapType(swap.SwapType),
		LogIndex:      logIndex,
		AllowUnstable: false,
	}
	swapInfo, err := bridge.VerifyTransaction(txid, verifyArgs)
	switch {
	case err == nil:
		if router.IsBigValueSwap(swapInfo) {
			dbErr = mongodb.UpdateRouterSwapStatus(fromChainID, txid, logIndex, mongodb.TxWithBigValue, now(), "big swap value")
		} else {
			dbErr = mongodb.PassRouterSwapVerify(fromChainID, txid, logIndex, now())
			if dbErr == nil {
				dbErr = AddInitialSwapResult(swapInfo, mongodb.MatchTxEmpty)
			}
		}
	case errors.Is(err, tokens.ErrTxNotStable),
		errors.Is(err, tokens.ErrRPCQueryError):
		isProcessed = false
		return err
	case errors.Is(err, tokens.ErrTxNotFound),
		errors.Is(err, tokens.ErrNotFound):
		nowMilli := common.NowMilli()
		if swap.InitTime+1000*maxTxNotFoundTime < nowMilli {
			duration := time.Duration((nowMilli - swap.InitTime) / 1000 * int64(time.Second))
			logWorker("verify", "set longer not found swap to verify failed", "fromChainID", fromChainID, "toChainID", swap.ToChainID, "txid", swap.TxID, "logIndex", swap.LogIndex, "inittime", swap.InitTime, "duration", duration.String())
			dbErr = mongodb.UpdateRouterSwapStatus(fromChainID, txid, logIndex, mongodb.TxVerifyFailed, now(), err.Error())
			_ = mongodb.UpdateRouterSwapResultStatus(fromChainID, txid, logIndex, mongodb.TxVerifyFailed, now(), err.Error())
		} else {
			isProcessed = false
			return err
		}
	case errors.Is(err, tokens.ErrTxWithWrongValue):
		dbErr = mongodb.UpdateRouterSwapStatus(fromChainID, txid, logIndex, mongodb.TxWithWrongValue, now(), err.Error())
	case errors.Is(err, tokens.ErrTxWithWrongPath):
		dbErr = mongodb.UpdateRouterSwapStatus(fromChainID, txid, logIndex, mongodb.TxWithWrongPath, now(), err.Error())
	case errors.Is(err, tokens.ErrMissTokenConfig):
		dbErr = mongodb.UpdateRouterSwapStatus(fromChainID, txid, logIndex, mongodb.MissTokenConfig, now(), err.Error())
	case errors.Is(err, tokens.ErrNoUnderlyingToken):
		dbErr = mongodb.UpdateRouterSwapStatus(fromChainID, txid, logIndex, mongodb.NoUnderlyingToken, now(), err.Error())
	default:
		dbErr = mongodb.UpdateRouterSwapStatus(fromChainID, txid, logIndex, mongodb.TxVerifyFailed, now(), err.Error())
	}

	if dbErr != nil {
		logWorkerError("verify", "verify router swap db error", dbErr, "fromChainID", fromChainID, "toChainID", swap.ToChainID, "txid", txid, "logIndex", logIndex)
	}

	if err != nil {
		logWorkerError("verify", "verify router swap error", err, "fromChainID", fromChainID, "toChainID", swap.ToChainID, "txid", swap.TxID, "logIndex", swap.LogIndex)
	}

	return err
}

// DeleteCachedVerifyingSwap delete cached verifying swap
func DeleteCachedVerifyingSwap(key string) {
	cachedVerifyingSwaps.Remove(key)
}
