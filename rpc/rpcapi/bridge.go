package rpcapi

import (
	"fmt"
	"math/big"
	"strings"

	"github.com/weijun-sh/checkTx-server/tokens"
)

func getRouterStubChainID(chainid string) string {
	fmt.Printf("getRouterStubChainID, chainid: %v\n", chainid)
	if strings.EqualFold(chainid, "xrp") {
		stubChainID := new(big.Int).SetBytes([]byte("XRP"))
		stubChainID.Mod(stubChainID, tokens.StubChainIDBase)
		stubChainID.Add(stubChainID, tokens.StubChainIDBase)
		fmt.Printf("getRouterStubChainID, chainid: %v -> %v\n", chainid, stubChainID.String())
		return stubChainID.String()
	}
	return chainid
}
