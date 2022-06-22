package client

import (
	//"context"

	"github.com/weijun-sh/checkTx-server/log"

	"github.com/ethereum/go-ethereum/ethclient"
)

func InitClient(chainid, url string) *ethclient.Client {
	ethcli, err := ethclient.Dial(url)
	if err != nil {
		//log.Fatal("ethclient.Dail failed", "gateway", url, "err", err)
		log.Warn("ethclient.Dail failed", "gateway", url, "err", err)
	}
	//log.Info("ethclient.Dail gateway success", "gateway", url)
	//chainID, errid := ethcli.ChainID(context.Background())
	//if errid != nil {
	//	log.Fatal("ethcli.ChainID failed", "err", errid)
	//}
	//if chainid != chainID.String() {
	//	log.Fatal("ethcli.ChainID mismatch", "chainid", chainid, "chainID", chainID)
	//}
	//log.Info("get chainID success", "chainID", chainID)
	return ethcli
}

