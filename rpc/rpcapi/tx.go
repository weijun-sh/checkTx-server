package rpcapi

import (
	"context"
	"errors"
	"fmt"
	"math/big"
	"strings"
	"strconv"
	"time"

	"github.com/weijun-sh/checkTx-server/common/hexutil"
	"github.com/weijun-sh/checkTx-server/log"
	rpcclient "github.com/weijun-sh/checkTx-server/rpc/client"
	"github.com/weijun-sh/checkTx-server/params"
	"github.com/weijun-sh/checkTx-server/router"
	"github.com/weijun-sh/checkTx-server/tokens"
	"github.com/weijun-sh/checkTx-server/tokens/eth/abicoder"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
)

var (
	retryRPCCount    = 3
	retryRPCInterval = 1 * time.Second

	RPCClientTimeout  = 60
	wrapRPCQueryError = tokens.WrapRPCQueryError
)

var (
	errEmptyURLs = errors.New("empty URLs")
)

var (
	transferLogTopic       = "0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef"
	addressSwapoutLogTopic = "0x6b616089d04950dc06c45c6dd787d657980543f89651aec47924752c7d16c888"
	stringSwapoutLogTopic  = "0x9c92ad817e5474d30a4378deface765150479363a897b0590fbb12ae9d89396b"

	routerAnySwapOutTopic                  = "0x97116cf6cd4f6412bb47914d6db18da9e16ab2142f543b86e207c24fbd16b23a"
	routerAnySwapOutTopic2                 = "0x409e0ad946b19f77602d6cf11d59e1796ddaa4828159a0b4fb7fa2ff6b161b79"
	routerAnySwapTradeTokensForTokensTopic = "0xfea6abdf4fd32f20966dff7619354cd82cd43dc78a3bee479f04c74dbfc585b3"
	routerAnySwapTradeTokensForNativeTopic = "0x278277e0209c347189add7bd92411973b5f6b8644f7ac62ea1be984ce993f8f4"

	logNFT721SwapOutTopic       = "0x0d45b0b9f5add3e1bb841982f1fa9303628b0b619b000cb1f9f1c3903329a4c7"
	logNFT1155SwapOutTopic      = "0x5058b8684cf36ffd9f66bc623fbc617a44dd65cf2273306d03d3104af0995cb0"
	logNFT1155SwapOutBatchTopic = "0xaa428a5ab688b49b415401782c170d216b33b15711d30cf69482f570eca8db38"

	logAnycallSwapOutTopic         = "0x9ca1de98ebed0a9c38ace93d3ca529edacbbe199cf1b6f0f416ae9b724d4a81c"
	logAnycallTransferSwapOutTopic = "0xcaac11c45e5fdb5c513e20ac229a3f9f99143580b5eb08d0fecbdd5ae8c81ef5"

	logAnycallV6SwapOutTopic       = "0xa17aef042e1a5dd2b8e68f0d0d92f9a6a0b35dc25be1d12c0cb3135bfd8951c9"
)

//txHash = common.HexToHash(hash)
func getTransaction(client *ethclient.Client, txHash common.Hash) (*types.Transaction, error) {
	fmt.Printf("getTransactionTo, txHash: %v\n", txHash)
	if client == nil {
		return nil, errors.New("client is nil")
	}
	for i := 0; i< 3; i++ {
		tx, _, err := client.TransactionByHash(context.Background(), txHash)
		if err == nil {
			fmt.Printf("getTransactionTo, return address: %v\n", tx.To().String())
			return tx, nil
		}
		fmt.Printf("getTransactionTo, err: %v, i: %v\n", err, i)
		time.Sleep(1 * time.Second)
	}
	return nil, errors.New("get tx failed")
}

func getTransactionReceipt(client *ethclient.Client, txHash common.Hash) (*types.Receipt, error) {
	for i := 0; i< 3; i++ {
		receipt, err := client.TransactionReceipt(context.Background(), txHash)
		if err == nil {
			return receipt, nil
		}
	}
	return nil, errors.New("get tx receipt failed")
}

