// Package rpcapi provides JSON RPC service.
package rpcapi

import (
	//"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	checktxcommon "github.com/weijun-sh/checkTx-server/common"
	"github.com/weijun-sh/checkTx-server/internal/swapapi"
	"github.com/weijun-sh/checkTx-server/params"
	"github.com/weijun-sh/checkTx-server/tokens"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	//"github.com/davecgh/go-spew/spew"
)

const (
	swapinTopic = iota + 1
	swapoutTopic
	routerTopic
)

type ResultStatus struct {
	Code uint64 `json:"code"`
	Msg string `json:"msg"`
	Data map[string]*GetStatusInfoResult `json:"data"`
}

type GetStatusInfoResult map[string]interface{}

// RPCQueryHistoryArgs args
type RPCQueryHistoryArgs struct {
	Bridge  string `json:"bridge"`
        Address string `json:"address"`
        PairID  string `json:"pairid"`
        Offset  int    `json:"offset"`
        Limit   int    `json:"limit"`
        Status  string `json:"status"`
	Timestamp string `json:"timestamp"`
}

// GetRouterStatusInfo api
func (s *RPCAPI) GetRouterStatusInfo(r *http.Request, args *RPCQueryHistoryArgs, result *ResultStatus) error {
	fmt.Printf("[rpcapi]GetRouterStatusInfo, args: %v\n", args)
	return GetStatusInfo(args, result, true)
}

// GetBridgeStatusInfo api
func (s *RPCAPI) GetBridgeStatusInfo(r *http.Request, args *RPCQueryHistoryArgs, result *ResultStatus) error {
	fmt.Printf("[rpcapi]GetBridgeStatusInfo, args: %v\n", args)
	return GetStatusInfo(args, result, false)
}

func GetStatusInfo(args *RPCQueryHistoryArgs, result *ResultStatus, isrouter bool) error {
	result.Code = 0
	result.Msg = ""
	result.Data = make(map[string]*GetStatusInfoResult, 0)
	dbname := args.Bridge
	status := args.Status
	if status == "" {
		status = "0,8,9,10,12,14,17" // default
	}
	if dbname == "all" {
		if isrouter {
			dbnames := params.GetRouterDbName()
		        for _, dbname := range dbnames {
		                getStatusInfo(*dbname, status, result)
		        }
		} else {
			dbnames := params.GetBridgeDbName()
		        for _, dbname := range dbnames {
		                getBridgeStatusInfo(dbname, status, result)
		        }
		}
	} else {
		dbname = params.SetRouterDbname_0(dbname)
		getStatusInfo(dbname, status, result)
	}
	return nil
}

func getStatusInfo(dbname, status string, result *ResultStatus) {
	fmt.Printf("getStatusInfo, dbname: %v\n", dbname)
	res, err := swapapi.GetStatusInfo(dbname, status)
	if err == nil && len(res) != 0 {
		var s GetStatusInfoResult
		s = res
		dbname = params.UpdateRouterDbname_0(dbname)
		result.Data[dbname] = &s
	}
}

func getBridgeStatusInfo(dbname, status string, result *ResultStatus) {
	fmt.Printf("getBridgeStatusInfo, dbname: %v\n", dbname)
	res, err := swapapi.GetBridgeStatusInfo(dbname, status)
	if err == nil && len(res) != 0 {
		var s GetStatusInfoResult
		s = res
		result.Data[dbname] = &s
	}
}

func (s *RPCAPI) GetSwapNotStable(r *http.Request, args *RPCNullArgs, result *ResultHistorySwap) error {
	fmt.Printf("[rpcapi]GetSwapNotStable, args: %v\n", args)
	var argsH RPCQueryHistoryArgs
	argsH.Bridge = "all"
	argsH.Status = "0,8,9,12,14,17" // default
	return s.GetSwapHistory(r, &argsH, result)
}

func (s *RPCAPI) GetSwapHistory(r *http.Request, args *RPCQueryHistoryArgs, result *ResultHistorySwap) error {
	fmt.Printf("[rpcapi]GetSwapHistory, args: %v\n", args)
	result.Code = 0
	result.Msg = ""
	result.Data = make(map[string][]interface{}, 0)
	dbname := args.Bridge
	status := args.Status
	if status == "" {
		status = "0,8,9,12,14,17" // default
	}
	if dbname == "all" {
		dbnames := params.GetRouterDbName()
		for _, dbname := range dbnames {
			getSwapHistory(*dbname, status, result)
		}
	} else {
		dbname = params.SetRouterDbname_0(dbname)
		getSwapHistory(dbname, status, result)
	}
	return nil
}

