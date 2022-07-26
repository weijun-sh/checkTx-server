package params

import (
	"math/big"
	"strings"
)

var (
	// StubChainIDBase stub chainID base value
        StubChainIDBase = big.NewInt(1000000000000)
)

var (
	routerStubArray []string = []string{"XRP"}
	routerStubChainID map[string]string = make(map[string]string) // xrp -> 1000005788240
	routerChainIDStub map[string]string = make(map[string]string) // 1000005788240 -> xrp
)

func initRouterChainID() {
	for _, stub := range routerStubArray {
		stubtmp := strings.ToUpper(stub)
		stubChainID := new(big.Int).SetBytes([]byte(stubtmp))
		stubChainID.Mod(stubChainID, StubChainIDBase)
		stubChainID.Add(stubChainID, StubChainIDBase)
		routerStubChainID[stubtmp] = stubChainID.String()
		routerChainIDStub[stubChainID.String()] = stubtmp
	}
}

func GetRouterStubChainID(chainid string) string {
	if routerStubChainID[strings.ToUpper(chainid)] == "" {
		return chainid
	}
	return routerStubChainID[strings.ToUpper(chainid)]
}

func GetRouterChainIDStub(chainid string) string {
	if routerChainIDStub[strings.ToUpper(chainid)] == "" {
		return chainid
	}
	return routerChainIDStub[strings.ToUpper(chainid)]
}

