package swapapi

import (
	"github.com/weijun-sh/checkTx-server/mongodb"
	"github.com/weijun-sh/checkTx-server/tokens"
)

// ConvertMgoSwapToSwapInfo convert
func ConvertMgoSwapToBridgeSwapInfo(ms *mongodb.MgoBridgeSwap) *BridgeSwapInfo {
	return &BridgeSwapInfo{
		PairID:    ms.PairID,
		TxID:      ms.TxID,
		TxTo:      ms.TxTo,
		Bind:      ms.Bind,
		Status:    ms.Status,
		StatusMsg: ms.Status.String(),
		InitTime:  ms.InitTime,
		Timestamp: ms.Timestamp,
		Memo:      ms.Memo,
	}
}

// ConvertMgoSwapsToSwapInfos convert
func ConvertMgoSwapsToBridgeSwapInfos(msSlice []*mongodb.MgoBridgeSwap) []*BridgeSwapInfo {
	result := make([]*BridgeSwapInfo, len(msSlice))
	for k, v := range msSlice {
		result[k] = ConvertMgoSwapToBridgeSwapInfo(v)
	}
	return result
}

// ConvertMgoSwapResultToSwapInfo convert
func ConvertMgoSwapResultToBridgeSwapInfo(mr *mongodb.MgoBridgeSwapResult) *BridgeSwapInfo {
	var confirmations uint64
	if mr.SwapHeight != 0 {
		var latest uint64
		switch mr.SwapType {
		case uint32(tokens.SwapinType):
			latest = tokens.DstLatestBlockHeight
		case uint32(tokens.SwapoutType):
			latest = tokens.SrcLatestBlockHeight
		}
		if latest > mr.SwapHeight {
			confirmations = latest - mr.SwapHeight
		}
	}
	return &BridgeSwapInfo{
		PairID:        mr.PairID,
		TxID:          mr.TxID,
		TxTo:          mr.TxTo,
		TxHeight:      mr.TxHeight,
		From:          mr.From,
		To:            mr.To,
		Bind:          mr.Bind,
		Value:         mr.Value,
		SwapTx:        mr.SwapTx,
		SwapHeight:    mr.SwapHeight,
		SwapValue:     mr.SwapValue,
		SwapType:      mr.SwapType,
		SwapNonce:     mr.SwapNonce,
		Status:        mr.Status,
		StatusMsg:     mr.Status.String(),
		InitTime:      mr.InitTime,
		Timestamp:     mr.Timestamp,
		Memo:          mr.Memo,
		ReplaceCount:  len(mr.OldSwapTxs),
		Confirmations: confirmations,
	}
}

// ConvertMgoSwapResultsToSwapInfos convert
func ConvertMgoSwapResultsToBridgeSwapInfos(mrSlice []*mongodb.MgoBridgeSwapResult) []*BridgeSwapInfo {
	result := make([]*BridgeSwapInfo, len(mrSlice))
	for k, v := range mrSlice {
		result[k] = ConvertMgoSwapResultToBridgeSwapInfo(v)
	}
	return result
}
