// Package rpcapi provides JSON RPC service.
package rpcapi

import (
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/weijun-sh/checkTx-server/internal/swapapi"
	"github.com/weijun-sh/checkTx-server/params"

	"github.com/ethereum/go-ethereum/common"
	"github.com/davecgh/go-spew/spew"
)

// #1
var (
	// tables: RouterSwapResults, RouterSwaps
	routerArray_1 = []string{"Router", "Router-Nevm"} //co-Router
	// tables: Blacklist, LatestScanInfo, LatestSwapNonces, P2shAddresses, RegisteredAddress, SwapHistory, SwapStatistics, SwapinResults, Swapins, SwapoutResults, Swapouts
	bridgeArray_1 = []string{"BTC2BSC", "ETH2BSC", "FSN2BSC", "ETH2FSN", "ETH2Fantom", "FSN2Fantom", "FSN2ETH", "BTC2ETH", "LTC2FSN", "LTC2ETH", "LTC2BSC", "LTC2Fantom", "BLOCK2ETH", "ETH2HT", "FSN2HT", "BTC2HT", "BNB2HT", "Fantom2ETH", "FORETH2Fantom", "HT2BSC", "FORETH2BSC", "FSN2MATIC", "FSN2XDAI", "ETH2XDAI", "ETH2MATIC", "ETH2AVAX", "FSN2AVAX", "BLOCK2AVAX", "BSC2AVAX", "BSC2MATIC", "BSC2ETH", "BSC2Fantom", "Harmony2MATIC", "BTC2Harmony", "COLX2BSC", "Fantom2BSC", "ETH2KCS", "COLX2ETH", "HT2MATIC", "MATIC2BSC", "MATIC2AVAX", "BSC2KCS", "BSC2Harmony", "BSC2OKT", "MATIC2OKT", "BLOCK2MATIC", "BLOCK2BSC", "BSC2MOON", "ETH2MOON", "MATIC2Fantom", "ETH2ARB", "ARB2ETH", "BSC2ARB", "MATIC2MOON", "BSC2IOTEX", "BSC2SHI", "ETH2SHI", "MOON2ETH", "BSC2CELO", "AVAX2Fantom", "ETH2HARM", "HT2Fantom", "ARB2MOON", "AVAX2BSC", "MOON2BSC", "ARB2MATIC", "ETH2TLOS", "CELO2BSC", "TERRA2Fantom", "MOON2SHI", "MATIC2HT", "ETH2IOTEX", "Harmony2BSC", "ETH2MOONBEAM", "BSC2MOONBEAM", "ETH2BOBA", "SHI2BSC", "ETH2astar", "ETH2OKT", "MATIC2ETH", "MATIC2Harmony", "MATIC2MOONBEAM", "AVAX2MOONBEAM", "BSC2astar", "BSC2ROSE", "ETH2ROSE", "ETH2VELAS", "MATIC2XDAI", "IOTEX2BSC", "XRP2AVAX", "ETH2CLV", "ETH2MIKO", "XRP2AVAX", "ETH2MIKO", "ETH2CONFLUX", "KCC2CONFLUX", "ETH2OPTIMISM", "ETH2RSK", "BSC2RSK", "JEWEL2Harmony", "TERRA2ETH", "ETH2EVMOS", "ETH2DOGE", "ETH2ETC", "ETH2CMP", "USDT2Fantom"}
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
	return GetStatusInfo(args, result, true)
}

