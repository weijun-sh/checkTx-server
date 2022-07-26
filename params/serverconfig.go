package params

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"github.com/weijun-sh/checkTx-server/common"
	"github.com/weijun-sh/checkTx-server/log"

	"github.com/BurntSushi/toml"
	"go.mongodb.org/mongo-driver/mongo"
)

var (
	serverDbConfigDirectory string

	serverDbConfig map[string]*ServerDbConfig = make(map[string]*ServerDbConfig)
	serverDbClient map[string]*mongo.Client = make(map[string]*mongo.Client)

	serverDbName map[string]string = make(map[string]string) // dbname -> Identifier
	addressName map[string]*string = make(map[string]*string) // address -> dbname
	Routers map[string]*string = make(map[string]*string) // address -> dbname, valid
	routerDbName []string // all
	bridgeDbName []string
	bridgeNevmDbName map[string][]string = make(map[string][]string)
	realDbName map[string]string = make(map[string]string)
)

var (
	nevmChainArray = []string{"BTC", "BLOCK", "BOBA", "CMP", "COLX", "DOGE", "JEWEL", "LTC", "NEBULAS", "REI", "RONIN", "RSK", "TERRA", "XRP"}
)

// ServerDbConfig server config
type ServerDbConfig struct {
	Identifier string
	Logs logsConfig
	MongoDB *MongoDBConfig
	Routers map[string]*string
	Bridges map[string]*string
}

type logsConfig struct {
	RsyslogDir string
	MaxLines uint64
}

// MongoDBConfig mongodb config
type MongoDBConfig struct {
	DBURL    string   `toml:",omitempty" json:",omitempty"`
	DBURLs   []string `toml:",omitempty" json:",omitempty"`
	DBName   string
	UserName string `json:"-"`
	Password string `json:"-"`
}

func InitServerDbConfig() {
	LoadServerDbConfig(true)
	initServerDbName()
	initRouterChainID()
}

// SetServerDbConfigDir set server db config directory
func SetServerDbConfigDir(dir string) {
	log.Printf("set server db config directory to '%v'", dir)
	fileStat, err := os.Stat(dir)
	if err != nil {
		log.Fatal("wrong server db config dir", "dir", dir, "err", err)
	}
	if !fileStat.IsDir() {
		log.Fatal("server db dir is not directory", "dir", dir)
	}
	serverDbConfigDirectory = dir
}

// GetTokenPairsDir get token pairs directory
func GetTokenPairsDir() string {
	return serverDbConfigDirectory
}

// SetTokenPairsConfig set token pairs config
func SetTokenPairsConfig(pairsConfig map[string]*ServerDbConfig, check bool) {
	if check {
		err := checkServerDbConfig(pairsConfig)
		if err != nil {
			log.Fatalf("check token pairs config error: %v", err)
		}
	}
	serverDbConfig = pairsConfig
}

// GetServerDbsConfig get server db config
func GetServerDbsConfig() map[string]*ServerDbConfig {
	return serverDbConfig
}

func SetServerDbClient(Identifier string, client *mongo.Client) {
	serverDbClient[Identifier] = client
}

func GetServerDbClient() map[string]*mongo.Client {
	return serverDbClient
}

// GetServerDbConfig get server db config
func GetServerDbConfig(Identifier string) *ServerDbConfig {
	serverDbCfg, exist := serverDbConfig[strings.ToLower(Identifier)]
	if !exist {
		log.Warn("GetTokenPairConfig: pairID not exist", "pairID", Identifier)
		return nil
	}
	return serverDbCfg
}

// IsServerDbExist is server db exist
func IsServerDbExist(Identifier string) bool {
	_, exist := serverDbConfig[strings.ToLower(Identifier)]
	return exist
}

// GetAllServerDbIDs get all pairIDs
func GetAllServerDbIDs() []string {
	serverDbIDs := make([]string, 0, len(serverDbConfig))
	for _, serverDbCfg := range serverDbConfig {
		serverDbIDs = append(serverDbIDs, strings.ToLower(serverDbCfg.Identifier))
	}
	return serverDbIDs
}

