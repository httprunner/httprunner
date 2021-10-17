package hrp

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"path/filepath"

	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"
)

func (tc *TestCase) ToTCase() (*TCase, error) {
	tCase := TCase{
		Config: tc.Config,
	}
	for _, step := range tc.TestSteps {
		tCase.TestSteps = append(tCase.TestSteps, step.ToStruct())
	}
	return &tCase, nil
}

func (tc *TCase) Dump2JSON(path string) error {
	path, err := filepath.Abs(path)
	if err != nil {
		log.Errorf("convert absolute path error: %v, path: %v", err, path)
		return err
	}
	log.Infof("dump testcase to json path: %s", path)
	file, _ := json.MarshalIndent(tc, "", "    ")
	err = ioutil.WriteFile(path, file, 0644)
	if err != nil {
		log.Errorf("dump json path error: %v", err)
		return err
	}
	return nil
}

func (tc *TCase) Dump2YAML(path string) error {
	path, err := filepath.Abs(path)
	if err != nil {
		log.Errorf("convert absolute path error: %v, path: %v", err, path)
		return err
	}
	log.Infof("dump testcase to yaml path: %s", path)

	// init yaml encoder
	buffer := new(bytes.Buffer)
	encoder := yaml.NewEncoder(buffer)
	encoder.SetIndent(4)

	// encode
	err = encoder.Encode(tc)
	if err != nil {
		return err
	}

	err = ioutil.WriteFile(path, buffer.Bytes(), 0644)
	if err != nil {
		log.Errorf("dump yaml path error: %v", err)
		return err
	}
	return nil
}

func loadFromJSON(path string) (*TCase, error) {
	path, err := filepath.Abs(path)
	if err != nil {
		log.Errorf("convert absolute path error: %v, path: %v", err, path)
		return nil, err
	}
	log.WithField("path", path).Info("load json testcase")

	file, err := ioutil.ReadFile(path)
	if err != nil {
		log.Errorf("dump json path error: %v", err)
		return nil, err
	}

	tc := &TCase{}
	decoder := json.NewDecoder(bytes.NewReader(file))
	decoder.UseNumber()
	err = decoder.Decode(tc)
	return tc, err
}

func loadFromYAML(path string) (*TCase, error) {
	path, err := filepath.Abs(path)
	if err != nil {
		log.Errorf("convert absolute path error: %v, path: %v", err, path)
		return nil, err
	}
	log.Infof("load testcase from yaml path: %s", path)

	file, err := ioutil.ReadFile(path)
	if err != nil {
		log.Errorf("dump yaml path error: %v", err)
		return nil, err
	}

	tc := &TCase{}
	err = yaml.Unmarshal(file, tc)
	return tc, err
}

func (tc *TCase) ToTestCase() (*TestCase, error) {
	testCase := &TestCase{
		Config: tc.Config,
	}
	for _, step := range tc.TestSteps {
		if step.Request != nil {
			testCase.TestSteps = append(testCase.TestSteps, &requestWithOptionalArgs{
				step: step,
			})
		} else if step.TestCase != nil {
			testCase.TestSteps = append(testCase.TestSteps, &testcaseWithOptionalArgs{
				step: step,
			})
		} else {
			log.Warnf("[convertTestCase] unexpected step: %+v", step)
		}
	}
	return testCase, nil
}

var ErrUnsupportedFileExt = fmt.Errorf("unsupported testcase file extension")

func (path *TestCasePath) ToTestCase() (*TestCase, error) {
	var tc *TCase
	var err error

	casePath := path.Path
	ext := filepath.Ext(casePath)
	switch ext {
	case ".json":
		tc, err = loadFromJSON(casePath)
	case ".yaml", ".yml":
		tc, err = loadFromYAML(casePath)
	default:
		err = ErrUnsupportedFileExt
	}
	if err != nil {
		return nil, err
	}
	testcase, err := tc.ToTestCase()
	if err != nil {
		return nil, err
	}
	return testcase, nil
}

func (path *TestCasePath) ToTCase() (*TCase, error) {
	testcase, err := path.ToTestCase()
	if err != nil {
		return nil, err
	}
	return testcase.ToTCase()
}