func getSwapHistory(dbname, statuses string, result *ResultHistorySwap) {
	fmt.Printf("getSwapHistory, dbname: %v\n", dbname)
	parts := strings.Split(statuses, ",")
	var s []interface{} = make([]interface{}, 0)
	var getH bool
	for _, status := range parts {
		if status == "10" {
			continue
		}
		si, errs := swapapi.GetRouterSwapHistory(dbname, "", "", 0, 20, status)
		if errs == nil && len(si) != 0 {
			for _, st := range si {
				s = append(s, st)
				getH = true
				//spew.Printf("%v\n", st)
			}
		}
	}
	if getH {
		var bridgeData map[string]interface{} = make(map[string]interface{}, 0)
		nametmp := params.UpdateRouterDbname_0(dbname)
		bridgeData[nametmp] = &s
		result.Data["router"] = append(result.Data["router"], bridgeData)
	}
}

type ResultHistorySwap struct {
	Code uint64 `json:"code"`
	Msg string `json:"msg"`
	Data map[string][]interface{} `json:"data"`
}

type ResultSwap struct {
	Code uint64 `json:"code"`
	Msg string `json:"msg"`
	Data map[string]interface{} `json:"data"`
}

func getDbname4Config(address string) *string {
	return params.GetDbName4Config(address)
}

// GetSwapTxUn get swap tx unconfirmed
func (s *RPCAPI) GetSwap(r *http.Request, args *RouterSwapKeyArgs, result *ResultSwap) error {
	fmt.Printf("[rpcapi]GetSwap, args: %v\n", args)
	if args.ChainID == "" || args.TxID == "" {
		fmt.Printf("args is nil.\n")
		return errors.New("args is nil")
	}
	if !params.IsSupportChainID(args.ChainID) {
		fmt.Printf("chainid: %v is not support.\n", args.ChainID)
		return errors.New("chainid not support")
	}

	result.Code = 0
	result.Msg = ""
	result.Data = make(map[string]interface{}, 0)

	var (
		dbname *string
		isbridge bool
		swaptx interface{}
		data []interface{}
	)


	if args.Bridge == "all" {
		dbname, swaptx, isbridge, data = getSwapAlldb(args.ChainID, args.TxID)
	} else {
		// 1 get swap, return dbname
		if params.IsNevmChain(args.ChainID) {
			dbname, swaptx, isbridge, data = getNevmChainSwap(r, args)
		} else {
			dbname, swaptx, isbridge, data = getChainSwap(r, args)
		}
	}
	if dbname == nil || len(data) == 0 {
		// error return
		fmt.Printf("GetSwap, txHash: %v not found\n", args.TxID)
		result.Code = 1
		result.Msg = "tx not found"
		return errors.New("tx not found")
	} else {
		returnName := "router"
		if isbridge {
			returnName = "bridge"
		}
		result.Data[returnName] = data
	}

	wg := new(sync.WaitGroup)
	// 2 get 2 get log
	wg.Add(1)
	go func() {
		defer wg.Done()
		reslog := swapapi.GetFileLogs(*dbname, args.TxID, isbridge)
		result.Data["log"] = reslog
	}()

	// 3 get swap tx
	wg.Add(1)
	go func() {
		defer wg.Done()
		tx := getSwaptx(swaptx, isbridge)
		result.Data["swaptx"] = tx
	}()

	wg.Wait()
	return nil
}

type swaptxConfig struct {
	Status string `json:"code"`
	Msg string `json:"msg"`
	Data *swaptxErrConfig `json:"data"`
}

type swaptxErrConfig struct {
	ChainID string `json:"toChainID"`
	TxID string `json:"swaptx"`
	Timestamp uint64 `json:"timestamp"`
	Transaction *types.Transaction `json:"transaction"`
}

