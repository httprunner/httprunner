package wiki

import (
	"github.com/httprunner/funplugin/myexec"
	"github.com/rs/zerolog/log"
)

func OpenWiki() error {
	log.Info().Msgf("%s https://httprunner.com", openCmd)
	return myexec.RunCommand(openCmd, "https://httprunner.com")
}
