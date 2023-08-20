package convert

import (
	"github.com/httprunner/funplugin/myexec"
	"github.com/pkg/errors"
)

// convert TCase to pytest case
func (c *TCaseConverter) toPyTest() (string, error) {
	jsonPath, err := c.toJSON()
	if err != nil {
		return "", errors.Wrap(err, "convert to JSON case failed")
	}

	args := append([]string{"make"}, jsonPath)
	err = myexec.ExecPython3Command("httprunner", args...)
	if err != nil {
		return "", err
	}
	return c.genOutputPath(suffixPyTest), nil
}