func getSwaptx(swaptx interface{}, isbridge bool) *swaptxConfig {
	fmt.Printf("swaptx: %v\n", swaptx)
	if swaptx == nil {
		return nil
	}
	var (
		stx swaptxConfig
		stxret swaptxErrConfig
	)
	stx.Data = &stxret

	chainid, txid := getSwaptxInfo(swaptx, isbridge)
	stxret.ChainID = chainid
	if len(txid) == 0 {
		stx.Status = receiptStatusFailed
		stx.Msg = fmt.Sprintf("swaptx is nil")
		return &stx
	}
	stxret.TxID = txid
	ethclient := params.GetEthClient(chainid)
	if ethclient == nil || !checktxcommon.IsHexHash(txid) {
		stx.Status = receiptStatusFailed
		stx.Msg = fmt.Sprintf("chainid '%v' client is nil or txhash '%v' format err", chainid, txid)
		return &stx
	}
	receipt, err := getTransactionReceipt(ethclient, common.HexToHash(txid))
	if err != nil {
		stx.Status = receiptStatusFailed
		stx.Msg = fmt.Sprintf("txhash '%v' of chainid '%v' not exist", txid, chainid)
		return &stx
	}
	stx.Status = receiptStatusSuccess
	if receipt.Status == types.ReceiptStatusFailed {
		stx.Status = receiptStatusFailed
		stx.Msg = fmt.Sprintf("txhash '%v' receipt status failed", stxret.TxID)
		return &stx
	}
	header, _ := getHeaderByHash(ethclient, receipt.BlockHash)
	if header != nil {
		stxret.Timestamp = header.Time
	}
	tx, _ := getTransaction(ethclient, common.HexToHash(txid))
	stxret.Transaction = tx
	return &stx
}

func getSwaptxInfo(swaptx interface{}, isbridge bool) (string, string) {
	if isbridge {
		tx := swaptx.(*swapapi.BridgeSwapInfo)
		return tx.ToChainID, tx.SwapTx
	} else {
		tx := swaptx.(*swapapi.SwapInfo)
		return tx.ToChainID, tx.SwapTx
	}
}

func getSwapAlldb(chainid, txid string) (dbname *string, swaptx interface{}, isbridge bool, data []interface{}) {
	dbnames := params.GetBridgeDbName()
	for _, name := range dbnames {
		datab, err := swapapi.GetBridgeSwap(name, chainid, txid)
		if err == nil {
			var bridgeData map[string]interface{} = make(map[string]interface{}, 0)
			bridgeData[name] = datab
			isbridge = true
			dbname = &name
			data = append(data, bridgeData)
			swaptx = datab
			break
		}
	}
	if !isbridge {
		dbnames := params.GetRouterAllDbName()
		for _, name := range dbnames {
			fmt.Printf("dbname: %v\n", name)
			datab, err := swapapi.GetRouterSwap(name, chainid, txid, "0")
			if err == nil {
				var bridgeData map[string]interface{} = make(map[string]interface{}, 0)
				bridgeData[name] = datab
				isbridge = false
				dbname = &name
				data = append(data, bridgeData)
				swaptx = datab
				break
			}
		}
	}
	return dbname, swaptx, isbridge, data
}

func getChainSwap(r *http.Request, args *RouterSwapKeyArgs) (dbname *string, swaptx interface{}, isbridge bool, data []interface{}) {
	ethclient := params.GetEthClient(args.ChainID)
	tx, err := getTransaction(ethclient, common.HexToHash(args.TxID))
	if err == nil {
		to := tx.To().String()
		// bridge deposit
		dbname = getDbname4Config(to)
		isbridge = true
		if dbname == nil {
			dbname, isbridge = getAddress4Contract(args.ChainID, args.TxID)
		}
	}
	var res interface{}
	if dbname != nil {
		var err error
		// bridge
		if isbridge {
			fmt.Printf("find bridge dbname: %v, txid: %v\n", *dbname, args.TxID)
			resb, errb := swapapi.GetBridgeSwap(*dbname, args.ChainID, args.TxID)
			addBridgeChainID(*dbname, resb)
			res = resb
			err = errb
		} else {
			fmt.Printf("find router dbname: %v, txid: %v\n", *dbname, args.TxID)
			res, err = swapapi.GetRouterSwap(*dbname, args.ChainID, args.TxID, "0")
		}
		if err == nil && res != nil {
			var bridgeData map[string]interface{} = make(map[string]interface{}, 0)
			nametmp := params.UpdateRouterDbname_0(*dbname)
			bridgeData[nametmp] = res
			swaptx = res
			data = append(data, bridgeData)
		}
	}
	return dbname, swaptx, isbridge, data
}