func checkServerDbConfig(serverDbsConfig map[string]*ServerDbConfig) (err error) {
	pairsMap := make(map[string]struct{})
	for _, serverdb := range serverDbsConfig {
		Identifier := strings.ToLower(serverdb.Identifier)
		pairsMap[Identifier] = struct{}{}
		// check config
		err = serverdb.CheckConfig()
		if err != nil {
			return err
		}
	}
	return nil
}

// CheckConfig check server db config
func (c *ServerDbConfig) CheckConfig() (err error) {
	if c.Identifier == "" {
		return errors.New("ServerDbConfig must config nonempty 'Identifier'")
	}
	if c.MongoDB == nil {
		return errors.New("ServerDbConfig must config 'MongoDBConfig'")
	}
	if err := c.MongoDB.CheckConfig(); err != nil {
		return err
	}
	return nil
}

// LoadServerDbConfig load server db config
func LoadServerDbConfig(check bool) {
	pairsConfig, err := LoadServerDbConfigInDir(serverDbConfigDirectory, check)
	if err != nil {
		log.Fatal("load token pair config error", "err", err)
	}
	SetTokenPairsConfig(pairsConfig, check)
}

// LoadServerDbConfigInDir load server db config
func LoadServerDbConfigInDir(dir string, check bool) (map[string]*ServerDbConfig, error) {
	fileInfoList, err := ioutil.ReadDir(dir)
	if err != nil {
		log.Error("read directory failed", "dir", dir, "err", err)
		return nil, err
	}
	serverDbsConfig := make(map[string]*ServerDbConfig)
	for _, info := range fileInfoList {
		if info.IsDir() {
			continue
		}
		fileName := info.Name()
		if !strings.HasSuffix(fileName, ".toml") {
			log.Info("ignore not *.toml file", "file", fileName)
			continue
		}
		var serverDbCfg *ServerDbConfig
		filePath := common.AbsolutePath(dir, fileName)
		serverDbCfg, err = loadServerDbConfig(filePath)
		if err != nil {
			return nil, err
		}
		// use all small case to identify
		Identifier := strings.ToLower(serverDbCfg.Identifier)
		// check duplicate Identifier
		if _, exist := serverDbsConfig[Identifier]; exist {
			return nil, fmt.Errorf("duplicate Identifier '%v'", serverDbCfg.Identifier)
		}
		serverDbCfg.Logs.RsyslogDir = setRsyslogDir(serverDbCfg.Logs.RsyslogDir)
		serverDbsConfig[strings.ToLower(Identifier)] = serverDbCfg
	}
	if check {
		err = checkServerDbConfig(serverDbsConfig)
		if err != nil {
			return nil, err
		}
	}
	return serverDbsConfig, nil
}

func loadServerDbConfig(configFile string) (config *ServerDbConfig, err error) {
	log.Println("start load token pair config file", configFile)
	if !common.FileExist(configFile) {
		return nil, fmt.Errorf("config file '%v' not exist", configFile)
	}
	config = &ServerDbConfig{}
	if _, err := toml.DecodeFile(configFile, &config); err != nil {
		return nil, fmt.Errorf("toml decode file error: %w", err)
	}
	var bs []byte
	if log.JSONFormat {
		bs, _ = json.Marshal(config)
	} else {
		bs, _ = json.MarshalIndent(config, "", "  ")
	}
	log.Tracef("load token pair finished. %v", string(bs))
	log.Info("finish load token pair config file", "file", configFile, "Identifier", config.Identifier)
	return config, nil
}

// setRsyslogDir set Log dir
func setRsyslogDir(dir string) string {
	if dir == "" {
		log.Warn("suggest config rsyslog dir")
		return ""
	}
	currDir, err := common.CurrentDir()
	if err != nil {
		log.Fatal("get current dir failed", "err", err)
	}
	absdir := common.AbsolutePath(currDir, dir)
	log.Info("set rsyslog dir success", "rsyslogdir", dir)
	return absdir
}

