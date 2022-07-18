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
	"sort"
	"strings"

	"github.com/weijun-sh/checkTx-server/common"
	"github.com/weijun-sh/checkTx-server/mongodb"
	"github.com/weijun-sh/checkTx-server/params"
)

var (
	maxLog = 0 // default 100
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
type ResultCheckBridge struct {
	Code uint64 `json:"code"`
	Msg string `json:"msg"`
	Data retData `json:"data"`
}

type retData struct {
	Log *bridgeTxhashStatus `json:"log"`
}

func getRsyslogFiles(dbname string, isbridge bool) (fileRet string, fileArray []string) {
	dir := params.GetRsyslogDir(dbname)
	if dir == "" {
		return fileRet, fileArray
	}
	suffix := params.GetRsyslogSuffix(isbridge)
	files, err := ioutil.ReadDir(dir)
	if err != nil {
		return fileRet, fileArray
	}
	dbname = params.UpdateRouterDbname_0(dbname)
	if strings.HasSuffix(dbname, "_#0") {
		slice := strings.Split(dbname, "_#0")
		dbname = slice[0]
	}
	filename := fmt.Sprintf("%v%v", dbname, suffix)
	filenameOthers := fmt.Sprintf("%v%v-", dbname, suffix)
	for _, file := range files {
		if file.IsDir() {
			continue
		}
		fileNameWithSuffix := path.Base(file.Name())
		if strings.EqualFold(fileNameWithSuffix, filename) {
			fileRet = fmt.Sprintf("%v/%v", dir, file.Name())
		}
		if strings.HasPrefix(fileNameWithSuffix, filenameOthers) {
			filenametmp := fmt.Sprintf("%v/%v", dir, file.Name())
			fileArray = append(fileArray, filenametmp)
		}
	}
	return fileRet, fileArray
}

type fileSlice []string
func (f fileSlice) Len() int {
	return len(f)
}

func (f fileSlice) Swap(i, j int) {
	f[i], f[j] = f[j], f[i]
}

func (f fileSlice) Less(i, j int) bool {
	return f[i] > f[j]
}

type syslogReturn struct {
	LogFile string `json:"logFile"`
	Status string `json:"status"` // 0: ok, 1, err
	Msg string `json:"msg"`
	Logs interface{} `json:"logs"`
}

// GetLogs check bridge/router txhash
func GetFileLogs(dbname, txhash string, isbridge bool) interface{} {
	if maxLog == 0 {
		maxLog = int(params.GetLogsMaxLines(dbname))
		if maxLog == 0 {
			maxLog = 100
		}
	}
	return getBridgeTxhash4Rsyslog(dbname, txhash, isbridge)
}

func getBridgeTxhash4Rsyslog(dbname, txhash string, isbridge bool) interface{} {
	fmt.Printf("getBridgeTxhash4Rsyslog, dbname: %v, txhash: %v, isbridge: %v\n", dbname, txhash, isbridge)
	var (
		logRet []interface{}
		statusret syslogReturn
	)
	statusret.Logs = logRet
	fmt.Printf("GetFileLogs, dbname: %v, isbridge: %v, txhash: %v\n", dbname, isbridge, txhash)
	if len(dbname) == 0 || !common.IsHexHash(txhash) {
		statusret.Status = "1"
		statusret.Msg = fmt.Sprintf("dbname '%v' is nil or txhash '%v' format error", strings.ToUpper(dbname), txhash)
		logRet = append(logRet, statusret.Status)
		return statusret
	}
	logFile, logFiles := getRsyslogFiles(dbname, isbridge)
	statusret.LogFile = logFile

	if len(logFiles) == 0 {
		statusret.Status = "1"
		statusret.Msg = fmt.Sprintf("log '%v' not exist", strings.ToUpper(dbname))
		logRet = append(logRet, statusret.Status)
		return statusret
	}
	sort.Sort(fileSlice(logFiles))

	finish := getTxhash4Logfile(logFile, txhash, &logRet)
	if finish {
		return logRet
	}
	for _, filePath := range logFiles {
		finish := getTxhash4Logfile(filePath, txhash, &logRet)
		if finish {
			break
		}
	}
	return statusret
}

func getTxhash4Logfile(filePath, txhash string, logRet *[]interface{}) bool {
	FileHandle, err := os.Open(filePath) // read only
	if err != nil {
		return false
	}
	defer FileHandle.Close()
	findTxhash := false
	lenLog := len(*logRet)
	lineReader := bufio.NewReader(FileHandle)
	for {
		line, _, err := lineReader.ReadLine()
		if err == io.EOF {
			break
		}
		find := strings.Contains(string(line), txhash)
		if find {
			retStr, err := getLogsParse(string(line))
			if err == nil {
				if lenLog >= maxLog {
					*logRet = (*logRet)[1:maxLog]
					findTxhash = true
				}
				lenLog++
				*logRet = append(*logRet, retStr)
			}
		}
	}
	fmt.Printf("getTxhash4Logfile, filePath: %v, txhash: %v, readline: %v/%v\n", filePath, txhash, lenLog, maxLog)
	return findTxhash
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

