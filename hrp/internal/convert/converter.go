package convert

import (
	_ "embed"
	"fmt"
	"path/filepath"

	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"

	"github.com/httprunner/httprunner/v4/hrp"
	"github.com/httprunner/httprunner/v4/hrp/internal/builtin"
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
	for _, path := range args {
		// loads source file and convert to TCase format
		tCase, err := LoadTCase(path)
		if err != nil {
			log.Warn().Err(err).Str("path", path).Msg("convert source file failed")
			continue
		}

		caseConverter := &TCaseConverter{
			SourcePath: path,
			OutputDir:  outputDir,
			TCase:      tCase,
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
				Str("source path", path).
				Msg("convert case failed")
			continue
		}
		outputFiles = append(outputFiles, outputFile)
	}
	log.Info().Strs("output files", outputFiles).Msg("conversion completed")
}

// LoadTCase loads source file and convert to TCase type
func LoadTCase(path string) (*hrp.TCase, error) {
	extName := filepath.Ext(path)
	if extName == "" {
		return nil, errors.New("file extension is not specified")
	}
	switch extName {
	case ".har":
		tCase, err := LoadHARCase(path)
		if err != nil {
			return nil, err
		}
		return tCase, nil
	case ".json":
		// priority: hrp JSON case > postman > swagger
		// check if hrp JSON case
		tCase, err := LoadJSONCase(path)
		if err == nil {
			return tCase, nil
		}

		// check if postman format
		casePostman, err := LoadPostmanCase(path)
		if err == nil {
			return casePostman, nil
		}

		// check if swagger format
		caseSwagger, err := LoadSwaggerCase(path)
		if err == nil {
			return caseSwagger, nil
		}

		return nil, errors.New("unexpected JSON format")
	case ".yaml", ".yml":
		// priority: hrp YAML case > swagger
		// check if hrp YAML case
		tCase, err := NewYAMLCase(path)
		if err == nil {
			return tCase, nil
		}

		// check if swagger format
		caseSwagger, err := LoadSwaggerCase(path)
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
	}

	return nil, fmt.Errorf("unsupported file type: %v", extName)
}

// TCaseConverter holds the common properties of case converter
type TCaseConverter struct {
	SourcePath string
	OutputDir  string
	TCase      *hrp.TCase
}

func (c *TCaseConverter) genOutputPath(suffix string) string {
	outFileFullName := builtin.GetOutputNameWithoutExtension(c.SourcePath) + suffix
	if c.OutputDir != "" {
		return filepath.Join(c.OutputDir, outFileFullName)
	} else {
		return filepath.Join(filepath.Dir(c.SourcePath), outFileFullName)
	}
	// TODO avoid outFileFullName conflict?
}

// convert TCase to pytest case
func (c *TCaseConverter) ToPyTest() (string, error) {
	args := append([]string{"make"}, c.SourcePath)
	err := builtin.ExecPython3Command("httprunner", args...)
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
