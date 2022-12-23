package convert

import "github.com/httprunner/httprunner/v4/hrp/internal/builtin"

// convert TCase to YAML case
func (c *TCaseConverter) toYAML() (string, error) {
	yamlPath := c.genOutputPath(suffixYAML)
	err := builtin.Dump2YAML(c.tCase, yamlPath)
	if err != nil {
		return "", err
	}
	return yamlPath, nil
}
