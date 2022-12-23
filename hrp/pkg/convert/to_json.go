package convert

import "github.com/httprunner/httprunner/v4/hrp/internal/builtin"

// convert TCase to JSON case
func (c *TCaseConverter) toJSON() (string, error) {
	jsonPath := c.genOutputPath(suffixJSON)
	err := builtin.Dump2JSON(c.tCase, jsonPath)
	if err != nil {
		return "", err
	}
	return jsonPath, nil
}
