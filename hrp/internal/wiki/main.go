package wiki

import (
	"os/exec"

	"github.com/rs/zerolog/log"
)

func OpenWiki() error {
	log.Info().Msgf("%s https://httprunner.com", openCmd)
	return exec.Command(openCmd, "https://httprunner.com").Run()
}
