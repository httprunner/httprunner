package main

import (
	"github.com/httprunner/hrp/cli/hrp/cmd"
	"github.com/httprunner/hrp/internal/sentry"
)

func main() {
	// Flush buffered events before the program terminates.
	defer sentry.Flush()

	cmd.Execute()
}
