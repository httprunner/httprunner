package convert

import (
	_ "embed"
	"fmt"
	"path/filepath"

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
	suffixHAR    = ".har"
)

type FromType int

const (
	FromTypeJSON FromType = iota
	FromTypeYAML
	FromTypeHAR
	FromTypePostman
	FromTypeCurl
	FromTypeSwagger
	FromTypePyest
	FromTypeGotest
)

func (fromType FromType) String() string {
	switch fromType {
	case FromTypeYAML:
		return "yaml"
	case FromTypeHAR:
		return "har"
	case FromTypePostman:
		return "postman"
	case FromTypeSwagger:
		return "swagger"
	case FromTypeCurl:
		return "curl"
	case FromTypeGotest:
		return "gotest"
	case FromTypePyest:
		return "pytest"
	default:
		return "json"
	}
}

func (fromType FromType) Extensions() []string {
	switch fromType {
	case FromTypeYAML:
		return []string{suffixYAML, ".yml"}
	case FromTypeHAR:
		return []string{suffixHAR}
	case FromTypePostman, FromTypeSwagger:
		return []string{suffixJSON}
	case FromTypeCurl:
		return []string{".txt", ".curl"}
	case FromTypeGotest:
		return []string{suffixGoTest}
	case FromTypePyest:
		return []string{suffixPyTest}
	default:
		return []string{suffixJSON}
	}
}

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

func NewConverter(outputDir, profilePath string) *TCaseConverter {
	return &TCaseConverter{
		profilePath: profilePath,
		outputDir:   outputDir,
	}
}

// TCaseConverter holds the common properties of case converter
type TCaseConverter struct {
	fromFile    string
	profilePath string
	outputDir   string
	tCase       *hrp.TCase
}

// LoadCase loads source file and convert to TCase type
func (c *TCaseConverter) loadCase(casePath string, fromType FromType) error {
	c.fromFile = casePath
	var err error
	switch fromType {
	case FromTypeJSON:
		c.tCase, err = LoadJSONCase(casePath)
	case FromTypeYAML:
		c.tCase, err = LoadYAMLCase(casePath)
	case FromTypeHAR:
		c.tCase, err = LoadHARCase(casePath)
	case FromTypePostman:
		c.tCase, err = LoadPostmanCase(casePath)
	case FromTypeSwagger:
		c.tCase, err = LoadSwaggerCase(casePath)
	case FromTypeCurl:
		c.tCase, err = LoadCurlCase(casePath)
	}
	return err
}

func (c *TCaseConverter) Convert(casePath string, fromType FromType, outputType OutputType) error {
	// report event
	sdk.SendEvent(sdk.EventTracking{
		Category: "ConvertTests",
		Action:   fmt.Sprintf("hrp convert --to-%s", outputType.String()),
	})
	log.Info().Str("path", casePath).
		Str("fromType", fromType.String()).
		Str("outputType", outputType.String()).
		Msg("convert testcase")

	// load source file
	err := c.loadCase(casePath, fromType)
	if err != nil {
		return err
	}

	// override TCase with profile
	if c.profilePath != "" {
		c.overrideWithProfile(c.profilePath)
	}

	// convert to target format
	var outputFile string
	switch outputType {
	case OutputTypeYAML:
		outputFile, err = c.toYAML()
	case OutputTypeGoTest:
		outputFile, err = c.toGoTest()
	case OutputTypePyTest:
		outputFile, err = c.toPyTest()
	default:
		outputFile, err = c.toJSON()
	}
	if err != nil {
		return err
	}

	log.Info().Str("outputFile", outputFile).Msg("conversion completed")
	return nil
}

func (c *TCaseConverter) genOutputPath(suffix string) string {
	outFileFullName := builtin.GetFileNameWithoutExtension(c.fromFile) + "_test" + suffix
	if c.outputDir != "" {
		return filepath.Join(c.outputDir, outFileFullName)
	} else {
		return filepath.Join(filepath.Dir(c.fromFile), outFileFullName)
	}
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
	for _, step := range c.tCase.TestSteps {
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
