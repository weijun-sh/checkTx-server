// Package rpcapi provides JSON RPC service.
package rpcapi

import (
	"fmt"
	"math/big"
	"net/http"
	"time"

	"github.com/weijun-sh/checkTx-server/internal/swapapi"
	"github.com/weijun-sh/checkTx-server/params"
	"github.com/weijun-sh/checkTx-server/router"
	"github.com/weijun-sh/checkTx-server/tokens"
)

// RPCAPI rpc api handler
type RPCAPI struct{}

// RPCNullArgs null args
type RPCNullArgs struct{}

// RouterSwapKeyArgs args
type RouterSwapKeyArgs struct {
	Bridge   string `json:"bridge"`
	ChainID  string `json:"chainid"`
	TxID     string `json:"txid"`
	LogIndex string `json:"logindex"`
}

// GetVersionInfo api
func (s *RPCAPI) GetVersionInfo(r *http.Request, args *RPCNullArgs, result *string) error {
	version := params.VersionWithMeta
	*result = version
	return nil
}

// GetServerInfo api
func (s *RPCAPI) GetServerInfo(r *http.Request, args *RPCNullArgs, result *swapapi.ServerInfo) error {
	serverInfo := swapapi.GetServerInfo()
	*result = *serverInfo
	return nil
}

type getOracleInfoResult map[string]*swapapi.OracleInfo

// GetOracleInfo api
func (s *RPCAPI) GetOracleInfo(r *http.Request, args *RPCNullArgs, result *getOracleInfoResult) error {
	oracleInfo := swapapi.GetOracleInfo()
	*result = oracleInfo
	return nil
}

// OracleInfoArgs args
type OracleInfoArgs struct {
	Enode     string `json:"enode"`
	Timestamp int64  `json:"timestamp"`
}

func (args *OracleInfoArgs) toOracleInfo() *swapapi.OracleInfo {
	return &swapapi.OracleInfo{
		Heartbeat:          time.Unix(args.Timestamp, 0).Format(time.RFC3339),
		HeartbeatTimestamp: args.Timestamp,
	}
}

// ReportOracleInfo api
func (s *RPCAPI) ReportOracleInfo(r *http.Request, args *OracleInfoArgs, result *string) error {
	err := swapapi.ReportOracleInfo(args.Enode, args.toOracleInfo())
	if err != nil {
		return err
	}
	*result = "Success"
	return nil
}

// GetRouterSwap api
func (s *RPCAPI) GetRouterSwap(r *http.Request, args *RouterSwapKeyArgs, result *swapapi.SwapInfo) error {
	res, err := swapapi.GetRouterSwap("", args.ChainID, args.TxID, args.LogIndex)
	if err == nil && res != nil {
		*result = *res
	}
	return err
}

// RouterGetSwapHistoryArgs args
type RouterGetSwapHistoryArgs struct {
	ChainID string `json:"chainid"`
	Address string `json:"address"`
	Offset  int    `json:"offset"`
	Limit   int    `json:"limit"`
	Status  string `json:"status"`
}

// GetRouterSwapHistory api
func (s *RPCAPI) GetRouterSwapHistory(r *http.Request, args *RouterGetSwapHistoryArgs, result *[]*swapapi.SwapInfo) error {
	res, err := swapapi.GetRouterSwapHistory("", args.ChainID, args.Address, args.Offset, args.Limit, args.Status)
	if err == nil && res != nil {
		*result = res
	}
	return err
}

// GetAllChainIDs api
func (s *RPCAPI) GetAllChainIDs(r *http.Request, args *RPCNullArgs, result *[]*big.Int) error {
	*result = router.AllChainIDs
	return nil
}

// GetAllTokenIDs api
func (s *RPCAPI) GetAllTokenIDs(r *http.Request, args *RPCNullArgs, result *[]string) error {
	*result = router.AllTokenIDs
	return nil
}

// GetAllMultichainTokens api
// nolint:gocritic // rpc need result of pointer type
func (s *RPCAPI) GetAllMultichainTokens(r *http.Request, args *string, result *map[string]string) error {
	tokenID := *args
	var m map[string]string
	tokensMap := router.GetCachedMultichainTokens(tokenID)
	if tokensMap != nil {
		tokensMap.Range(func(k, v interface{}) bool {
			key := k.(string)
			val := v.(string)
			m[key] = val
			return true
		})
	}
	*result = m
	return nil
}

// GetChainConfig api
func (s *RPCAPI) GetChainConfig(r *http.Request, args *string, result *swapapi.ChainConfig) error {
	chainID := *args
	bridge := router.GetBridgeByChainID(chainID)
	if bridge == nil {
		return fmt.Errorf("chainID %v not exist", chainID)
	}
	chainConfig := swapapi.ConvertChainConfig(bridge.GetChainConfig())
	if chainConfig != nil {
		*result = *chainConfig
		return nil
	}
	return fmt.Errorf("chain config not found")
}

// GetTokenConfigArgs args
type GetTokenConfigArgs struct {
	ChainID string `json:"chainid"`
	Address string `json:"address"`
}

// GetTokenConfig api
func (s *RPCAPI) GetTokenConfig(r *http.Request, args *GetTokenConfigArgs, result *swapapi.TokenConfig) error {
	chainID := args.ChainID
	address := args.Address
	bridge := router.GetBridgeByChainID(chainID)
	if bridge == nil {
		return fmt.Errorf("chainID %v not exist", chainID)
	}
	tokenConfig := swapapi.ConvertTokenConfig(bridge.GetTokenConfig(address))
	if tokenConfig != nil {
		*result = *tokenConfig
		if result.RouterContract == "" {
			result.RouterContract = bridge.GetChainConfig().RouterContract
		}
		return nil
	}
	return fmt.Errorf("token config not found")
}

// GetSwapConfigArgs args
type GetSwapConfigArgs struct {
	TokenID string `json:"tokenid"`
	ChainID string `json:"chainid"`
}

// GetSwapConfig api
func (s *RPCAPI) GetSwapConfig(r *http.Request, args *GetSwapConfigArgs, result *swapapi.SwapConfig) error {
	tokenID := args.TokenID
	chainID := args.ChainID
	swapConfig := swapapi.ConvertSwapConfig(tokens.GetSwapConfig(tokenID, chainID))
	if swapConfig != nil {
		*result = *swapConfig
		return nil
	}
	return fmt.Errorf("swap config not found")
}

