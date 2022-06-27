package convert

import (
	"reflect"

	"github.com/pkg/errors"

	"github.com/httprunner/httprunner/v4/hrp"
	"github.com/httprunner/httprunner/v4/hrp/internal/builtin"
)

func LoadJSONCase(path string) (*hrp.TCase, error) {
	// load json case file
	caseJSON := new(hrp.TCase)
	err := builtin.LoadFile(path, caseJSON)
	if err != nil {
		return nil, errors.Wrap(err, "load json file failed")
	}
	if reflect.ValueOf(*caseJSON).IsZero() {
		return nil, errors.New("invalid json file")
	}

	err = caseJSON.MakeCompat()
	if err != nil {
		return nil, err
	}
	return caseJSON, nil
}