func addBridgeChainID(dbname string, res *swapapi.BridgeSwapInfo) {
	if res == nil {
		return
	}
	chainid := strings.Split(dbname, "2")
	if len(chainid) != 2 {
		return
	}
	fromChainid := params.GetChainID(chainid[0])
	toChainid := params.GetChainID(chainid[1])
	if res.SwapType == uint32(tokens.SwapinType) {
		res.FromChainID = fromChainid
		res.ToChainID = toChainid
	} else if res.SwapType == uint32(tokens.SwapoutType) {
		res.FromChainID = toChainid
		res.ToChainID = fromChainid
	}
}

func getNevmChainSwap(r *http.Request, args *RouterSwapKeyArgs) (dbnameFound *string, swaptx interface{}, isbridge bool, data []interface{}) {
	chainid := args.ChainID
	txid := args.TxID
	dbnames := params.GetBridgeNevmDbName(chainid)
	for _, dbname := range dbnames {
		fmt.Printf("getNevmChainSwap, bridge dbname: %v\n", dbname)
		res, err := swapapi.GetBridgeSwap(dbname, chainid, txid)
		if err == nil && res != nil {
			var bridgeData map[string]interface{} = make(map[string]interface{}, 0)
			addBridgeChainID(dbname, res)
			bridgeData[dbname] = res
			swaptx = res
			data = append(data, bridgeData)
			dbnameFound = &dbname
			isbridge = true
			break
		}
	}
	if dbnameFound == nil {
		chainid := params.GetRouterStubChainID(chainid)
		dbnamer := params.GetRouterDbName()
		for _, dbname := range dbnamer {
			fmt.Printf("getNevmChainSwap router dbname: %v\n", *dbname)
			res, err := swapapi.GetRouterSwap(*dbname, chainid, txid, args.LogIndex)
			if err == nil && res != nil {
				var bridgeData map[string]interface{} = make(map[string]interface{}, 0)
				nametmp := params.UpdateRouterDbname_0(*dbname)
				bridgeData[nametmp] = res
				swaptx = res
				data = append(data, bridgeData)
				dbnameFound = dbname
				isbridge = false
				break
			}
		}
	}
	return dbnameFound, swaptx, isbridge, data
}

func getAddress4Contract(chainid, txid string) (*string, bool) {
	fmt.Printf("getAddress4Contract, txHash: %v\n", txid)
	var dbname *string
	isbridge := true
	ethclient := params.GetEthClient(chainid)
	to, token, topic, err := getTransactionReceiptTo(ethclient, common.HexToHash(txid))
	fmt.Printf("getTransactionReceiptTo, to: %v\n", to)
	if err != nil {
		fmt.Printf("getTransactionReceiptTo, chainid: %v, txid: %v, err: %v\n", chainid, txid, err)
		return nil, false
	}
	switch(topic) {
	case swapoutTopic:
		fmt.Printf("getTransactionReceiptTo, isBridgeSwapout, to: %v\n", to)
		minter, err := GetMinersAddress(ethclient, to)
		if err == nil {
			for _, m := range minter {
				fmt.Printf("getTransactionReceiptTo, minter: %v\n", *m)
				dn := getDbname4Config(*m)
				if dn != nil {
					dbname = dn
					break
				}
			}
		} else {
			minter, err := GetOwnerAddress(ethclient, to)
			if err == nil {
				dbname = getDbname4Config(minter)
			}
		}
	case routerTopic:
		minter, err := GetRouterAddress(params.GetEthClient("56"), chainid, to, token)
		fmt.Printf("getTransactionReceiptTo, isRouter, minter: %v, err: %v\n", minter, err)
		if err == nil {
			dbname = getDbname4Config(minter)
			fmt.Printf("to: %v, dbname: %v\n", to, *dbname)
			isbridge = false
		}
	case swapinTopic:
		fmt.Printf("getTransactionReceiptTo, isBridgeSwapin\n")
		dbname = getDbname4Config(to)
	}

	return dbname, isbridge
}

