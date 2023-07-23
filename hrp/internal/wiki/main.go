package wiki

import (
	"github.com/rs/zerolog/log"

	"github.com/httprunner/httprunner/v4/hrp/internal/myexec"
)

func OpenWiki() error {
	log.Info().Msgf("%s https://httprunner.com", openCmd)
	return myexec.RunCommand(openCmd, "https://httprunner.com")
}
