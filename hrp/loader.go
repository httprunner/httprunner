package hrp

import (
	"bytes"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	"gopkg.in/yaml.v2"

	"github.com/httprunner/httprunner/v4/hrp/internal/code"
	"github.com/httprunner/httprunner/v4/hrp/internal/json"
)

// LoadTestCases load testcases from TestCasePath or TestCase
func LoadTestCases(tests ...ITestCase) ([]*TestCase, error) {
	testCases := make([]*TestCase, 0)

	for _, iTestCase := range tests {
		if testcase, ok := iTestCase.(*TestCase); ok {
			testCases = append(testCases, testcase)
			continue
		}

		// iTestCase should be a TestCasePath, file path or folder path
		tcPath, ok := iTestCase.(*TestCasePath)
		if !ok {
			return nil, errors.New("invalid iTestCase type")
		}

		casePath := string(*tcPath)
		err := fs.WalkDir(os.DirFS(casePath), ".", func(path string, dir fs.DirEntry, e error) error {
			if dir == nil {
				// casePath is a file other than a dir
				path = casePath
			} else if dir.IsDir() && path != "." && strings.HasPrefix(path, ".") {
				// skip hidden folders
				return fs.SkipDir
			} else {
				// casePath is a dir
				path = filepath.Join(casePath, path)
			}

			// ignore non-testcase files
			ext := filepath.Ext(path)
			if ext != ".yml" && ext != ".yaml" && ext != ".json" {
				return nil
			}

			// filtered testcases
			testCasePath := TestCasePath(path)
			tc, err := testCasePath.GetTestCase()
			if err != nil {
				log.Warn().Err(err).Str("path", path).Msg("load testcase failed")
				return nil
			}
			testCases = append(testCases, tc)
			return nil
		})
		if err != nil {
			return nil, errors.Wrap(err, "read dir failed")
		}
	}

	log.Info().Int("count", len(testCases)).Msg("load testcases successfully")
	return testCases, nil
}

// LoadFileObject loads file content with file extension and assigns to structObj
func LoadFileObject(path string, structObj interface{}) (err error) {
	log.Info().Str("path", path).Msg("load file")
	file, err := readFile(path)
	if err != nil {
		return errors.Wrap(err, "read file failed")
	}
	// remove BOM at the beginning of file
	file = bytes.TrimLeft(file, "\xef\xbb\xbf")
	ext := filepath.Ext(path)
	switch ext {
	case ".json", ".har":
		decoder := json.NewDecoder(bytes.NewReader(file))
		decoder.UseNumber()
		err = decoder.Decode(structObj)
		if err != nil {
			err = errors.Wrap(code.LoadJSONError, err.Error())
		}
	case ".yaml", ".yml":
		err = yaml.Unmarshal(file, structObj)
		if err != nil {
			err = errors.Wrap(code.LoadYAMLError, err.Error())
		}
	case ".env":
		err = parseEnvContent(file, structObj)
		if err != nil {
			err = errors.Wrap(code.LoadEnvError, err.Error())
		}
	default:
		err = code.UnsupportedFileExtension
	}
	return err
}

func parseEnvContent(file []byte, obj interface{}) error {
	envMap := obj.(map[string]string)
	lines := strings.Split(string(file), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			// empty line or comment line
			continue
		}
		var kv []string
		if strings.Contains(line, "=") {
			kv = strings.SplitN(line, "=", 2)
		} else if strings.Contains(line, ":") {
			kv = strings.SplitN(line, ":", 2)
		}
		if len(kv) != 2 {
			return errors.New(".env format error")
		}

		key := strings.TrimSpace(kv[0])
		value := strings.TrimSpace(kv[1])
		envMap[key] = value

		// set env
		log.Info().Str("key", key).Msg("set env")
		os.Setenv(key, value)
	}
	return nil
}

func readFile(path string) ([]byte, error) {
	var err error
	path, err = filepath.Abs(path)
	if err != nil {
		log.Error().Err(err).Str("path", path).Msg("convert absolute path failed")
		return nil, errors.Wrap(code.LoadFileError, err.Error())
	}

	file, err := os.ReadFile(path)
	if err != nil {
		log.Error().Err(err).Msg("read file failed")
		return nil, errors.Wrap(code.LoadFileError, err.Error())
	}
	return file, nil
}
