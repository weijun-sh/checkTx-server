package swapapi

import (
	"github.com/weijun-sh/checkTx-server/mongodb"
	"github.com/weijun-sh/checkTx-server/router"
	"github.com/weijun-sh/checkTx-server/tokens"
)

// ConvertMgoSwapToSwapInfo convert
func ConvertMgoSwapToSwapInfo(ms *mongodb.MgoSwap) *SwapInfo {
	return &SwapInfo{
		SwapType:    ms.SwapType,
		TxID:        ms.TxID,
		TxTo:        ms.TxTo,
		From:        ms.From,
		Bind:        ms.Bind,
		Value:       ms.Value,
		LogIndex:    ms.LogIndex,
		FromChainID: ms.FromChainID,
		ToChainID:   ms.ToChainID,
		SwapInfo:    ms.SwapInfo,
		Status:      ms.Status,
		StatusMsg:   ms.Status.String(),
		InitTime:    ms.InitTime,
		Timestamp:   ms.Timestamp,
		Memo:        ms.Memo,
	}
}

// ConvertMgoSwapsToSwapInfos convert
func ConvertMgoSwapsToSwapInfos(msSlice []*mongodb.MgoSwap) []*SwapInfo {
	result := make([]*SwapInfo, len(msSlice))
	for k, v := range msSlice {
		result[k] = ConvertMgoSwapToSwapInfo(v)
	}
	return result
}

// ConvertMgoSwapResultToSwapInfo convert
func ConvertMgoSwapResultToSwapInfo(mr *mongodb.MgoSwapResult) *SwapInfo {
	var confirmations uint64
	if mr.SwapHeight != 0 {
		resBridge := router.GetBridgeByChainID(mr.ToChainID)
		if resBridge != nil {
			latest, _ := resBridge.GetLatestBlockNumber()
			if latest > mr.SwapHeight {
				confirmations = latest - mr.SwapHeight
			}
		}
	}
	return &SwapInfo{
		SwapType:      mr.SwapType,
		TxID:          mr.TxID,
		TxTo:          mr.TxTo,
		TxHeight:      mr.TxHeight,
		From:          mr.From,
		To:            mr.To,
		Bind:          mr.Bind,
		Value:         mr.Value,
		LogIndex:      mr.LogIndex,
		FromChainID:   mr.FromChainID,
		ToChainID:     mr.ToChainID,
		SwapInfo:      mr.SwapInfo,
		SwapTx:        mr.SwapTx,
		SwapHeight:    mr.SwapHeight,
		SwapValue:     mr.SwapValue,
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
func ConvertMgoSwapResultsToSwapInfos(mrSlice []*mongodb.MgoSwapResult) []*SwapInfo {
	result := make([]*SwapInfo, len(mrSlice))
	for k, v := range mrSlice {
		result[k] = ConvertMgoSwapResultToSwapInfo(v)
	}
	return result
}

// ConvertChainConfig convert chain config
func ConvertChainConfig(c *tokens.ChainConfig) *ChainConfig {
	if c == nil {
		return nil
	}
	return &ChainConfig{
		ChainID:        c.ChainID,
		BlockChain:     c.BlockChain,
		RouterContract: c.RouterContract,
		Confirmations:  c.Confirmations,
		InitialHeight:  c.InitialHeight,
	}
}

// ConvertTokenConfig convert token config
func ConvertTokenConfig(c *tokens.TokenConfig) *TokenConfig {
	if c == nil {
		return nil
	}
	return &TokenConfig{
		TokenID:         c.TokenID,
		Decimals:        c.Decimals,
		ContractAddress: c.ContractAddress,
		ContractVersion: c.ContractVersion,
		RouterContract:  c.RouterContract,
		Underlying:      c.GetUnderlying(),
	}
}

// ConvertSwapConfig convert swap config
func ConvertSwapConfig(c *tokens.SwapConfig) *SwapConfig {
	if c == nil {
		return nil
	}
	return &SwapConfig{
		MaximumSwap:           c.MaximumSwap.String(),
		MinimumSwap:           c.MinimumSwap.String(),
		BigValueThreshold:     c.BigValueThreshold.String(),
		SwapFeeRatePerMillion: c.SwapFeeRatePerMillion,
		MaximumSwapFee:        c.MaximumSwapFee.String(),
		MinimumSwapFee:        c.MinimumSwapFee.String(),
	}
}

// bridge
// ConvertMgoBridgeSwapResultToSwapInfo convert
func ConvertMgoBridgeSwapResultToSwapInfo(mr *mongodb.MgoBridgeSwapResult) *BridgeSwapInfo {
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

// ConvertMgoBridgeSwapResultsToSwapInfos convert
func ConvertMgoBridgeSwapResultsToSwapInfos(mrSlice []*mongodb.MgoBridgeSwapResult) []*BridgeSwapInfo {
        result := make([]*BridgeSwapInfo, len(mrSlice))
        for k, v := range mrSlice {
                result[k] = ConvertMgoBridgeSwapResultToSwapInfo(v)
        }
        return result
}
