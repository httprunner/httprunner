package convert

import (
	"os"

	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
)

func convertToPyTest(iCaseConverter ICaseConverter) (string, error) {
	// convert to temporary json testcase
	jsonPath, err := iCaseConverter.ToJSON()
	inputType := iCaseConverter.Struct().InputType
	if err != nil {
		return "", errors.Wrapf(err, "(%s -> pytest step 1) failed to convert to temporary json testcase", inputType.String())
	}
	defer func() {
		if jsonPath != "" {
			if err = os.Remove(jsonPath); err != nil {
				log.Error().Err(err).Msgf("(%s -> pytest step defer) failed to clean temporary json testcase", inputType.String())
			}
		}
	}()

	// convert from temporary json testcase to pytest
	converterJSON := NewConverterJSON(NewTCaseConverter(jsonPath))
	pyTestPath, err := converterJSON.MakePyTestScript()
	if err != nil {
		return "", errors.Wrap(err, "(json -> pytest step 2) failed to convert from temporary json testcase to pytest ")
	}

	// rename resultant pytest
	renamedPyTestPath := iCaseConverter.Struct().genOutputPath(suffixPyTest)
	err = os.Rename(pyTestPath, renamedPyTestPath)
	if err != nil {
		log.Error().Err(err).Msg("(json -> pytest step 3) failed to rename the resultant pytest file")
		return pyTestPath, nil
	}
	return renamedPyTestPath, nil
}
