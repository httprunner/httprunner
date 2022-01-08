package builtin

import (
	"fmt"

	"github.com/httprunner/hrp/internal/ga"
)

func CreateScaffold(projectName string) error {
	// report event
	ga.SendEvent(ga.EventTracking{
		Category: "Scaffold",
		Action:   "hrp startproject",
	})

	return fmt.Errorf("not implemented")
}
