package wiki

import (
	"github.com/rs/zerolog/log"

	"github.com/httprunner/httprunner/v4/hrp/internal/myexec"
	"github.com/httprunner/httprunner/v4/hrp/internal/sdk"
)

func OpenWiki() error {
	sdk.SendGA4Event("hrp_wiki", nil)
	log.Info().Msgf("%s https://httprunner.com", openCmd)
	return myexec.RunCommand(openCmd, "https://httprunner.com")
}
