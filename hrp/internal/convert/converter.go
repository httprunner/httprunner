package convert

import (
	_ "embed"
	"fmt"
	"os"
	"path/filepath"
	"reflect"

	"github.com/go-openapi/spec"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"

	"github.com/httprunner/httprunner/v4/hrp"
	"github.com/httprunner/httprunner/v4/hrp/internal/builtin"
	"github.com/httprunner/httprunner/v4/hrp/internal/sdk"
)

const (
	suffixJSON   = ".json"
	suffixYAML   = ".yaml"
	suffixGoTest = ".go"
	suffixPyTest = ".py"
)

type InputType int

const (
	InputTypeUnknown InputType = iota // default input type: unknown
	InputTypeHAR
	InputTypePostman
	InputTypeSwagger
	InputTypeJMeter
	InputTypeJSON
	InputTypeYAML
	InputTypeGoTest
	InputTypePyTest
)

func (inputType InputType) String() string {
	switch inputType {
	case InputTypeHAR:
		return "har"
	case InputTypePostman:
		return "postman"
	case InputTypeSwagger:
		return "swagger"
	case InputTypeJMeter:
		return "jmeter"
	case InputTypeJSON:
		return "json testcase"
	case InputTypeYAML:
		return "yaml testcase"
	case InputTypeGoTest:
		return "gotest script"
	case InputTypePyTest:
		return "pytest script"
	default:
		return "unknown"
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

// TCaseConverter holds the common properties of case converter
type TCaseConverter struct {
	InputPath   string
	OutputDir   string
	Profile     *Profile
	InputType   InputType
	OutputType  OutputType
	CaseHAR     *CaseHar
	CasePostman *CasePostman
	CaseSwagger *spec.Swagger
	TCase       *hrp.TCase
}

// Profile is used to override or update(create if not existed) original headers and cookies
type Profile struct {
	Override bool              `json:"override" yaml:"override"`
	Headers  map[string]string `json:"headers" yaml:"headers"`
	Cookies  map[string]string `json:"cookies" yaml:"cookies"`
}

func NewTCaseConverter(path string) (tCaseConverter *TCaseConverter) {
	tCaseConverter = &TCaseConverter{
		InputPath: path,
		InputType: InputTypeUnknown,
	}
	extName := filepath.Ext(path)
	if extName == "" {
		log.Warn().Msg("extension name should be specified")
		return
	}
	var err error
	switch extName {
	case ".har":
		caseHAR := new(CaseHar)
		err = builtin.LoadFile(path, caseHAR)
		if err == nil && !reflect.ValueOf(*caseHAR).IsZero() {
			tCaseConverter.InputType = InputTypeHAR
			tCaseConverter.CaseHAR = caseHAR
		}
	case ".json":
		tCase := new(hrp.TCase)
		err = builtin.LoadFile(path, tCase)
		if err == nil && !reflect.ValueOf(*tCase).IsZero() {
			tCaseConverter.InputType = InputTypeJSON
			tCaseConverter.TCase = tCase
			break
		}
		casePostman := new(CasePostman)
		err = builtin.LoadFile(path, casePostman)
		// deal with postman field name conflict with swagger
		descriptionBackup := casePostman.Info.Description
		casePostman.Info.Description = ""
		if err == nil && !reflect.ValueOf(*casePostman).IsZero() {
			tCaseConverter.InputType = InputTypePostman
			casePostman.Info.Description = descriptionBackup
			tCaseConverter.CasePostman = casePostman
			break
		}
		caseSwagger := new(spec.Swagger)
		err = builtin.LoadFile(path, caseSwagger)
		if err == nil && !reflect.ValueOf(*caseSwagger).IsZero() {
			tCaseConverter.InputType = InputTypeSwagger
			tCaseConverter.CaseSwagger = caseSwagger
		}
	case ".yaml", ".yml":
		tCase := new(hrp.TCase)
		err = builtin.LoadFile(path, tCase)
		if err == nil && !reflect.ValueOf(*tCase).IsZero() {
			tCaseConverter.InputType = InputTypeYAML
			tCaseConverter.TCase = tCase
			break
		}
		caseSwagger := new(spec.Swagger)
		err = builtin.LoadFile(path, caseSwagger)
		if err == nil && !reflect.ValueOf(*caseSwagger).IsZero() {
			tCaseConverter.InputType = InputTypeSwagger
			tCaseConverter.CaseSwagger = caseSwagger
		}
	case ".go": // TODO
		tCaseConverter.InputType = InputTypeGoTest
	case ".py": // TODO
		tCaseConverter.InputType = InputTypePyTest
	case ".jmx": // TODO
		tCaseConverter.InputType = InputTypeJMeter
	default:
		log.Warn().
			Str("input path", tCaseConverter.InputPath).
			Msgf("unsupported file type: %v", extName)
	}
	if tCaseConverter.InputType != InputTypeUnknown {
		log.Info().
			Str("input path", tCaseConverter.InputPath).
			Msgf("load case as: %s", tCaseConverter.InputType.String())
	} else {
		log.Error().Err(err).
			Str("input path", tCaseConverter.InputPath).
			Msgf("failed to load case")
	}
	return
}

func (c *TCaseConverter) SetProfile(path string) {
	log.Info().Str("input path", c.InputPath).Str("profile", path).Msg("set profile")
	profile := new(Profile)
	err := builtin.LoadFile(path, profile)
	if err != nil {
		log.Warn().Str("path", path).
			Msg("failed to load profile, ignore!")
		return
	}
	c.Profile = profile
}

func (c *TCaseConverter) SetOutputDir(dir string) {
	log.Info().Str("input path", c.InputPath).Str("output directory", dir).Msg("set output directory")
	c.OutputDir = dir
}

func (c *TCaseConverter) genOutputPath(suffix string) string {
	outFileFullName := builtin.GetOutputNameWithoutExtension(c.InputPath) + suffix
	if c.OutputDir != "" {
		return filepath.Join(c.OutputDir, outFileFullName)
	} else {
		return filepath.Join(filepath.Dir(c.InputPath), outFileFullName)
	}
	// TODO avoid outFileFullName conflict?
}

func (c *TCaseConverter) ToPyTest() (string, error) {
	script := convertConfig(c.TCase.Config)
	println(script)
	return script, nil
}

func convertConfig(config *hrp.TConfig) string {
	script := fmt.Sprintf("Config('%s')", config.Name)

	if config.Variables != nil {
		script += fmt.Sprintf(".variables(**{%v})", config.Variables)
	}
	if config.BaseURL != "" {
		script += fmt.Sprintf(".base_url('%s')", config.BaseURL)
	}
	if config.Export != nil {
		script += fmt.Sprintf(".export(*%v)", config.Export)
	}
	script += fmt.Sprintf(".verify(%v)", config.Verify)

	return script
}

func (c *TCaseConverter) ToGoTest() (string, error) {
	return "", nil
}

// ICaseConverter represents all kinds of case converters which could convert case into JSON/YAML/gotest/pytest format
type ICaseConverter interface {
	Struct() *TCaseConverter
	ToJSON() (string, error)
	ToYAML() (string, error)
	ToGoTest() (string, error)
	ToPyTest() (string, error)
}

func Run(outputType OutputType, outputDir, profilePath string, args []string) {
	// report event
	sdk.SendEvent(sdk.EventTracking{
		Category: "ConvertTests",
		Action:   fmt.Sprintf("hrp convert --to-%s", outputType.String()),
	})

	// identify input and load converters
	var iCaseConverters []ICaseConverter
	for _, arg := range args {
		tCaseConverter := NewTCaseConverter(arg)
		tCaseConverter.OutputType = outputType
		if outputDir != "" {
			tCaseConverter.SetOutputDir(outputDir)
		}
		if profilePath != "" {
			tCaseConverter.SetProfile(profilePath)
		}
		switch tCaseConverter.InputType {
		case InputTypeHAR:
			iCaseConverters = append(iCaseConverters, NewConverterHAR(tCaseConverter))
		case InputTypePostman:
			iCaseConverters = append(iCaseConverters, NewConverterPostman(tCaseConverter))
		case InputTypeJSON:
			iCaseConverters = append(iCaseConverters, NewConverterJSON(tCaseConverter))
		case InputTypeYAML:
			iCaseConverters = append(iCaseConverters, NewConverterYAML(tCaseConverter))
		case InputTypeSwagger, InputTypeJMeter, InputTypeGoTest, InputTypePyTest:
			log.Warn().
				Str("input path", tCaseConverter.InputPath).
				Msg("case type not supported yet, ignore!")
		default:
			log.Warn().
				Str("input path", tCaseConverter.InputPath).
				Msg("unknown case type, ignore!")
		}
	}

	// start converting
	var outputFiles []string
	var err error
	for _, iCaseConverter := range iCaseConverters {
		log.Info().Str("input path", iCaseConverter.Struct().InputPath).Msg("start converting")
		var outputFile string
		switch iCaseConverter.Struct().OutputType {
		case OutputTypeYAML:
			outputFile, err = iCaseConverter.ToYAML()
		case OutputTypeGoTest:
			outputFile, err = iCaseConverter.ToGoTest()
		case OutputTypePyTest:
			outputFile, err = iCaseConverter.ToPyTest()
		default:
			outputFile, err = iCaseConverter.ToJSON()
		}
		if err != nil {
			log.Error().Err(err).
				Str("input path", iCaseConverter.Struct().InputPath).
				Msg("error occurs during converting")
			continue
		}
		outputFiles = append(outputFiles, outputFile)
	}
	log.Info().Strs("output files", outputFiles).Msg("conversion completed")
}

func makeTestCaseFromJSONYAML(iCaseConverter ICaseConverter) (*hrp.TCase, error) {
	tCase := iCaseConverter.Struct().TCase
	if tCase == nil {
		return nil, errors.Errorf("empty json/yaml testcase occurs")
	}
	profile := iCaseConverter.Struct().Profile
	if profile == nil {
		return tCase, nil
	}
	for _, step := range tCase.TestSteps {
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
	return tCase, nil
}

func convertToPyTest(iCaseConverter ICaseConverter) (string, error) {
	// convert to temporary json testcase
	jsonPath, err := iCaseConverter.ToJSON()
	inputType := iCaseConverter.Struct().InputType
	if err != nil {
		return "", errors.Wrapf(err, "(%s -> pytest step 1) failed to convert to temporary json testcase", inputType.String())
	}
	defer func() {
		if jsonPath != "" {
			if err = os.Remove(jsonPath); err != nil {
				log.Error().Err(err).Msgf("(%s -> pytest step defer) failed to clean temporary json testcase", inputType.String())
			}
		}
	}()

	// convert from temporary json testcase to pytest
	converterJSON := NewConverterJSON(NewTCaseConverter(jsonPath))
	pyTestPath, err := converterJSON.MakePyTestScript()
	if err != nil {
		return "", errors.Wrap(err, "(json -> pytest step 2) failed to convert from temporary json testcase to pytest ")
	}

	// rename resultant pytest
	renamedPyTestPath := iCaseConverter.Struct().genOutputPath(suffixPyTest)
	err = os.Rename(pyTestPath, renamedPyTestPath)
	if err != nil {
		log.Error().Err(err).Msg("(json -> pytest step 3) failed to rename the resultant pytest file")
		return pyTestPath, nil
	}
	return renamedPyTestPath, nil
}
