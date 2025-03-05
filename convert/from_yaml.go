package convert

import (
	"reflect"

	hrp "github.com/httprunner/httprunner/v5"
	"github.com/pkg/errors"
)

func LoadYAMLCase(path string) (*hrp.TestCaseDef, error) {
	// load yaml case file
	caseJSON := new(hrp.TestCaseDef)
	err := hrp.LoadFileObject(path, caseJSON)
	if err != nil {
		return nil, errors.Wrap(err, "load yaml file failed")
	}
	if reflect.ValueOf(*caseJSON).IsZero() {
		return nil, errors.New("invalid yaml file")
	}

	err = hrp.ConvertCaseCompatibility(caseJSON)
	if err != nil {
		return nil, err
	}
	return caseJSON, nil
}
