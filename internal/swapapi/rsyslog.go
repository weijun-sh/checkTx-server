package swapapi

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strings"
	"os"

	"github.com/weijun-sh/checkTx-server/common"
	"github.com/weijun-sh/checkTx-server/mongodb"
)

var (
	mongdbArray = []string{"ETH2BSC"}
	tableArray = []string{"SwapinResults", "Swapins", "SwapoutResults", "Swapouts"}
			//"Blacklist", "LatestScanInfo", "LatestSwapNonces"
)

// GetTxhash get bridge/router txhash
func GetTxhash(chainid, txhash string) *ResultBridge {
	fmt.Printf("GetTxhash, chainid: %v, txhash: %v\n", chainid, txhash)
	if len(chainid) == 0 || !common.IsHexHash(txhash) {
		return &ResultBridge{
			Code: 2,
			Msg: "chainid or txhash format error",
		}
	}
	return getTxhash(chainid, txhash)
}

func getTxhash(chainid, txhash string) *ResultBridge {
	msg := ""
	ret := mongodb.GetTxhash4Mgodb("ETH2BSC", "SwapinResults", txhash)
	if ret == nil {
		msg = "get nothing"
	}
	return &ResultBridge{
		Code: 2,
		Msg: msg,
		Data: ret,
	}
}

type ResultBridge struct {
	Code uint64 `json:"code"`
	Msg string `json:"msg"`
	Data interface{} `json:"data"`
}

// ===== get from log
// GetLogs check bridge/router txhash
func GetLogs(bridge, txhash string) interface{} {
	fmt.Printf("CheckBridgeTxhash, bridge: %v, txhash: %v\n", bridge, txhash)
	if len(bridge) == 0 || !common.IsHexHash(txhash) {
		return errors.New("bridge or txhash format error")
	}
	return getLogs4Rsyslog(bridge, txhash)
}

func getLogs4Rsyslog(bridge, txhash string) interface{} {
	return getBridgeTxhash4Rsyslog(bridge, txhash)
}

type ResultCheckBridge struct {
	Code uint64 `json:"code"`
	Msg string `json:"msg"`
	Data retData `json:"data"`
}

type retData struct {
	Log *bridgeTxhashStatus `json:"log"`
}

func getBridgeTxhash4Rsyslog(bridge, txhash string) []interface{} {
	//readLine := 10000
	var logRet []interface{}
	filePath := fmt.Sprintf("/opt/rsyslog/dcrm-node1/%v-server.log", bridge)
	FileHandle, err := os.Open(filePath)
	if err != nil {
		return logRet
	}
	defer FileHandle.Close()
	lineReader := bufio.NewReader(FileHandle)
	//for i := 0; i < readLine; i++ {
	for {
		line, _, err := lineReader.ReadLine()
		if err == io.EOF {
			break
		}
		find := strings.Contains(string(line), txhash)
		findStatus := strings.Contains(string(line), "status")
		if find && findStatus {
			retStr, err := getLogsParse(string(line))
			if err == nil {
				logRet = append(logRet, retStr)
			}
		}
	}
	return logRet
}

func getLogsParse(logRet string) (interface{}, error) {
	fmt.Printf("logRet: %v\n", logRet)
	if len(logRet) == 0 {
		return "", errors.New("log not found")
	}
	logSlice := strings.Split(logRet, "log ")
	if len(logSlice) < 2 {
		return "", errors.New("log wrong format")
	}
	fmt.Printf("logSlice: %v\n", logSlice[1])
	var status bridgeTxhashStatus
	if err := json.Unmarshal([]byte(logSlice[1]), &status); err != nil {
		return "", err
	}
	return status, nil
}

type bridgeTxhashStatus struct {
	Status interface{} `json:"status"`
	Txhash interface{} `json:"txid,swaptxid"`
	Bind interface{} `json:"bind"`
	IsSwapin interface{} `json:"isSwapin"`
	Level interface{} `json:"level"`
	Msg interface{} `json:"msg"`
	PairID interface{} `json:"pairID"`
	Time interface{} `json:"time"`
}