// GetBridgeStatusInfo api
func (s *RPCAPI) GetBridgeStatusInfo(r *http.Request, args *RPCQueryHistoryArgs, result *ResultStatus) error {
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
	fmt.Printf("GetStatusInfo, status: %v\n", status)
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

type ResultHistory struct {
	Code uint64 `json:"code"`
	Msg string `json:"msg"`
	Data map[string]*statusConfig `json:"data"`
}

type statusConfig = []interface{}

func (s *RPCAPI) GetSwapNotStable(r *http.Request, args *RPCNullArgs, result *ResultHistory) error {
	var argsH RPCQueryHistoryArgs
	argsH.Bridge = "all"
	argsH.Status = "0,8,9,12,14,17" // default
	return s.GetSwapHistory(r, &argsH, result)
}

func (s *RPCAPI) GetSwapHistory(r *http.Request, args *RPCQueryHistoryArgs, result *ResultHistory) error {
	result.Code = 0
	result.Msg = ""
	result.Data = make(map[string]*statusConfig, 0)
	dbname := args.Bridge
	status := args.Status
	if status == "" {
		status = "0,8,9,12,14,17" // default
	}
	fmt.Printf("dbname: %v, status: %v\n", dbname, status)
	if dbname == "all" {
		dbnames := params.GetRouterDbName()
		for _, dbname := range dbnames {
			getSwapHistory(dbname, status, result)
		}
	} else {
		getSwapHistory(dbname, status, result)
	}
	return nil
}

func getSwapHistory(dbname, statuses string, result *ResultHistory) {
	fmt.Printf("\nfind dbname: %v\n", dbname)
	parts := strings.Split(statuses, ",")
	var s statusConfig = make(statusConfig, 0)
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
				spew.Printf("%v\n", st)
			}
			//return nil
		}
	}
	if getH {
		result.Data[dbname] = &s
	}
}

type ResultSwap struct {
	Code uint64 `json:"code"`
	Msg string `json:"msg"`
	Data map[string]interface{} `json:"data"`
}

var ResultData map[string]interface{}

func getDbname4Config(address string) *string {
	return params.GetDbName4Config(address)
}

// GetSwapTxUn get swap tx unconfirmed
func (s *RPCAPI) GetSwap(r *http.Request, args *RouterSwapKeyArgs, result *ResultSwap) error {
	fmt.Printf("GetSwap, args: %v\n", args)
	chainid := args.ChainID
	txid := args.TxID
	if chainid == "" || txid == "" {
		return errors.New("args err")
	}
	result.Code = 0
	result.Msg = ""
	result.Data = make(map[string]interface{}, 0)

	if params.IsNevmChain(args.ChainID) {
		return getNevmChainSwap(r, args, result)
	}

	var dbname *string

	isbridge := true
	to, err := getTransactionTo(params.EthClient[chainid], common.HexToHash(txid))
	if err == nil {
		// bridge deposit
		dbname = getDbname4Config(to)
		if dbname == nil {
			dbname, isbridge = getAddress4Contract(chainid, txid)
		}
	}
	if dbname != nil {
		var res interface{}
		var err error
		// bridge
		if isbridge {
			fmt.Printf("find bridge dbname: %v\n", *dbname)
			res, err = swapapi.GetBridgeSwap(*dbname, args.ChainID, args.TxID)
		} else {
			fmt.Printf("find router dbname: %v\n", *dbname)
			res, err = swapapi.GetRouterSwap(*dbname, args.ChainID, args.TxID, "0")
		}
		if err == nil && res != nil {
			var bridgeData map[string]interface{} = make(map[string]interface{}, 0)
			nametmp := updateRouterDbname_0(*dbname)
			bridgeData[nametmp] = res
			result.Data["bridge"] = bridgeData
		}
		// log
		reslog := swapapi.GetFileLogs(*dbname, args.TxID, isbridge)
		if reslog != nil {
			result.Data["log"] = reslog
		}
	} else {
		fmt.Printf("GetSwap, txHash: %v not found\n", txid)
		result.Code = 1
		result.Msg = "tx not found"
		return errors.New("tx not found")
	}
	return nil
}

func getNevmChainSwap(r *http.Request, args *RouterSwapKeyArgs, result *ResultSwap) error {
	var dbnameFound *string
	isbridge := true

	dbnames := params.GetBridgeNevmDbName(args.ChainID)
	for _, dbname := range dbnames {
		fmt.Printf("find dbname: %v\n", dbname)
		res, err := swapapi.GetBridgeSwap(dbname, args.ChainID, args.TxID)
		if err == nil && res != nil {
			result.Data[dbname] = res
			dbnameFound = &dbname
			break
		}
	}
	if dbnameFound == nil {
		dbnames = params.GetRouterDbName()
		for _, dbname := range dbnames {
			fmt.Printf("find dbname: %v\n", dbname)
			res, err := swapapi.GetRouterSwap(dbname, args.ChainID, args.TxID, args.LogIndex)
			if err == nil && res != nil {
				result.Data[dbname] = res
				dbnameFound = &dbname
				isbridge = false
				break
			}
		}
	}
	// log
	if dbnameFound != nil {
		reslog := swapapi.GetFileLogs(*dbnameFound, args.TxID, isbridge)
		if reslog != nil {
			result.Data["log"] = reslog
		}
	}
	return nil
}

