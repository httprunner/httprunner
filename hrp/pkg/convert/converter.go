package convert

import (
	_ "embed"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"

	"github.com/httprunner/httprunner/v4/hrp"
	"github.com/httprunner/httprunner/v4/hrp/internal/builtin"
	"github.com/httprunner/httprunner/v4/hrp/internal/env"
	"github.com/httprunner/httprunner/v4/hrp/internal/myexec"
	"github.com/httprunner/httprunner/v4/hrp/internal/sdk"
)

// target testcase format extensions
const (
	suffixJSON   = ".json"
	suffixYAML   = ".yaml"
	suffixGoTest = ".go"
	suffixPyTest = ".py"
)

type OutputType int

const (
	OutputTypeJSON OutputType = iota // default output type: JSON
	OutputTypeYAML
	OutputTypeGoTest
	OutputTypePyTest
)

func (outputType OutputType) String() string {
	switch outputType {
	case OutputTypeYAML:
		return "yaml"
	case OutputTypeGoTest:
		return "gotest"
	case OutputTypePyTest:
		return "pytest"
	default:
		return "json"
	}
}

// Profile is used to override or update(create if not existed) original headers and cookies
type Profile struct {
	Override bool              `json:"override" yaml:"override"`
	Headers  map[string]string `json:"headers" yaml:"headers"`
	Cookies  map[string]string `json:"cookies" yaml:"cookies"`
}

func Run(outputType OutputType, outputDir, profilePath string, args []string) {
	// report event
	sdk.SendEvent(sdk.EventTracking{
		Category: "ConvertTests",
		Action:   fmt.Sprintf("hrp convert --to-%s", outputType.String()),
	})

	var outputFiles []string
	for _, inputSample := range args {
		// loads source file and convert to TCase format
		tCase, err := LoadTCase(inputSample)
		if err != nil {
			log.Warn().Err(err).Str("input sample", inputSample).Msg("convert input sample failed")
			continue
		}

		caseConverter := &TCaseConverter{
			InputSample: inputSample,
			OutputDir:   outputDir,
			TCase:       tCase,
		}

		// override TCase with profile
		if profilePath != "" {
			caseConverter.overrideWithProfile(profilePath)
		}

		// convert TCase format to target case format
		var outputFile string
		switch outputType {
		case OutputTypeYAML:
			outputFile, err = caseConverter.ToYAML()
		case OutputTypeGoTest:
			outputFile, err = caseConverter.ToGoTest()
		case OutputTypePyTest:
			outputFile, err = caseConverter.ToPyTest()
		default:
			outputFile, err = caseConverter.ToJSON()
		}
		if err != nil {
			log.Error().Err(err).
				Str("input sample", caseConverter.InputSample).
				Msg("convert case failed")
			continue
		}
		outputFiles = append(outputFiles, outputFile)
	}
	log.Info().Strs("output files", outputFiles).Msg("conversion completed")
}

// LoadTCase loads source file and convert to TCase type
func LoadTCase(inputSample string) (*hrp.TCase, error) {
	if strings.HasPrefix(inputSample, "curl ") {
		// 'path' contains curl command
		curlCase, err := LoadSingleCurlCase(inputSample)
		if err != nil {
			return nil, err
		}
		return curlCase, nil
	}
	extName := filepath.Ext(inputSample)
	if extName == "" {
		return nil, errors.New("file extension is not specified")
	}
	switch extName {
	case ".har":
		tCase, err := LoadHARCase(inputSample)
		if err != nil {
			return nil, err
		}
		return tCase, nil
	case ".json":
		// priority: hrp JSON case > postman > swagger
		// check if hrp JSON case
		tCase, err := LoadJSONCase(inputSample)
		if err == nil {
			return tCase, nil
		}

		// check if postman format
		casePostman, err := LoadPostmanCase(inputSample)
		if err == nil {
			return casePostman, nil
		}

		// check if swagger format
		caseSwagger, err := LoadSwaggerCase(inputSample)
		if err == nil {
			return caseSwagger, nil
		}

		return nil, errors.New("unexpected JSON format")
	case ".yaml", ".yml":
		// priority: hrp YAML case > swagger
		// check if hrp YAML case
		tCase, err := NewYAMLCase(inputSample)
		if err == nil {
			return tCase, nil
		}

		// check if swagger format
		caseSwagger, err := LoadSwaggerCase(inputSample)
		if err == nil {
			return caseSwagger, nil
		}

		return nil, errors.New("unexpected YAML format")
	case ".go": // TODO
		return nil, errors.New("convert gotest is not implemented")
	case ".py": // TODO
		return nil, errors.New("convert pytest is not implemented")
	case ".jmx": // TODO
		return nil, errors.New("convert JMeter jmx is not implemented")
	case ".txt":
		curlCase, err := LoadCurlCase(inputSample)
		if err != nil {
			return nil, err
		}
		return curlCase, nil
	}

	return nil, fmt.Errorf("unsupported file type: %v", extName)
}