//txHash = common.HexToHash(hash)
func getTransactionReceiptTo(client *ethclient.Client, txHash common.Hash) (string, string, int, error) {
	for i := 0; i< 3; i++ {
		receipt, err := client.TransactionReceipt(context.Background(), txHash)
		if err == nil {
			if len(receipt.Logs) == 0 {
				return "", "", 0, errors.New("no receipt")
			}
			for _, log := range receipt.Logs {
				//fmt.Printf("topic: %v\n", log.Topics[0])
				logTopic := log.Topics[0].String()
				if isRouterTopic(logTopic) {
					return log.Address.String(), string(common.BytesToAddress(log.Topics[1][:]).Hex()), routerTopic, nil
				}
			}
			for _, log := range receipt.Logs {
				//fmt.Printf("topic: %v\n", log.Topics[0])
				logTopic := log.Topics[0].String()
				if isSwapoutTopic(logTopic) {
					return log.Address.String(), string(common.BytesToAddress(log.Topics[1][:]).Hex()), swapoutTopic, nil
				}
			}
			for _, log := range receipt.Logs {
				fmt.Printf("topic: %v\n", log.Topics[0])
				logTopic := log.Topics[0].String()
				if isSwapinTopic(logTopic) {
					return string(common.BytesToAddress(log.Topics[2][:]).Hex()), string(common.BytesToAddress(log.Topics[1][:]).Hex()), swapinTopic, nil
				}
			}
			return "", "", 0, errors.New("get receipt topic mismatch")
		}
		//fmt.Printf("getTransactionReceiptTo, txHash: %v, err: %v\n", txHash, err)
		time.Sleep(1 * time.Second)
	}
	return "", "", 0, errors.New("get receipt failed")
}

func isContractAddress(gateway *[]string, address string) (bool, error) {
	code, err := getContractCode(gateway, address)
	if err == nil {
		return len(code) > 1, nil // unexpect RSK getCode return 0x00
	}
	return false, err
}

func getContractCode(gateway *[]string, contract string) (code []byte, err error) {
	for i := 0; i < retryRPCCount; i++ {
		code, err = GetCode(gateway, contract)
		if err == nil && len(code) > 1 {
			return code, nil
		}
		if err != nil {
			log.Warn("get contract code failed", "contract", contract, "err", err)
		}
		time.Sleep(retryRPCInterval)
	}
	return code, err
}

func GetCode(gateway *[]string, contract string) ([]byte, error) {
	return getCode(contract, gateway)
}

func getCode(contract string, urls *[]string) ([]byte, error) {
	if len(*urls) == 0 {
		return nil, errEmptyURLs
	}
	var result hexutil.Bytes
	var err error
	for _, url := range *urls {
		err = rpcclient.RPCPostWithTimeout(RPCClientTimeout, &result, url, "eth_getCode", contract, "latest")
		if err == nil {
			return []byte(result), nil
		}
	}
	return nil, wrapRPCQueryError(err, "eth_getCode", contract)
}

func isSwapinTopic(logTopic string) bool {
	for _, topic := range []string{transferLogTopic} {
		//fmt.Printf("isBridgeSwapin, logTopic: %v, topic: %v\n", logTopic, topic)
		if strings.EqualFold(logTopic, topic) {
			return true
		}
	}
	return false
}

func isSwapoutTopic(logTopic string) bool {
	for _, topic := range []string{addressSwapoutLogTopic, stringSwapoutLogTopic} {
		//fmt.Printf("isBridgeSwapout, logTopic: %v, topic: %v\n", logTopic, topic)
		if strings.EqualFold(logTopic, topic) {
			return true
		}
	}
	return false
}

func isRouterTopic(logTopic string) bool {
	for _, topic := range []string{routerAnySwapOutTopic,routerAnySwapOutTopic2, routerAnySwapTradeTokensForTokensTopic, routerAnySwapTradeTokensForNativeTopic, logNFT721SwapOutTopic, logNFT1155SwapOutTopic, logNFT1155SwapOutBatchTopic, logAnycallSwapOutTopic, logAnycallTransferSwapOutTopic, logAnycallV6SwapOutTopic} {
		//fmt.Printf("isRouterTopic, logTopic: %v, topic: %v\n", logTopic, topic)
		if strings.EqualFold(logTopic, topic) {
			return true
		}
	}
	return false
}

// GetOwnerAddress call "owner()"
func GetOwnerAddress(client *ethclient.Client, contract string) (string, error) {
	fmt.Printf("GetOwnerAddress\n")
        data := common.FromHex("0x8da5cb5b")

        to := common.HexToAddress(contract)
        msg := ethereum.CallMsg{
                To:   &to,
                Data: data,
        }
        result, err := client.CallContract(context.Background(), msg, nil)
        if err != nil {
                return "", err
        }
	//fmt.Printf("GetOwnerAddress, result: %v\n", string(common.BytesToAddress(result).Hex()))
	return string(common.BytesToAddress(result).Hex()), nil
}