func updateRouterDbname_0(dbname string) string {
	if dbname == "Router-1029_#0" {
		return "Router-2_#0"
	}
	return dbname
}

func getAddress4Contract(chainid, txid string) (*string, bool) {
	fmt.Printf("getAddress4Contract, txHash: %v\n", txid)
	var dbname *string
	isbridge := true
	to, token, topic, err := getTransactionReceiptTo(params.EthClient[chainid], common.HexToHash(txid))
	fmt.Printf("getTransactionReceiptTo, to: %v\n", to)
	if err != nil {
		fmt.Printf("getTransactionReceiptTo, chainid: %v, txid: %v, err: %v\n", chainid, txid, err)
		return nil, false
	}
	switch(topic) {
	case swapoutTopic:
		fmt.Printf("getTransactionReceiptTo, isBridgeSwapout\n")
		minter, err := GetMinersAddress(params.EthClient[chainid], to)
		fmt.Printf("getTransactionReceiptTo, isBridgeSwapout, miner: %v, err: %v, to: %v\n", minter, err, to)
		if err == nil {
			for _, m := range minter {
				fmt.Printf("getTransactionReceiptTo, m: %v\n", *m)
				dn := getDbname4Config(*m)
				if dn != nil {
					dbname = dn
					break
				}
			}
		} else {
			minter, err := GetOwnerAddress(params.EthClient[chainid], to)
			if err == nil {
				dbname = getDbname4Config(minter)
			}
		}
	case routerTopic:
		minter, err := GetRouterAddress(params.EthClient["56"], chainid, to, token)
		fmt.Printf("getTransactionReceiptTo, isRouter, minter: %v, err: %v\n", minter, err)
		if err == nil {
			dbname = getDbname4Config(minter)
			fmt.Printf("to: %v, dbname: %v\n", to, dbname)
			isbridge = false
		}
	case swapinTopic:
		fmt.Printf("getTransactionReceiptTo, isBridgeSwapin\n")
		dbname = getDbname4Config(to)
	}

	fmt.Printf("dbname: %v\n", dbname)
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
func (s *RPCAPI) GetSwapinHistory(r *http.Request, args *RPCQueryHistoryArgs, result *ResultHistory) error {
	result.Code = 0
	result.Msg = ""
	result.Data = make(map[string]*statusConfig, 0)
	dbname := args.Bridge
	status := args.Status
	if status == "" {
		status = "0,8,9,12,14,17" // default
	}
	fmt.Printf("dbname: %v, status: %v\n", dbname, status)
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

func getBridgeSwapHistory(dbname, statuses string, result *ResultHistory, isSwapin bool) error {
	fmt.Printf("\nfind dbname: %v\n", dbname)
	parts := strings.Split(statuses, ",")
	var s statusConfig = make(statusConfig, 0)
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
				s = append(s, st)
				getH = true
				spew.Printf("%v\n", st)
			}
			//return nil
		}
	}
	if getH {
		result.Data[dbname] = &s
	}
        return nil
}

// GetSwapoutHistory api
func (s *RPCAPI) GetSwapoutHistory(r *http.Request, args *RPCQueryHistoryArgs, result *ResultHistory) error {
	result.Code = 0
	result.Msg = ""
	result.Data = make(map[string]*statusConfig, 0)
	dbname := args.Bridge
	status := args.Status
	if status == "" {
		status = "0,8,9,12,14,17" // default
	}
	fmt.Printf("dbname: %v, status: %v\n", dbname, status)
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

