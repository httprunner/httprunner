package convert

import (
	"fmt"

	"github.com/pkg/errors"

	"github.com/httprunner/httprunner/v4/hrp"
	"github.com/httprunner/httprunner/v4/hrp/internal/builtin"
	"github.com/httprunner/httprunner/v4/hrp/internal/version"
)

func NewConverterJSON(converter *TCaseConverter) *ConverterJSON {
	return &ConverterJSON{
		converter: converter,
	}
}

type ConverterJSON struct {
	converter *TCaseConverter
}

func (c *ConverterJSON) Struct() *TCaseConverter {
	return c.converter
}

func (c *ConverterJSON) ToJSON() (string, error) {
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

func (c *ConverterJSON) ToYAML() (string, error) {
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

func (c *ConverterJSON) ToGoTest() (string, error) {
	//TODO implement me
	return "", errors.New("convert from json testcase to gotest scripts is not supported yet")
}

func (c *ConverterJSON) ToPyTest() (string, error) {
	return convertToPyTest(c)
}

func (c *ConverterJSON) MakePyTestScript() (string, error) {
	httprunner := fmt.Sprintf("httprunner>=%s", version.HttpRunnerMinVersion)
	python3, err := builtin.EnsurePython3Venv(httprunner)
	if err != nil {
		return "", err
	}
	args := append([]string{"-m", "httprunner", "make"}, c.converter.InputPath)
	err = builtin.ExecCommand(python3, args...)
	if err != nil {
		return "", err
	}
	return c.converter.genOutputPath(suffixPyTest), nil
}

func (c *ConverterJSON) makeTestCase() (*hrp.TCase, error) {
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
