package worker

import (
	"time"

	"github.com/weijun-sh/rsyslog/router/bridge"
)

const interval = 10 * time.Millisecond

// StartRouterSwapWork start router swap job
func StartRouterSwapWork(isServer bool) {
	logWorker("worker", "start router swap worker")

	bridge.InitRouterBridges(isServer)
	bridge.StartReloadRouterConfigTask()

	bridge.StartAdjustGatewayOrderJob()
	time.Sleep(interval)

	if !isServer {
		StartAcceptSignJob()
		time.Sleep(interval)
		StartReportStatJob()
		return
	}

	StartSwapJob()
	time.Sleep(interval)

	StartVerifyJob()
	time.Sleep(interval)

	StartStableJob()
	time.Sleep(interval)

	StartReplaceJob()
	time.Sleep(interval)

	StartPassBigValueJob()
	time.Sleep(interval)

	StartCheckFailedSwapJob()
}