// GetMintersAddress call "getAllMinters()"
func GetMinersAddress(client *ethclient.Client, contract string) ([]*string, error) {
	//fmt.Printf("GetMinersAddress\n")
        data := common.FromHex("0xa045442c")

        to := common.HexToAddress(contract)
        msg := ethereum.CallMsg{
                To:   &to,
                Data: data,
        }
        result, err := client.CallContract(context.Background(), msg, nil)
        if err != nil {
                return nil, err
        }
	//fmt.Printf("GetMinersAddress, result: %v\n", result)
	minter, errm := router.ParseMinterConfig(result)
        if errm != nil {
                return nil, errm
        }
	fmt.Printf("GetMinersAddress, minter: %v\n", minter.Minters)
	return minter.Minters, nil
}

func GetRouterAddress(client *ethclient.Client, chainid, to, token string) (string, error) {
	for router, dbname := range params.Routers {
		var isnevm bool
		if strings.Contains(strings.ToLower(*dbname), "-nevm") {
			isnevm = true
		}
		address, err := getRouterCustomAddress(client, router, chainid, token)
		fmt.Printf("router(custom): %v, address: %v, dbname: %v, chainid: %v, token: %v\n", router, address, *dbname, chainid, token)
		if err == nil && strings.EqualFold(address, to) {
			return router, nil
		}
		address, err = getRouterAddress(client, router, chainid, isnevm)
		fmt.Printf("router: %v, address: %v, to: %v, i: %v\n", router, address, to, *dbname)
		if err == nil && strings.EqualFold(address, to) {
			return router, nil
		}
	}
	return "",nil
}

// getRouterCustomAddress call "getCustomConfig(uint256,string)"
func getRouterCustomAddress(client *ethclient.Client, contract string, chainid, tokenaddress string) (string, error) {
	tokenaddress = strings.ToLower(tokenaddress)
	fmt.Printf("GetRouterCustomAddress, contract: %v, chainid: %v, tokenaddress: %v\n", contract, chainid, tokenaddress)
	n, _ := strconv.ParseUint(chainid, 10, 32)
	//data := make(hexutil.Bytes, 164)
	//copy(data[:4], common.FromHex("0x61387d61"))
	////copy(data[4:36], common.LeftPadBytes(common.FromHex("0x40"), 32))
	//copy(data[4:36], common.LeftPadBytes(common.FromHex(fmt.Sprintf("0x%x", n)), 32))
	//copy(data[36:68], common.LeftPadBytes(common.FromHex("0x40"), 32))
	//copy(data[68:100], common.LeftPadBytes(common.FromHex(fmt.Sprintf("0x%x", len(tokenaddress))), 32))
	//copy(data[100:], tokenaddress)
	//fmt.Printf("getRouterCustomAddress, data: %v\n", []byte(data))
	chainID := big.NewInt(int64(n))
	funcHash := common.FromHex("0x61387d61")
	data2 := abicoder.PackDataWithFuncHash(funcHash, chainID, []byte(tokenaddress))

	//fmt.Printf("getRouterCustomAddress, data2: %v\n", data2)
        to := common.HexToAddress(contract)
        msg := ethereum.CallMsg{
                To:   &to,
                Data: data2,
        }
        result, err := client.CallContract(context.Background(), msg, nil)
	//fmt.Printf("err: %v, result: %v\n", err, result)
        if err != nil {
                return "", err
        }
	//fmt.Printf("getRouterCustomAddress, result: %v, ok\n", result)
	return abicoder.ParseStringInData(result, 0)
}

// getRouterAddress call "getChainConfig(uint256)"
func getRouterAddress(client *ethclient.Client, contract string, chainid string, isNevm bool) (string, error) {
	//fmt.Printf("GetRouterAddress\n")
	data := make(hexutil.Bytes, 36)
	copy(data[:4], common.FromHex("0x19ed16dc"))
	n, _ := strconv.ParseUint(chainid, 10, 32)
	copy(data[4:], common.LeftPadBytes(common.FromHex(fmt.Sprintf("0x%x", n)), 32))

        to := common.HexToAddress(contract)
        msg := ethereum.CallMsg{
                To:   &to,
                Data: data,
        }
        result, err := client.CallContract(context.Background(), msg, nil)
	//fmt.Printf("err: %v\n", err)
        if err != nil {
                return "", err
        }
	//fmt.Printf("getRouterAddress, result: %v\n", result)
	return getChainConfigAddress(result, isNevm)
}

func getChainConfigAddress(data []byte, isNevm bool) (string, error) {
	var config *tokens.ChainConfig
	var err error
	if isNevm {
		config, err = router.ParseChainConfigNevm(data)
	} else {
		config, err = router.ParseChainConfig(data)
	}
        if err != nil {
                return "", err
        }
	return config.RouterContract, nil
}

