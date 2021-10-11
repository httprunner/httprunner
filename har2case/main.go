package har2case

import (
	"path/filepath"

	"github.com/httprunner/httpboomer"
)

func NewHAR(path string) *HAR {
	return &HAR{
		path: path,
	}
}

type HAR struct {
	path       string
	filterStr  string
	excludeStr string
}

func (h *HAR) GenJSON() (jsonPath string, err error) {
	ext := filepath.Ext(h.path)
	jsonPath = h.path[0:len(h.path)-len(ext)] + ".json"

	tCase := h.makeTestCase()
	err = tCase.Dump2JSON(jsonPath)
	return
}

func (h *HAR) GenYAML() (yamlPath string, err error) {
	ext := filepath.Ext(h.path)
	yamlPath = h.path[0:len(h.path)-len(ext)] + ".yaml"

	tCase := h.makeTestCase()
	err = tCase.Dump2YAML(yamlPath)
	return
}

func (h *HAR) makeTestCase() *httpboomer.TCase {
	return &httpboomer.TCase{
		Config:    *h.prepareConfig(),
		TestSteps: h.prepareTestSteps(),
	}
}

func (h *HAR) prepareConfig() *httpboomer.TConfig {
	return &httpboomer.TConfig{
		Name:      "testcase description",
		Variables: make(map[string]interface{}),
		Verify:    false,
	}
}

func (h *HAR) prepareTestSteps() []*httpboomer.TStep {
	var steps []*httpboomer.TStep
	return steps
}
