package swapapi

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path"
	"strings"

	"github.com/weijun-sh/checkTx-server/common"
	"github.com/weijun-sh/checkTx-server/mongodb"
	"github.com/weijun-sh/checkTx-server/params"
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
func GetFileLogs(dbname, txhash string, isbridge bool) []interface{} {
	//fmt.Printf("GetFileLogs, bridge: %v, txhash: %v\n", bridge, txhash)
	if len(dbname) == 0 || !common.IsHexHash(txhash) {
		return nil
	}
	return getBridgeTxhash4Rsyslog(dbname, txhash, isbridge)
}

type ResultCheckBridge struct {
	Code uint64 `json:"code"`
	Msg string `json:"msg"`
	Data retData `json:"data"`
}

type retData struct {
	Log *bridgeTxhashStatus `json:"log"`
}

func getRsyslogFiles(dbname string, isbridge bool) []string {
	var ret []string
	dir := params.GetRsyslogDir(dbname)
	if dir == "" {
		return ret
	}
	suffix := params.GetRsyslogSuffix(isbridge)
	files, err := ioutil.ReadDir(dir)
	if err != nil {
		return ret
	}
	if strings.HasSuffix(dbname, "_#0") {
		slice := strings.Split(dbname, "_#0")
		dbname = slice[0]
	}
	filename := fmt.Sprintf("%v%v", dbname, suffix)
	for _, file := range files {
		if file.IsDir() {
			continue
		}
		fileNameWithSuffix := path.Base(file.Name())
		if strings.HasPrefix(fileNameWithSuffix, filename) {
			filenametmp := fmt.Sprintf("%v/%v", dir, file.Name())
			ret = append(ret, filenametmp)
		}
	}
	return ret
}

func getBridgeTxhash4Rsyslog(dbname, txhash string, isbridge bool) []interface{} {
	var logRet []interface{}
	finish := 2 // find 2 files, from newest
	logFiles := getRsyslogFiles(dbname, isbridge)
	for _, filePath := range logFiles {
		if finish <= 0 {
			break
		}
		FileHandle, err := os.Open(filePath) // read only
		if err != nil {
			continue
		}
		defer FileHandle.Close()
		lineReader := bufio.NewReader(FileHandle)
		find := false
		for {
			line, _, err := lineReader.ReadLine()
			if err == io.EOF {
				break
			}
			find := strings.Contains(string(line), txhash)
			if find {
				retStr, err := getLogsParse(string(line))
				if err == nil {
					find = true
					logRet = append(logRet, retStr)
				}
			}
		}
		if find {
			finish -= 1
		}
	}
	return logRet
}

func getLogsParse(logRet string) (interface{}, error) {
	//fmt.Printf("logRet: %v\n", logRet)
	if len(logRet) == 0 {
		return "", errors.New("log not found")
	}
	logSlice := strings.Split(logRet, "log ")
	if len(logSlice) < 2 {
		return "", errors.New("log wrong format")
	}
	//fmt.Printf("logSlice: %v\n", logSlice[1])
	var status bridgeTxhashStatus
	if err := json.Unmarshal([]byte(logSlice[1]), &status); err != nil {
		return "", err
	}
	return status, nil
}

type bridgeTxhashStatus struct {
	Status interface{} `json:"status"`
	Txhash interface{} `json:"txHash"`
	TxID interface{} `json:"txid"`
	SwapID interface{} `json:"swapID"`
	Bind interface{} `json:"bind"`
	IsSwapin interface{} `json:"isSwapin"`
	Level interface{} `json:"level"`
	Error interface{} `json:"err"`
	Msg interface{} `json:"msg"`
	PairID interface{} `json:"pairID"`
	Time interface{} `json:"time"`
}

