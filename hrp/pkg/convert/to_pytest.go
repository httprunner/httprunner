package convert

import (
	"github.com/pkg/errors"

	"github.com/httprunner/httprunner/v4/hrp/internal/myexec"
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