func isBridgeSwapin(topic int) bool {
	return topic == swapinTopic
}
func isBridgeSwapout(topic int) bool {
	return topic == swapoutTopic
}

// bridge
// GetSwapinHistory api
func (s *RPCAPI) GetSwapinHistory(r *http.Request, args *RPCQueryHistoryArgs, result *ResultHistorySwap) error {
	fmt.Printf("[rpcapi]GetSwapinHistory, args: %v\n", args)
	result.Code = 0
	result.Msg = ""
	result.Data = make(map[string][]interface{}, 0)
	dbname := args.Bridge
	status := args.Status
	if status == "" {
		status = "0,8,9,12,14,17" // default
	}
	if dbname == "all" {
		dbnames := params.GetBridgeDbName()
		for _, dbname := range dbnames {
			getBridgeSwapHistory(dbname, status, result, true)
		}
	} else {
		getBridgeSwapHistory(dbname, status, result, true)
	}
	return nil
}

func getBridgeSwapHistory(dbname, statuses string, result *ResultHistorySwap, isSwapin bool) error {
	fmt.Printf("getBridgeSwapHistory, dbname: %v\n", dbname)
	parts := strings.Split(statuses, ",")
	var s []interface{} = make([]interface{}, 0)
	var getH bool
	for _, status := range parts {
		if status == "10" {
			continue
		}
		var si []*swapapi.BridgeSwapInfo
		var errs error
		if isSwapin {
			si, errs = swapapi.GetSwapinHistory(dbname, "", "", 0, 20, status)
		} else {
			si, errs = swapapi.GetSwapoutHistory(dbname, "", "", 0, 20, status)
		}
		if errs == nil && len(si) != 0 {
			for _, st := range si {
				addBridgeChainID(dbname, st)
				s = append(s, st)
				getH = true
				//spew.Printf("%v\n", st)
			}
			//return nil
		}
	}
	if getH {
		var bridgeData map[string]interface{} = make(map[string]interface{}, 0)
		bridgeData[dbname] = &s
		result.Data["bridge"] = append(result.Data["bridge"], bridgeData)
	}
        return nil
}

// GetSwapoutHistory api
func (s *RPCAPI) GetSwapoutHistory(r *http.Request, args *RPCQueryHistoryArgs, result *ResultHistorySwap) error {
	fmt.Printf("[rpcapi]GetSwapoutHistory, args: %v\n", args)
	result.Code = 0
	result.Msg = ""
	result.Data = make(map[string][]interface{}, 0)
	dbname := args.Bridge
	status := args.Status
	if status == "" {
		status = "0,8,9,12,14,17" // default
	}
	if dbname == "all" {
		dbnames := params.GetBridgeDbName()
		for _, dbname := range dbnames {
			getBridgeSwapHistory(dbname, status, result, false)
		}
	} else {
		getBridgeSwapHistory(dbname, status, result, false)
	}
	return nil
}

// GetSwapReview api
func (s *RPCAPI) GetSwapReview(r *http.Request, args *RPCQueryHistoryArgs, result *ResultHistorySwap) error {
	fmt.Printf("[rpcapi]GetSwapReview, args: %v\n", args)
	return getSwapReview(r, args, result)
}

func getSwapReview(r *http.Request, args *RPCQueryHistoryArgs, result *ResultHistorySwap) error {
	daytime, err := getReviewTime(args.Timestamp)
	if err != nil {
		result.Code = 1
		result.Msg = "timestamp format err"
		return err
	}

	result.Code = 0
	result.Msg = ""
	result.Data = make(map[string][]interface{}, 0)
	dbname := args.Bridge
	if dbname == "all" {
		dbnames := params.GetRouterDbName()
		for _, dbname := range dbnames {
			getSwapWithTime(*dbname, daytime, result)
		}
	} else {
		getSwapWithTime(dbname, daytime, result)
	}
	return nil
}

// GetSwapinReview api
func (s *RPCAPI) GetSwapinReview(r *http.Request, args *RPCQueryHistoryArgs, result *ResultHistorySwap) error {
	fmt.Printf("[rpcapi]GetSwapinReview, args: %v\n", args)
	return getBridgeSwapReview(r, args, result, true)
}

