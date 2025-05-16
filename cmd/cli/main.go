package main

import (
	"os"
	"time"

	"github.com/getsentry/sentry-go"

	"github.com/httprunner/httprunner/v5/cmd"
	"github.com/httprunner/httprunner/v5/cmd/adb"
	"github.com/httprunner/httprunner/v5/cmd/ios"
	"github.com/httprunner/httprunner/v5/code"
)

func addAllCommands() {
	// adds all child commands to the root command and sets flags appropriately.
	cmd.RootCmd.AddCommand(cmd.CmdBuild)
	cmd.RootCmd.AddCommand(cmd.CmdConvert)
	cmd.RootCmd.AddCommand(cmd.CmdPytest)
	cmd.RootCmd.AddCommand(cmd.CmdRun)
	cmd.RootCmd.AddCommand(cmd.CmdScaffold)
	cmd.RootCmd.AddCommand(cmd.CmdServer)
	cmd.RootCmd.AddCommand(cmd.CmdWiki)
	cmd.RootCmd.AddCommand(cmd.CmdMCPHost)

	cmd.RootCmd.AddCommand(ios.CmdIOSRoot)
	cmd.RootCmd.AddCommand(adb.CmdAndroidRoot)
}

func main() {
	defer func() {
		if err := recover(); err != nil {
			// report panic to sentry
			sentry.CurrentHub().Recover(err)
			sentry.Flush(time.Second * 5)

			// print panic trace
			panic(err)
		}
	}()

	addAllCommands()

	err := cmd.RootCmd.Execute()
	exitCode := code.GetErrorCode(err)
	os.Exit(exitCode)
}
