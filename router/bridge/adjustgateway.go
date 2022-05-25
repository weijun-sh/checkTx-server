package bridge

import (
	"fmt"
	"time"

	"github.com/weijun-sh/rsyslog/cmd/utils"
	"github.com/weijun-sh/rsyslog/log"
	"github.com/weijun-sh/rsyslog/router"
	"github.com/weijun-sh/rsyslog/tokens"
	"github.com/weijun-sh/rsyslog/tools"
)

var (
	adjustCount    = 0
	adjustInterval = 60 // seconds
)

// StartAdjustGatewayOrderJob adjust gateway order job
func StartAdjustGatewayOrderJob() {
	log.Info("star adjust gateway order job")

	go doAdjustGatewayOrderJob()
}

func doAdjustGatewayOrderJob() {
	for {
		for _, chainID := range router.AllChainIDs {
			if utils.IsCleanuping() {
				return
			}
			adjustGatewayOrder(chainID.String())
		}
		for i := 0; i < adjustInterval; i++ {
			if utils.IsCleanuping() {
				return
			}
			time.Sleep(1 * time.Second)
		}
		adjustCount++
	}
}

func adjustGatewayOrder(chainID string) {
	bridge := router.GetBridgeByChainID(chainID)
	if bridge != nil {
		AdjustGatewayOrder(bridge, chainID)
	}
}

// AdjustGatewayOrder adjust gateway order once
func AdjustGatewayOrder(bridge tokens.IBridge, chainID string) {
	// use block number as weight
	var weightedAPIs tools.WeightedStringSlice
	gateway := bridge.GetGatewayConfig()
	length := len(gateway.APIAddress)
	for i := length; i > 0; i-- { // query in reverse order
		apiAddress := gateway.APIAddress[i-1]
		height, _ := bridge.GetLatestBlockNumberOf(apiAddress)
		weightedAPIs = weightedAPIs.Add(apiAddress, height)
	}
	weightedAPIs.Reverse() // reverse as iter in reverse order in the above
	weightedAPIs = weightedAPIs.Sort()
	gateway.APIAddress = weightedAPIs.GetStrings()
	if adjustCount%3 == 0 {
		log.Info(fmt.Sprintf("adjust gateways of chain %v", chainID), "result", weightedAPIs)
	}
}
