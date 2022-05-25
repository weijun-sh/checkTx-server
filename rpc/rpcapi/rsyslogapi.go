// Package rpcapi provides JSON RPC service.
package rpcapi

import (
	"fmt"
	"net/http"

	"github.com/weijun-sh/rsyslog/internal/swapapi"
)

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

