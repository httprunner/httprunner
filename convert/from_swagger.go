package convert

import (
	"github.com/go-openapi/spec"
	hrp "github.com/httprunner/httprunner/v5"
	"github.com/pkg/errors"
)

func LoadSwaggerCase(path string) (*hrp.TestCaseDef, error) {
	// load swagger file
	caseSwagger := new(spec.Swagger)
	err := hrp.LoadFileObject(path, caseSwagger)
	if err != nil {
		return nil, errors.Wrap(err, "load swagger file failed")
	}
	if caseSwagger.Definitions == nil {
		return nil, errors.New("invalid swagger case file, missing definitions")
	}

	// TODO: convert swagger to TCase
	return nil, nil
}
