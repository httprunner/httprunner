package convert

import (
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"

	"github.com/httprunner/httprunner/v5/hrp"
)

func LoadJSONCase(path string) (*hrp.TestCaseDef, error) {
	log.Info().Str("path", path).Msg("load json case file")
	caseJSON := new(hrp.TestCaseDef)
	err := hrp.LoadFileObject(path, caseJSON)
	if err != nil {
		return nil, errors.Wrap(err, "load json file failed")
	}

	if caseJSON.Steps == nil {
		return nil, errors.New("invalid json case file, missing teststeps")
	}

	err = hrp.ConvertCaseCompatibility(caseJSON)
	if err != nil {
		return nil, err
	}
	return caseJSON, nil
}
