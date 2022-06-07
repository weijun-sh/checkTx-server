// Package rpcapi provides JSON RPC service.
package rpcapi

import (
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/weijun-sh/rsyslog/internal/swapapi"

	"github.com/davecgh/go-spew/spew"
)

// #1
var (
	// tables: RouterSwapResults, RouterSwaps
	routerArray_1 = []string{"Router", "Router-Nevm"} //co-Router
	// tables: Blacklist, LatestScanInfo, LatestSwapNonces, P2shAddresses, RegisteredAddress, SwapHistory, SwapStatistics, SwapinResults, Swapins, SwapoutResults, Swapouts
	bridgeArray_1 = []string{"BTC2BSC", "ETH2BSC", "FSN2BSC", "ETH2FSN", "ETH2Fantom", "FSN2Fantom", "FSN2ETH", "BTC2ETH", "LTC2FSN", "LTC2ETH", "LTC2BSC", "LTC2Fantom", "BLOCK2ETH", "ETH2HT", "FSN2HT", "BTC2HT", "BNB2HT", "Fantom2ETH", "FORETH2Fantom", "HT2BSC", "FORETH2BSC", "FSN2MATIC", "FSN2XDAI", "ETH2XDAI", "ETH2MATIC", "ETH2AVAX", "FSN2AVAX", "BLOCK2AVAX", "BSC2AVAX", "BSC2MATIC", "BSC2ETH", "BSC2Fantom", "Harmony2MATIC", "BTC2Harmony", "COLX2BSC", "Fantom2BSC", "ETH2KCS", "COLX2ETH", "HT2MATIC", "MATIC2BSC", "MATIC2AVAX", "BSC2KCS", "BSC2Harmony", "BSC2OKT", "MATIC2OKT", "BLOCK2MATIC", "BLOCK2BSC", "BSC2MOON", "ETH2MOON", "MATIC2Fantom", "ETH2ARB", "ARB2ETH", "BSC2ARB", "MATIC2MOON", "BSC2IOTEX", "BSC2SHI", "ETH2SHI", "MOON2ETH", "BSC2CELO", "AVAX2Fantom", "ETH2HARM", "HT2Fantom", "ARB2MOON", "AVAX2BSC", "MOON2BSC", "ARB2MATIC", "ETH2TLOS", "CELO2BSC", "TERRA2Fantom", "MOON2SHI", "MATIC2HT", "ETH2IOTEX", "Harmony2BSC", "ETH2MOONBEAM", "BSC2MOONBEAM", "ETH2BOBA", "SHI2BSC", "ETH2astar", "ETH2OKT", "MATIC2ETH", "MATIC2Harmony", "MATIC2MOONBEAM", "AVAX2MOONBEAM", "BSC2astar", "BSC2ROSE", "ETH2ROSE", "ETH2VELAS", "MATIC2XDAI", "IOTEX2BSC", "XRP2AVAX", "ETH2CLV", "ETH2MIKO", "XRP2AVAX", "ETH2MIKO", "ETH2CONFLUX", "KCC2CONFLUX", "ETH2OPTIMISM", "ETH2RSK", "BSC2RSK", "JEWEL2Harmony", "TERRA2ETH", "ETH2EVMOS", "ETH2DOGE", "ETH2ETC", "ETH2CMP"}
)

type GetStatusInfoResult map[string]interface{}

// GetStatusInfo api
// TODO
func (s *RouterSwapAPI) GetStatusInfo(r *http.Request, statuses *string, result *GetStatusInfoResult) error {
	fmt.Printf("GetStatusInfo, statuses: %v\n", *statuses)
	//ret := make(map[string]interface{}, 0)
	for _, dbname := range routerArray_1 {
		fmt.Printf("\nfind dbname: %v\n", dbname)
		res, err := swapapi.GetStatusInfo(dbname, *statuses)
		if err == nil && res != nil {
			*result = res
			return nil
		}
	}
	return nil
}

type ResultBridge struct {
	Code uint64 `json:"code"`
	Msg string `json:"msg"`
	Data map[string][]*statusConfig `json:"data"`
}

type statusConfig = map[string][]*swapapi.SwapInfo

func (s *RouterSwapAPI) GetSwapNotStable(r *http.Request, args *RPCNullArgs, result *ResultBridge) error {
	var status string = "0,8,9,12,14,17"
	return s.GetSwapHistory(r, &status, result)
}

func (s *RouterSwapAPI) GetSwapHistory(r *http.Request, statuses *string, result *ResultBridge) error {
	result.Code = 0
	result.Msg = ""
	result.Data = make(map[string][]*statusConfig, 0)
	for _, dbname := range routerArray_1 {
		fmt.Printf("\nfind dbname: %v\n", dbname)
		parts := strings.Split(*statuses, ",")
		for _, status := range parts {
			si, errs := swapapi.GetRouterSwapHistory(dbname, "", "", 0, 20, status)
			if errs == nil && len(si) != 0 {
				var s statusConfig
				s = make(statusConfig, 0)
				for _, st := range si {
					s[status] = append(s[status], st)
					spew.Printf("%v\n", st)
				}
				result.Data[dbname] = append(result.Data[dbname], &s)
				//return nil
			}
		}
	}
	return nil
}

type ResultSwap struct {
	Code uint64 `json:"code"`
	Msg string `json:"msg"`
	Data *swapapi.SwapInfo `json:"data"`
}

// GetSwapTxUn get swap tx unconfirmed
func (s *RouterSwapAPI) GetSwap(r *http.Request, args *RouterSwapKeyArgs, result *ResultSwap) error {
       fmt.Printf("GetStatusInfo\n")
	result.Code = 0
	result.Msg = ""
	for _, dbname := range routerArray_1 {
		fmt.Printf("find dbname: %v\n", dbname)
		res, err := swapapi.GetRouterSwap(dbname, args.ChainID, args.TxID, args.LogIndex)
		if err == nil && res != nil {
			result.Data = res
			return nil
		}
	}
	for _, dbname := range bridgeArray_1 {
		fmt.Printf("find dbname: %v\n", dbname)
		//res, err := swapapi.GetBridgeSwap(dbname, args.ChainID, args.TxID)
		//if err == nil && res != nil {
		//	*result = *res
		//	return nil
		//}
	}
	return errors.New("not found")
}

// get from logs
// CheckBridgeArgs args
type CheckBridgeArgs struct {
	Bridge string `json:"bridge"`
	TxID   string `json:"txid"`
}

// CheckBridgeTxhash api
func (s *RouterSwapAPI) CheckBridgeTxhash(r *http.Request, args *CheckBridgeArgs, result *swapapi.ResultCheckBridge) error {
	fmt.Printf("CheckBridgeTxhash, args: %v\n", args)
	res := swapapi.CheckBridgeTxhash(args.Bridge, args.TxID)
	if res != nil {
		*result = *res
	}
	return nil
}

