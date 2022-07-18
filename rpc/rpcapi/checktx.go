// Package rpcapi provides JSON RPC service.
package rpcapi

import (
	//"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"sync"

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
				fmt.Printf("dbname: %v\n", dbname)
		                getStatusInfo(dbname, status, result)
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
	fmt.Printf("\nfind dbname: %v\n", dbname)
	res, err := swapapi.GetStatusInfo(dbname, status)
	if err == nil && len(res) != 0 {
		var s GetStatusInfoResult
		s = res
		dbname = params.UpdateRouterDbname_0(dbname)
		result.Data[dbname] = &s
	}
}

func getBridgeStatusInfo(dbname, status string, result *ResultStatus) {
	fmt.Printf("\nfind dbname: %v\n", dbname)
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
			getSwapHistory(dbname, status, result)
		}
	} else {
		dbname = params.SetRouterDbname_0(dbname)
		getSwapHistory(dbname, status, result)
	}
	return nil
}

func getSwapHistory(dbname, statuses string, result *ResultHistorySwap) {
	fmt.Printf("\nfind dbname: %v\n", dbname)
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

	// 1 get swap, return dbname
	if params.IsNevmChain(args.ChainID) {
		dbname, swaptx, isbridge, data = getNevmChainSwap(r, args)
	} else {
		dbname, swaptx, isbridge, data = getChainSwap(r, args)
	}
	returnName := "router"
	if isbridge {
		returnName = "bridge"
	}
	result.Data[returnName] = data
	// error return
	if dbname == nil {
		fmt.Printf("GetSwap, txHash: %v not found\n", args.TxID)
		result.Code = 1
		result.Msg = "tx not found"
		return errors.New("tx not found")
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
	fmt.Printf("[rpcapi]GetSwap finished, args: %v\n", args)
	return nil
}

type swaptxConfig struct {
	ChainID string `json:"fromChainID"`
	TxID string `json:"txid"`
	Status string `json:"status"`
	Msg string `json:"msg"`
	Timestamp uint64 `json:"timestamp"`
	Transaction *types.Transaction `json:"transaction"`
}

func getSwaptx(swaptx interface{}, isbridge bool) *swaptxConfig {
	chainid, txid := getSwaptxInfo(swaptx, isbridge)
	ethclient := params.GetEthClient(chainid)
	var stx swaptxConfig
	if ethclient == nil || !checktxcommon.IsHexHash(txid) {
		stx.Status = "0"
		stx.Msg = fmt.Sprintf("chainid '%v' client is nil or txhash '%v' format err", chainid, txid)
		return &stx
	}
	receipt, err := getTransactionReceipt(ethclient, common.HexToHash(txid))
	if err != nil {
		stx.Status = "0"
		stx.Msg = fmt.Sprintf("txhash '%v' of chainid '%v' not exist", txid, chainid)
		return &stx
	}
	stx.ChainID = chainid
	stx.TxID = txid
	stx.Status = fmt.Sprintf("%v", receipt.Status)
	header, _ := getHeaderByHash(ethclient, receipt.BlockHash)
	if header != nil {
		stx.Timestamp = header.Time
	}
	tx, _ := getTransaction(ethclient, common.HexToHash(txid))
	stx.Transaction = tx
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
			fmt.Printf("find bridge dbname: %v\n", *dbname)
			resb, errb := swapapi.GetBridgeSwap(*dbname, args.ChainID, args.TxID)
			addBridgeChainID(*dbname, resb)
			res = resb
			err = errb
		} else {
			fmt.Printf("find router dbname: %v\n", *dbname)
			res, err = swapapi.GetRouterSwap(*dbname, args.ChainID, args.TxID, "0")
		}
		if err == nil && res != nil {
			var bridgeData map[string]interface{} = make(map[string]interface{}, 0)
			nametmp := params.UpdateRouterDbname_0(*dbname)
			bridgeData[nametmp] = res
			data = append(data, bridgeData)
		}
	}
	return dbname, res, isbridge, data
}

func addBridgeChainID(dbname string, res *swapapi.BridgeSwapInfo) {
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
	dbnames := params.GetBridgeNevmDbName(args.ChainID)
	for _, dbname := range dbnames {
		fmt.Printf("find dbname: %v\n", dbname)
		res, err := swapapi.GetBridgeSwap(dbname, args.ChainID, args.TxID)
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
		dbnames = params.GetRouterDbName()
		for _, dbname := range dbnames {
			fmt.Printf("find dbname: %v\n", dbname)
			res, err := swapapi.GetRouterSwap(dbname, args.ChainID, args.TxID, args.LogIndex)
			if err == nil && res != nil {
				var bridgeData map[string]interface{} = make(map[string]interface{}, 0)
				nametmp := params.UpdateRouterDbname_0(dbname)
				bridgeData[nametmp] = res
				swaptx = res
				data = append(data, bridgeData)
				dbnameFound = &dbname
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
	fmt.Printf("\nfind dbname: %v\n", dbname)
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