// TCaseConverter holds the common properties of case converter
type TCaseConverter struct {
	InputSample string
	OutputDir   string
	TCase       *hrp.TCase
}

func (c *TCaseConverter) genOutputPath(suffix string) string {
	var outFileFullName string
	if curlCmd := strings.TrimSpace(c.InputSample); strings.HasPrefix(curlCmd, "curl ") {
		outFileFullName = fmt.Sprintf("curl_%v_test%v", env.StartTimeStr, suffix)
		if c.OutputDir != "" {
			return filepath.Join(c.OutputDir, outFileFullName)
		} else {
			return filepath.Join(env.RootDir, outFileFullName)
		}
	}
	outFileFullName = builtin.GetFileNameWithoutExtension(c.InputSample) + "_test" + suffix
	if c.OutputDir != "" {
		return filepath.Join(c.OutputDir, outFileFullName)
	} else {
		return filepath.Join(filepath.Dir(c.InputSample), outFileFullName)
	}
	// TODO avoid outFileFullName conflict?
}

// convert TCase to pytest case
func (c *TCaseConverter) ToPyTest() (string, error) {
	jsonPath, err := c.ToJSON()
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

// TODO: convert TCase to gotest case
func (c *TCaseConverter) ToGoTest() (string, error) {
	return "", nil
}

// convert TCase to JSON case
func (c *TCaseConverter) ToJSON() (string, error) {
	jsonPath := c.genOutputPath(suffixJSON)
	err := builtin.Dump2JSON(c.TCase, jsonPath)
	if err != nil {
		return "", err
	}
	return jsonPath, nil
}

// convert TCase to YAML case
func (c *TCaseConverter) ToYAML() (string, error) {
	yamlPath := c.genOutputPath(suffixYAML)
	err := builtin.Dump2YAML(c.TCase, yamlPath)
	if err != nil {
		return "", err
	}
	return yamlPath, nil
}

func (c *TCaseConverter) overrideWithProfile(path string) error {
	log.Info().Str("path", path).Msg("load profile")
	profile := new(Profile)
	err := builtin.LoadFile(path, profile)
	if err != nil {
		log.Warn().Str("path", path).
			Msg("failed to load profile, ignore!")
		return err
	}

	log.Info().Interface("profile", profile).Msg("override with profile")
	for _, step := range c.TCase.TestSteps {
		// override original headers and cookies
		if profile.Override {
			step.Request.Headers = make(map[string]string)
			step.Request.Cookies = make(map[string]string)
		}
		// update (create if not existed) original headers and cookies
		if step.Request.Headers == nil {
			step.Request.Headers = make(map[string]string)
		}
		if step.Request.Cookies == nil {
			step.Request.Cookies = make(map[string]string)
		}
		for k, v := range profile.Headers {
			step.Request.Headers[k] = v
		}
		for k, v := range profile.Cookies {
			step.Request.Cookies[k] = v
		}
	}
	return nil
}
