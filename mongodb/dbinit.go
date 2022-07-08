// Package mongodb is a wrapper of mongo-go-driver that
// defines the collections and CRUD apis on them.
package mongodb

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/weijun-sh/checkTx-server/cmd/utils"
	"github.com/weijun-sh/checkTx-server/log"
	"github.com/weijun-sh/checkTx-server/params"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/davecgh/go-spew/spew"
)

var (
	clientCtx = context.Background()

	appIdentifier string
	databaseName  string

	// MgoWaitGroup wait all mongodb related task done
	MgoWaitGroup = new(sync.WaitGroup)
)

//// HasClient has client
//func HasClient() bool {
//	return client != nil
//}

// MongoServerInit int mongodb server session
func MongoServerInit(appName string, hosts []string, dbName, user, pass string) (*mongo.Client) {
	appIdentifier = appName
	databaseName = dbName

	clientOpts := &options.ClientOptions{
		AppName: &appName,
		Hosts:   hosts,
		Auth: &options.Credential{
			AuthSource: dbName,
			Username:   user,
			Password:   pass,
		},
	}

	client, err := connect(clientOpts)
	if err != nil {
		log.Fatal("[mongodb] connect database failed", "hosts", hosts, "dbName", dbName, "appName", appName, "err", err)
	}

	log.Info("[mongodb] connect database success", "hosts", hosts, "dbName", dbName, "appName", appName)

	utils.TopWaitGroup.Add(1)
	go utils.WaitAndCleanup(doCleanup)
	return client
}

func doCleanup() {
	defer utils.TopWaitGroup.Done()
	MgoWaitGroup.Wait()

	client := params.GetServerDbClient()
	for _, c := range client {
		err := c.Disconnect(clientCtx)
		if err != nil {
			log.Error("[mongodb] close connection failed", "appName", appIdentifier, "err", err)
		} else {
			log.Info("[mongodb] close connection success", "appName", appIdentifier)
		}
	}
}

func connect(opts *options.ClientOptions) (client *mongo.Client, err error) {
	ctx, cancel := context.WithTimeout(clientCtx, 10*time.Second)
	defer cancel()

	client, err = mongo.Connect(ctx, opts)
	if err != nil {
		return client, err
	}

	err = client.Ping(clientCtx, nil)
	if err != nil {
		return client, err
	}

	//initCollections()
	return client, nil
}

func GetTxhash4Mgodb(dbname, tablename, txhash string) interface{} {
	fmt.Printf("TestSwitchDB, txhash: %v\n", txhash)
	swap := &MgoSwap{}
	client, err := params.GetClientByDbName(dbname)
	if err != nil {
		return swap
	}
	database := client.Database(dbname)
	collRouterSwapinResult := database.Collection(tablename)
	errt := collRouterSwapinResult.FindOne(clientCtx, bson.M{"txid": txhash}).Decode(swap)
        if errt == nil {
		spew.Printf("%v\n", swap)
		return swap
	}
	fmt.Printf("err: %v\n", errt)
	return nil
}

