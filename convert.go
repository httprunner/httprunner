package hrp

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"strings"

	"github.com/rs/zerolog/log"
	"gopkg.in/yaml.v3"
)

func (tc *TCase) Dump2JSON(path string) error {
	path, err := filepath.Abs(path)
	if err != nil {
		log.Error().Err(err).Msg("convert absolute path failed")
		return err
	}
	log.Info().Str("path", path).Msg("dump testcase to json")
	file, _ := json.MarshalIndent(tc, "", "    ")
	err = ioutil.WriteFile(path, file, 0644)
	if err != nil {
		log.Error().Err(err).Msg("dump json path failed")
		return err
	}
	return nil
}

func (tc *TCase) Dump2YAML(path string) error {
	path, err := filepath.Abs(path)
	if err != nil {
		log.Error().Err(err).Msg("convert absolute path failed")
		return err
	}
	log.Info().Str("path", path).Msg("dump testcase to yaml")

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
		log.Error().Err(err).Msg("dump yaml path failed")
		return err
	}
	return nil
}

func loadFromJSON(path string) (*TCase, error) {
	path, err := filepath.Abs(path)
	if err != nil {
		log.Error().Str("path", path).Err(err).Msg("convert absolute path failed")
		return nil, err
	}
	log.Info().Str("path", path).Msg("load json testcase")

	file, err := ioutil.ReadFile(path)
	if err != nil {
		log.Error().Err(err).Msg("load json path failed")
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
		log.Error().Str("path", path).Err(err).Msg("convert absolute path failed")
		return nil, err
	}
	log.Info().Str("path", path).Msg("load yaml testcase")

	file, err := ioutil.ReadFile(path)
	if err != nil {
		log.Error().Err(err).Msg("load yaml path failed")
		return nil, err
	}

	tc := &TCase{}
	err = yaml.Unmarshal(file, tc)
	return tc, err
}

func LoadFromCSV(path string) []map[string]string {
	path, err := filepath.Abs(path)
	if err != nil {
		log.Error().Str("path", path).Err(err).Msg("convert absolute path failed")
		panic(err)
	}
	log.Info().Str("path", path).Msg("load csv file")

	file, err := ioutil.ReadFile(path)
	if err != nil {
		log.Error().Err(err).Msg("load csv file failed")
		panic(err)
	}
	r := csv.NewReader(strings.NewReader(string(file)))
	content, err := r.ReadAll()
	if err != nil {
		log.Error().Err(err).Msg("parse csv file failed")
		panic(err)
	}
	var result []map[string]string
	for i := 1; i < len(content); i++ {
		row := make(map[string]string)
		for j := 0; j < len(content[i]); j++ {
			row[content[0][j]] = content[i][j]
		}
		result = append(result, row)
	}
	return result
}

func (tc *TCase) ToTestCase() (*TestCase, error) {
	testCase := &TestCase{
		Config: &Config{cfg: tc.Config},
	}
	for _, step := range tc.TestSteps {
		if step.Request != nil {
			testCase.TestSteps = append(testCase.TestSteps, &StepRequestWithOptionalArgs{
				step: step,
			})
		} else if step.TestCase != nil {
			testCase.TestSteps = append(testCase.TestSteps, &StepTestCaseWithOptionalArgs{
				step: step,
			})
		} else if step.Transaction != nil {
			testCase.TestSteps = append(testCase.TestSteps, &StepTransaction{
				step: step,
			})
		} else if step.Rendezvous != nil {
			testCase.TestSteps = append(testCase.TestSteps, &StepRendezvous{
				step: step,
			})
		} else {
			log.Warn().Interface("step", step).Msg("[convertTestCase] unexpected step")
		}
	}
	return testCase, nil
}

var ErrUnsupportedFileExt = fmt.Errorf("unsupported testcase file extension")

// TestCasePath implements ITestCase interface.
type TestCasePath struct {
	Path string
}

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
