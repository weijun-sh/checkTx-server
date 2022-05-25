package swapapi

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"strings"
	"os"

	"github.com/weijun-sh/rsyslog/common"
)

// CheckBridgeTxhash check bridge/router txhash
func CheckBridgeTxhash(bridge, txhash string) *ResultCheckBridge {
	fmt.Printf("CheckBridgeTxhash, bridge: %v, txhash: %v\n", bridge, txhash)
	if len(bridge) == 0 || !common.IsHexHash(txhash) {
		return &ResultCheckBridge{
			Code: 2,
			Msg: "bridge or txhash format error",
		}
	}
	return checkBridgeTxhash4Rsyslog(bridge, txhash)
}

func checkBridgeTxhash4Rsyslog(bridge, txhash string) *ResultCheckBridge {
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

func getBridgeTxhash4Rsyslog(bridge, txhash string) *ResultCheckBridge {
	//readLine := 10000
	var ret ResultCheckBridge
	var logRet string
	filePath := fmt.Sprintf("/opt/rsyslog/dcrm-node1/%v-server.log", bridge)
	FileHandle, err := os.Open(filePath)
	if err != nil {
		log.Println(err)
		ret.Code = 1
		ret.Msg = fmt.Sprintf("%v", err)
		return &ret
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
			logRet = string(line)
		}
	}
	fmt.Printf("logRet: %v\n", logRet)
	logSlice := strings.Split(logRet, "log ")
	if len(logSlice) < 2 {
		log.Println("log len < 2")
		ret.Code = 3
		ret.Msg = fmt.Sprintf("log len < 2")
		return &ret
	}
	fmt.Printf("logSlice: %v\n", logSlice[1])
	var status bridgeTxhashStatus
	if err := json.Unmarshal([]byte(logSlice[1]), &status); err != nil {
		log.Println(err)
		ret.Code = 4
		ret.Msg = fmt.Sprintf("%v", err)
		return &ret
	}
	ret.Data.Log = &status
	return &ret
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

