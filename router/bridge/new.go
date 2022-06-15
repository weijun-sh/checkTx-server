package bridge

import (
	"math/big"

	"github.com/weijun-sh/checkTx-server/log"
	"github.com/weijun-sh/checkTx-server/tokens"
	"github.com/weijun-sh/checkTx-server/tokens/eth"
)

// NewCrossChainBridge new bridge
func NewCrossChainBridge(chainID *big.Int) tokens.IBridge {
	switch {
	case chainID.Sign() <= 0:
		log.Fatal("wrong chainID", "chainID", chainID)
	default:
		return eth.NewCrossChainBridge()
	}
	return nil
}
