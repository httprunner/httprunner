package main

import (
	"time"

	"github.com/getsentry/sentry-go"
	"github.com/httprunner/hrp/hrp/cmd"
)

func main() {
	// Flush buffered events before the program terminates.
	defer sentry.Flush(2 * time.Second)

	cmd.Execute()
}
