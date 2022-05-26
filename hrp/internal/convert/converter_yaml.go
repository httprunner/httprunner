package convert

import (
	"github.com/pkg/errors"

	"github.com/httprunner/httprunner/v4/hrp"
	"github.com/httprunner/httprunner/v4/hrp/internal/builtin"
)

func NewConverterYAML(converter *TCaseConverter) *ConverterYAML {
	return &ConverterYAML{
		converter: converter,
	}
}

type ConverterYAML struct {
	converter *TCaseConverter
}

func (c *ConverterYAML) Struct() *TCaseConverter {
	return c.converter
}

func (c *ConverterYAML) ToJSON() (string, error) {
	testCase, err := c.makeTestCase()
	if err != nil {
		return "", err
	}
	jsonPath := c.converter.genOutputPath(suffixJSON)
	err = builtin.Dump2JSON(testCase, jsonPath)
	if err != nil {
		return "", err
	}
	return jsonPath, nil
}

func (c *ConverterYAML) ToYAML() (string, error) {
	testCase, err := c.makeTestCase()
	if err != nil {
		return "", err
	}
	yamlPath := c.converter.genOutputPath(suffixYAML)
	err = builtin.Dump2YAML(testCase, yamlPath)
	if err != nil {
		return "", err
	}
	return yamlPath, nil
}

func (c *ConverterYAML) ToGoTest() (string, error) {
	//TODO implement me
	return "", errors.New("convert from yaml testcase to gotest scripts is not supported yet")
}

func (c *ConverterYAML) ToPyTest() (string, error) {
	return convertToPyTest(c)
}

func (c *ConverterYAML) makeTestCase() (*hrp.TCase, error) {
	tCase, err := makeTestCaseFromJSONYAML(c)
	if err != nil {
		return nil, err
	}
	err = tCase.MakeCompat()
	if err != nil {
		return nil, err
	}
	return tCase, nil
}
