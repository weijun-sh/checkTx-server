// Command checktx is main program to start swap router or its sub commands.
package main

import (
	"fmt"
	"os"
	"time"

	"github.com/weijun-sh/checkTx-server/cmd/utils"
	"github.com/weijun-sh/checkTx-server/log"
	//"github.com/weijun-sh/checkTx-server/mongodb"
	"github.com/weijun-sh/checkTx-server/params"
	rpcserver "github.com/weijun-sh/checkTx-server/rpc/server"
	//"github.com/weijun-sh/checkTx-server/tokens"
	"github.com/weijun-sh/checkTx-server/worker"
	"github.com/urfave/cli/v2"
)

var (
	clientIdentifier = "checktx"
	// Git SHA1 commit hash of the release (set via linker flags)
	gitCommit = ""
	gitDate   = ""
	// The app that holds all commands and flags.
	app = utils.NewApp(clientIdentifier, gitCommit, gitDate, "the checktx command line interface")
)

func initApp() {
	// Initialize the CLI app and start action
	app.Action = checktx
	app.HideVersion = true // we have a command to print the version
	app.Copyright = "Copyright 2017-2020 The Rsyslog Authors"
	app.Commands = []*cli.Command{
		adminCommand,
		configCommand,
		toolsCommand,
		utils.LicenseCommand,
		utils.VersionCommand,
	}
	app.Flags = []cli.Flag{
		utils.DataDirFlag,
		utils.ServerDbConfigDirFlag,
		utils.ConfigFileFlag,
		utils.LogFileFlag,
		utils.LogRotationFlag,
		utils.LogMaxAgeFlag,
		utils.VerbosityFlag,
		utils.JSONFormatFlag,
		utils.ColorFormatFlag,
	}
}

func main() {
	initApp()
	if err := app.Run(os.Args); err != nil {
		log.Println(err)
		os.Exit(1)
	}
}

func checktx(ctx *cli.Context) error {
	utils.SetLogger(ctx)
	if ctx.NArg() > 0 {
		return fmt.Errorf("invalid command: %q", ctx.Args().Get(0))
	}

	//params.SetDataDir(utils.GetDataDir(ctx), isServer)
	configFile := utils.GetConfigFilePath(ctx)
	params.LoadRouterConfig(configFile, true)

	params.SetServerDbConfigDir(utils.GetServerDbConfigDir(ctx))
	worker.StartSwapWork()

	time.Sleep(100 * time.Millisecond)
	rpcserver.StartAPIServer()

	utils.TopWaitGroup.Wait()
	return nil
}

