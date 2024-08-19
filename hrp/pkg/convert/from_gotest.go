package convert

import (
	_ "embed"
	"os"

	"github.com/rs/zerolog/log"
)

func convert2GoTestScripts(paths ...string) error {
	log.Warn().Msg("convert to gotest scripts is not supported yet")
	os.Exit(1)

	// format pytest scripts with black
	return nil
}

//go:embed testcase.tmpl
var testcaseTemplate string