func initServerDbName() {
	var dbnameStore map[string]*string = make(map[string]*string)
	for _, config := range serverDbConfig {
		for address, name := range config.Routers {
			nametmp := strings.ToLower(*name)
			serverDbName[nametmp] = config.Identifier
			addressName[strings.ToLower(address)] = name
			routerDbName = append(routerDbName, nametmp)
			if dbnameStore[nametmp] == nil {
				dbnameStore[nametmp] = name
				if !strings.Contains(strings.ToLower(address), "invalid") {
					Routers[strings.ToLower(address)] = name
					realDbName[nametmp] = *name
				}
			}
			//fmt.Printf("initServerDbName, addressName[%v] = %v\n", strings.ToLower(address), *name)
		}
		for address, name := range config.Bridges {
			nametmp := strings.ToLower(*name)
			serverDbName[nametmp] = config.Identifier
			addressName[strings.ToLower(address)] = name
			if dbnameStore[nametmp] == nil {
				dbnameStore[nametmp] = name
				bridgeDbName = append(bridgeDbName, *name)
				initNevmDbName(*name)
				realDbName[nametmp] = *name
			}
		}
	}
}

func initNevmDbName(name string) {
	for _, btc := range nevmChainArray {
		if strings.Contains(name, btc) {
			bridgeNevmDbName[strings.ToLower(btc)] = append(bridgeNevmDbName[strings.ToLower(btc)], name)
		}
	}
}

func GetDbName4Config(address string) *string {
	return addressName[strings.ToLower(address)]
}

func GetRouterAllDbName() []string {
	return routerDbName
}

func GetRouterDbName() map[string]*string {
	return Routers
}

func GetBridgeDbName() []string {
	return bridgeDbName
}

func GetBridgeNevmDbName(btc string) []string {
	return bridgeNevmDbName[strings.ToLower(btc)]
}

func IsNevmChain(btc string) bool {
	return len(bridgeNevmDbName[strings.ToLower(btc)]) != 0
}

func GetClientByDbName(name string) (*mongo.Client, error) {
	Identifier := serverDbName[strings.ToLower(name)]
	if Identifier != "" {
		client := serverDbClient[Identifier]
		if client == nil {
			return nil, fmt.Errorf("client is nil")
		}
		return client, nil
	}
	return nil, fmt.Errorf("identifier is nil")
}

func GetRealDbName(name string) string {
	nametmp := strings.ToLower(name)
	if realDbName[nametmp] == "" {
		fmt.Printf("getRealDbName, dbname '%v' not exist\n", name)
		return name
	}
	//fmt.Printf("GetRealDbName, realDbName: %v\n", realDbName[nametmp])
	return realDbName[nametmp]
}

func GetRsyslogDir(dbname string) string {
	Identifier := serverDbName[strings.ToLower(dbname)]
	config := serverDbConfig[Identifier]
	if config == nil {
		return ""
	}
	return config.Logs.RsyslogDir
}

func GetLogsMaxLines(dbname string) uint64 {
	Identifier := serverDbName[strings.ToLower(dbname)]
	config := serverDbConfig[Identifier]
	if config == nil {
		return 0
	}
	return config.Logs.MaxLines
}

func UpdateRouterDbname_0(dbname string) string {
	dbname = GetRealDbName(dbname)
	if strings.EqualFold(dbname, "Router-1029_#0") {
		return "Router-2_#0"
	}
	if strings.EqualFold(dbname, "Router-0715_#0") {
		return "Router_#0"
	}
	if strings.EqualFold(dbname, "foreignETH2Fantom") {
		return "FORETH2Fantom"
	}
	if strings.EqualFold(dbname, "foreignETH2BSC") {
		return "FORETH2BSC"
	}
	if strings.EqualFold(dbname, "USDT-alone") {
		return "USDT2Fantom"
	}
	return dbname
}

func SetRouterDbname_0(dbname string) string {
	dbname = GetRealDbName(dbname)
	fmt.Printf("setDbname, dbname: %v\n", dbname)
	if strings.EqualFold(dbname, "Router-2_#0") {
		return "Router-1029_#0"
	}
	if strings.EqualFold(dbname, "Router_#0") {
		return "Router-0715_#0"
	}
	if strings.EqualFold(dbname, "FORETH2Fantom") {
		return "foreignETH2Fantom"
	}
	if strings.EqualFold(dbname, "FORETH2BSC") {
		return "foreignETH2BSC"
	}
	if strings.EqualFold(dbname, "USDT2Fantom") {
		return "USDT-alone"
	}
	return dbname
}