// GetSwapoutReview api
func (s *RPCAPI) GetSwapoutReview(r *http.Request, args *RPCQueryHistoryArgs, result *ResultHistorySwap) error {
	fmt.Printf("[rpcapi]GetSwapoutReview, args: %v\n", args)
	return getBridgeSwapReview(r, args, result, false)
}

func getReviewTime(argtime string) (uint64, error) {
	now := time.Now().Unix()
	daytime := uint64(now) - params.DayUnixTime * params.LimitSwapTime
	if argtime != "" {
		timestamp, err := strconv.Atoi(argtime)
		if err != nil {
			return 0, errors.New("timestamp format err")
		}
		if uint64(timestamp) > daytime {
			daytime = uint64(timestamp)
		}
		fmt.Printf("getReviewTime, daytime: %v, timestamp: %v\n", daytime, timestamp)
	}
	return daytime, nil
}

func getBridgeSwapReview(r *http.Request, args *RPCQueryHistoryArgs, result *ResultHistorySwap, isSwapin bool) error {
	daytime, err := getReviewTime(args.Timestamp)
	if err != nil {
		result.Code = 1
		result.Msg = "timestamp format err"
		return err
	}

	result.Code = 0
	result.Msg = ""
	result.Data = make(map[string][]interface{}, 0)
	dbname := args.Bridge
	if dbname == "all" {
		dbnames := params.GetBridgeDbName()
		for _, dbname := range dbnames {
			if isSwapin {
				getSwapinWithTime(dbname, daytime, result)
			} else {
				getSwapoutWithTime(dbname, daytime, result)
			}
		}
	} else {
		if isSwapin {
			getSwapinWithTime(dbname, daytime, result)
		} else {
			getSwapoutWithTime(dbname, daytime, result)
		}
	}
	return nil
}

// getSwapinWithTime get swap from time limit
func getSwapinWithTime(dbname string, daytime uint64, result *ResultHistorySwap) {
	fmt.Printf("getSwapinWithTime, dbname: %v, daytime: %v\n", dbname, daytime)
	res, err := swapapi.GetSwapinWithTime(dbname, daytime)
	if err == nil && len(res) != 0 {
		dbname = params.UpdateRouterDbname_0(dbname)
		var bridgeData map[string]interface{} = make(map[string]interface{}, 0)
		for _, st := range res {
			addBridgeChainID(dbname, st)
		}
		bridgeData[dbname] = &res
		result.Data["bridge"] = append(result.Data["bridge"], bridgeData)
	}
}

// getSwapoutWithTime get swap from time limit
func getSwapoutWithTime(dbname string, daytime uint64, result *ResultHistorySwap) {
	fmt.Printf("getSwapoutWithTime, dbname: %v, daytime: %v\n", dbname, daytime)
	res, err := swapapi.GetSwapoutWithTime(dbname, daytime)
	if err == nil && len(res) != 0 {
		dbname = params.UpdateRouterDbname_0(dbname)
		var bridgeData map[string]interface{} = make(map[string]interface{}, 0)
		for _, st := range res {
			addBridgeChainID(dbname, st)
		}
		bridgeData[dbname] = &res
		result.Data["bridge"] = append(result.Data["bridge"], bridgeData)
	}
}

// getSwapWithTime get swap router from time limit
func getSwapWithTime(dbname string, daytime uint64, result *ResultHistorySwap) {
	fmt.Printf("getSwapWithTime, dbname: %v, daytime: %v\n", dbname, daytime)
	res, err := swapapi.GetSwapWithTime(dbname, daytime)
	if err == nil && len(res) != 0 {
		dbname = params.UpdateRouterDbname_0(dbname)
		var bridgeData map[string]interface{} = make(map[string]interface{}, 0)
		for _, st := range res {
			updateRouterChainID(st)
		}
		bridgeData[dbname] = &res
		result.Data["router"] = append(result.Data["router"], bridgeData)
	}
}

func updateRouterChainID(st *swapapi.SwapInfo) {
	fromChainid := params.GetRouterChainIDStub(st.FromChainID)
	toChainid := params.GetRouterChainIDStub(st.ToChainID)
	st.FromChainID = fromChainid
	st.ToChainID = toChainid
}

