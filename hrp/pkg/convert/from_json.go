package convert

import (
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"

	"github.com/httprunner/httprunner/v4/hrp"
)

func LoadJSONCase(path string) (*hrp.TestCase, error) {
	log.Info().Str("path", path).Msg("load json case file")
	caseJSON := new(hrp.TestCase)
	err := hrp.LoadFile(path, caseJSON)
	if err != nil {
		return nil, errors.Wrap(err, "load json file failed")
	}

	if caseJSON.Steps == nil {
		return nil, errors.New("invalid json case file, missing teststeps")
	}

	err = caseJSON.MakeCompat()
	if err != nil {
		return nil, err
	}
	return caseJSON, nil
}
