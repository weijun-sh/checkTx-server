package worker

import (
	"time"

	"github.com/weijun-sh/checkTx-server/params"
	"github.com/weijun-sh/checkTx-server/mongodb"
	//"github.com/weijun-sh/checkTx-server/router/bridge"
)

const interval = 10 * time.Millisecond

// StartSwapWork start router swap job
func StartSwapWork() {
	logWorker("worker", "start swap worker")

	params.InitServerDbConfig()
	initServerDbClient()
	return
}

func initServerDbClient() {
	configs := params.GetServerDbsConfig()
	for _, config := range configs {
		dbConfig := config.MongoDB
		client := mongodb.MongoServerInit(
			config.Identifier,
			dbConfig.DBURLs,
			dbConfig.DBName,
			dbConfig.UserName,
			dbConfig.Password,
		)
		params.SetServerDbClient(config.Identifier, client)
	}
}

