package convert

import (
	"reflect"

	"github.com/go-openapi/spec"
	"github.com/pkg/errors"

	"github.com/httprunner/httprunner/v4/hrp"
	"github.com/httprunner/httprunner/v4/hrp/internal/builtin"
)

func LoadSwaggerCase(path string) (*hrp.TCase, error) {
	// load swagger file
	caseSwagger := new(spec.Swagger)
	err := builtin.LoadFile(path, caseSwagger)
	if err != nil {
		return nil, errors.Wrap(err, "load swagger file failed")
	}
	if reflect.ValueOf(*caseSwagger).IsZero() {
		return nil, errors.New("invalid swagger file")
	}

	// convert swagger to TCase
	return nil, nil
}
