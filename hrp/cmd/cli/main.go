package main

import (
	"os"
	"time"

	"github.com/getsentry/sentry-go"
	"github.com/rs/zerolog/log"

	"github.com/httprunner/httprunner/hrp/cmd"
)

func main() {
	defer func() {
		if err := recover(); err != nil {
			// report panic to sentry
			sentry.CurrentHub().Recover(err)
			sentry.Flush(time.Second * 5)
			log.Error().Interface("err", err).Msg("recover panic")
			os.Exit(1)
		}
	}()
	cmd.Execute()
}
