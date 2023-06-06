package main

import (
	"os"
	"time"

	"github.com/getsentry/sentry-go"
	"github.com/rs/zerolog/log"

	"github.com/httprunner/httprunner/v4/hrp/cmd"
)

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

	exitCode := cmd.Execute()
	log.Info().Int("code", exitCode).Msg("hrp exit")
	os.Exit(exitCode)
}
