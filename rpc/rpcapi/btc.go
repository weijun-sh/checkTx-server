package rpcapi

import (
	"fmt"
	"strings"

	//"github.com/weijun-sh/checkTx-server/log"
	"github.com/weijun-sh/checkTx-server/params"
	"github.com/weijun-sh/checkTx-server/rpc/client"
	"github.com/weijun-sh/checkTx-server/tokens"
	"github.com/weijun-sh/checkTx-server/tokens/btc/electrs"
)

// getTransactionStatus impl
func getTransactionStatus(chainid, txHash string) *swaptxConfig {
	var stx swaptxConfig
	var stxret swaptxErrConfig
	stx.Data = &stxret
	stxret.ChainID = chainid
	stxret.TxID = txHash

	txStatus := &tokens.TxStatus{}
	electStatus, err := getElectTransactionStatus(chainid, txHash)
	fmt.Printf("GetTransactionStatus, electStatus: %v, err: %v\n", electStatus, err)
	if err != nil {
		stx.Status = receiptStatusFailed
		stx.Msg = err.Error()
		return &stx
	}
	if !*electStatus.Confirmed {
		stx.Status = receiptStatusFailed
		err = tokens.ErrTxNotStable
		stx.Msg = err.Error()
		return &stx
	}
	stx.Status = receiptStatusSuccess
	if electStatus.BlockHash != nil {
		txStatus.BlockHash = *electStatus.BlockHash
	}
	if electStatus.BlockTime != nil {
		txStatus.BlockTime = *electStatus.BlockTime
	}
	if electStatus.BlockHeight != nil {
		txStatus.BlockHeight = *electStatus.BlockHeight
		latest, errt := getLatestBlockNumber(chainid)
		if errt == nil {
			if latest > txStatus.BlockHeight {
				txStatus.Confirmations = latest - txStatus.BlockHeight
			}
		}
	}

	stxret.Timestamp = txStatus.BlockTime
	tx, _ := getTransactionByHash(chainid, txHash)
	stxret.Transaction = tx
	return &stx
}

// getElectTransactionStatus call /tx/{txHash}/status
func getElectTransactionStatus(chainid, txHash string) (*electrs.ElectTxStatus, error) {
	gateway := params.GetGateway(strings.ToLower(chainid))
	fmt.Printf("gateway: %v\n", gateway)
	var result electrs.ElectTxStatus
	var err error
	for _, apiAddress := range gateway {
		url := apiAddress + "/tx/" + txHash + "/status"
		err = client.RPCGet(&result, url)
		if err == nil {
			return &result, nil
		}
	}
	return nil, err
}

// getLatestBlockNumber call /blocks/tip/height
func getLatestBlockNumber(chainid string) (result uint64, err error) {
	gateway := params.GetGateway(strings.ToLower(chainid))
	for _, apiAddress := range gateway {
		url := apiAddress + "/blocks/tip/height"
		err = client.RPCGet(&result, url)
		if err == nil {
			return result, nil
		}
	}
	return 0, err
}

// getTransactionByHash call /tx/{txHash}
func getTransactionByHash(chainid, txHash string) (*electrs.ElectTx, error) {
	gateway := params.GetGateway(strings.ToLower(chainid))
	var result electrs.ElectTx
	var err error
	for _, apiAddress := range gateway {
		url := apiAddress + "/tx/" + txHash
		err = client.RPCGet(&result, url)
		if err == nil {
			return &result, nil
		}
	}
	return nil, err
}

