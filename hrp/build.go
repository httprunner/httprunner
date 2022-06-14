package hrp

import (
	"bufio"
	_ "embed"
	"fmt"
	"html/template"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/httprunner/funplugin/shared"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"

	"github.com/httprunner/httprunner/v4/hrp/internal/builtin"
	"github.com/httprunner/httprunner/v4/hrp/internal/version"
)

//go:embed internal/scaffold/templates/plugin/debugtalkPythonTemplate
var pyTemplate string

//go:embed internal/scaffold/templates/plugin/debugtalkGoTemplate
var goTemplate string

// regex for finding all function names
var (
	regexPyFunctionName = regexp.MustCompile(`def ([a-zA-Z_]\w*)\(.*\)`)
	regexGoFunctionName = regexp.MustCompile(`func ([A-Z][a-zA-Z_]\w*)\(.*\)`)
)

type pluginTemplateContent struct {
	Version       string   // hrp version
	FunctionNames []string // function names
}

func findAllFunctionNames(t *regexp.Regexp, path string) (functionNames []string, err error) {
	log.Info().Str("path", path).Msg("find all function names from plugin file")

	content, err := os.ReadFile(path)
	if err != nil {
		log.Error().Err(err).Msg("failed to read file")
		return nil, errors.Wrap(err, "read file failed")
	}

	// find all function names
	functionNameSlice := t.FindAllStringSubmatch(string(content), -1)
	for _, elem := range functionNameSlice {
		name := strings.Trim(elem[1], " ")
		functionNames = append(functionNames, name)
	}

	return
}

func generate(data *pluginTemplateContent, tmpl, output string) error {
	file, err := os.Create(output)
	if err != nil {
		log.Error().Err(err).Msg("open output file failed")
		return err
	}
	defer file.Close()

	writer := bufio.NewWriter(file)
	err = template.Must(template.New("debugtalk").Parse(tmpl)).Execute(writer, data)
	if err != nil {
		log.Error().Err(err).Msg("execute template parsing failed")
		return err
	}

	err = writer.Flush()
	if err == nil {
		log.Info().Str("output", output).Msg("generate debugtalk success")
	} else {
		log.Error().Str("output", output).Msg("generate debugtalk failed")
	}
	return err
}

// buildGo builds debugtalk.go to debugtalk.bin
func buildGo(path string, output string) error {
	functionNames, err := findAllFunctionNames(regexGoFunctionName, path)
	if err != nil {
		return errors.Wrap(err, "find all function names failed")
	}
	// filter main and init function
	var filteredFunctionNames []string
	for _, name := range functionNames {
		if name == "main" || name == "init" {
			continue
		}
		filteredFunctionNames = append(filteredFunctionNames, name)
	}

	templateContent := &pluginTemplateContent{
		Version:       version.VERSION,
		FunctionNames: filteredFunctionNames,
	}

	pluginDir := filepath.Dir(path)
	err = generate(templateContent, goTemplate, filepath.Join(pluginDir, PluginGoSourceGenFile))
	if err != nil {
		return errors.Wrap(err, "generate hashicorp plugin failed")
	}

	// check go sdk in tempDir
	if err := builtin.ExecCommandInDir(exec.Command("go", "version"), pluginDir); err != nil {
		return errors.Wrap(err, "go sdk not installed")
	}

	if !builtin.IsFilePathExists(filepath.Join(pluginDir, "go.mod")) {
		// create go mod
		if err := builtin.ExecCommandInDir(exec.Command("go", "mod", "init", "main"), pluginDir); err != nil {
			return err
		}

		// download plugin dependency
		// funplugin version should be locked
		funplugin := fmt.Sprintf("github.com/httprunner/funplugin@%s", shared.Version)
		if err := builtin.ExecCommandInDir(exec.Command("go", "get", funplugin), pluginDir); err != nil {
			return errors.Wrap(err, "go get funplugin failed")
		}
	}

	// add missing and remove unused modules
	if err := builtin.ExecCommandInDir(exec.Command("go", "mod", "tidy"), pluginDir); err != nil {
		return errors.Wrap(err, "go mod tidy failed")
	}

	// specify output file path
	if output == "" {
		dir, _ := os.Getwd()
		output = filepath.Join(dir, PluginHashicorpGoBuiltFile)
	} else if builtin.IsFolderPathExists(output) {
		output = filepath.Join(output, PluginHashicorpGoBuiltFile)
	}
	outputPath, _ := filepath.Abs(output)

	// build go plugin to debugtalk.bin
	cmd := exec.Command("go", "build", "-o", outputPath, PluginGoSourceGenFile, filepath.Base(path))
	if err := builtin.ExecCommandInDir(cmd, pluginDir); err != nil {
		return errors.Wrap(err, "go build plugin failed")
	}
	log.Info().Str("output", outputPath).Str("plugin", path).Msg("build go plugin successfully")
	return nil
}

// buildPy completes funppy information in debugtalk.py
func buildPy(path string, output string) error {
	// check the syntax of debugtalk.py
	err := builtin.ExecCommand("python3", "-m", "py_compile", path)
	if err != nil {
		return errors.Wrap(err, "python plugin syntax invalid")
	}

	functionNames, err := findAllFunctionNames(regexPyFunctionName, path)
	if err != nil {
		return err
	}
	templateContent := &pluginTemplateContent{
		Version:       version.VERSION,
		FunctionNames: functionNames,
	}

	// specify output file path
	if output == "" {
		dir, _ := os.Getwd()
		output = filepath.Join(dir, PluginPySourceGenFile)
	} else if builtin.IsFolderPathExists(output) {
		output = filepath.Join(output, PluginPySourceGenFile)
	}

	// generate .debugtalk_gen.py
	err = generate(templateContent, pyTemplate, output)
	if err != nil {
		return err
	}

	log.Info().Str("output", output).Str("plugin", path).Msg("build python plugin successfully")
	return nil
}

func BuildPlugin(path string, output string) (err error) {
	ext := filepath.Ext(path)
	switch ext {
	case ".py":
		err = buildPy(path, output)
	case ".go":
		err = buildGo(path, output)
	default:
		return errors.New("type error, expected .py or .go")
	}
	if err != nil {
		log.Error().Err(err).Str("path", path).Msg("build plugin failed")
		os.Exit(1)
	}
	return nil
}
