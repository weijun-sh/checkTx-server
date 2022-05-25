package bridge

import (
	"math/big"

	"github.com/weijun-sh/rsyslog/log"
	"github.com/weijun-sh/rsyslog/tokens"
	"github.com/weijun-sh/rsyslog/tokens/eth"
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
