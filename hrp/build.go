package hrp

import (
	"bufio"
	_ "embed"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"text/template"

	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"

	"github.com/httprunner/funplugin/shared"
	"github.com/httprunner/httprunner/v4/hrp/internal/builtin"
	"github.com/httprunner/httprunner/v4/hrp/internal/version"
)

const (
	funppy                  = `import funppy`
	fungo                   = `"github.com/httprunner/funplugin/fungo"`
	regexPythonFunctionName = `def ([a-zA-Z_]\w*)\(.*\)`
	regexGoImports          = `import \(([\s\S]*?)\)`
	regexGoImport           = `import (\"[\s\S]*\")`
	regexGoFunctionName     = `func ([A-Z][a-zA-Z_]\w*)\(.*\)`
	regexGoFunctionContent  = `func [\s\S]*?\n}`
)

//go:embed internal/scaffold/templates/plugin/debugtalkPythonTemplate
var pyTemplate string

//go:embed internal/scaffold/templates/plugin/debugtalkGoTemplate
var goTemplate string

type TemplateContent struct {
	Version       string   // hrp version
	Fun           string   // funplugin package
	Regexps       *Regexps // match import/function
	Imports       []string // python/go import
	FromImports   []string // python from...import...
	Functions     []string // python/go function
	FunctionNames []string // function name set by user
}

type Regexps struct {
	Import          *regexp.Regexp
	Imports         *regexp.Regexp
	FunctionName    *regexp.Regexp
	FunctionContent *regexp.Regexp // including function define and body
}

func (t *TemplateContent) parseGoContent(path string) error {
	log.Info().Str("path", path).Msg("start to parse debugtalk.go")

	content, err := os.ReadFile(path)
	if err != nil {
		log.Error().Err(err).Msg("failed to read file")
		return err
	}
	originalContent := string(content)

	// parse imports
	importSlice := t.Regexps.Imports.FindAllStringSubmatch(originalContent, -1)
	if len(importSlice) != 0 {
		imports := strings.Replace(importSlice[0][1], "\t", "", -1)
		for _, elem := range strings.Split(imports, "\n") {
			t.Imports = append(t.Imports, strings.TrimSpace(elem))
		}
	}
	// parse import
	importSlice = t.Regexps.Import.FindAllStringSubmatch(originalContent, -1)
	if len(importSlice) != 0 {
		for _, elem := range importSlice {
			t.Imports = append(t.Imports, strings.TrimSpace(elem[1]))
		}
	}
	// import fungo package
	if !builtin.Contains(t.Imports, fungo) {
		t.Imports = append(t.Imports, t.Fun)
	}

	// parse function name
	functionNameSlice := t.Regexps.FunctionName.FindAllStringSubmatch(originalContent, -1)
	for _, elem := range functionNameSlice {
		name := strings.Trim(elem[1], " ")
		if name == "main" {
			continue
		}
		t.FunctionNames = append(t.FunctionNames, name)
	}

	// parse function content
	functionContentSlice := t.Regexps.FunctionContent.FindAllStringSubmatch(originalContent, -1)
	for _, f := range functionContentSlice {
		if strings.Contains(f[0], "func main") {
			continue
		}
		t.Functions = append(t.Functions, strings.Trim(f[0], "\n"))
	}
	return nil
}

func (t *TemplateContent) parsePyContent(path string) error {
	file, err := os.Open(path)
	if err != nil {
		log.Error().Err(err).Str("path", path).Msg("failed to open file")
		return err
	}
	defer file.Close()

	r := bufio.NewReader(file)

	// record content excluding import and main
	content := ""

	// parse python content line by line
	for {
		l, _, err := r.ReadLine()
		if err == io.EOF {
			break
		}
		line := string(l)

		if strings.HasPrefix(line, "import") {
			t.Imports = append(t.Imports, strings.Trim(line, " "))
		} else if strings.HasPrefix(line, "from") {
			t.FromImports = append(t.FromImports, strings.Trim(line, " "))
		} else {
			// no parse content at under of `if __name__ == "__main__"`
			if strings.HasPrefix(line, "if __name__") {
				break
			}
			if strings.HasPrefix(line, "def") {
				functionNameSlice := t.Regexps.FunctionName.FindAllStringSubmatch(line, -1)
				if len(functionNameSlice) == 0 {
					continue
				}
				t.FunctionNames = append(t.FunctionNames, functionNameSlice[0][1])
			}
			content += line + "\n"
		}
	}
	// function content
	t.Functions = append(t.Functions, strings.Trim(content, "\n"))

	// import funppy
	if !builtin.Contains(t.Imports, t.Fun) {
		t.Imports = append(t.Imports, t.Fun)
	}
	return nil
}

func (t *TemplateContent) genDebugTalk(path string, templ string) error {
	file, err := os.Create(path)
	if err != nil {
		log.Error().Err(err).Msg("open file failed")
		return err
	}
	defer file.Close()
	writer := bufio.NewWriter(file)
	tmpl := template.Must(template.New("debugtalk").Parse(templ))
	err = tmpl.Execute(writer, t)
	if err != nil {
		log.Error().Err(err).Msg("execute applies a parsed template to the specified data object failed")
		return err
	}
	err = writer.Flush()
	if err == nil {
		log.Info().Str("path", path).Msg("generate debugtalk success")
	} else {
		log.Error().Str("path", path).Msg("generate debugtalk failed")
	}
	return err
}

// buildGo builds debugtalk.go to debugtalk.bin
func buildGo(path string, output string) error {
	templateContent := &TemplateContent{
		Version: version.VERSION,
		Fun:     fungo,
		Regexps: &Regexps{
			Import:          regexp.MustCompile(regexGoImport),
			Imports:         regexp.MustCompile(regexGoImports),
			FunctionName:    regexp.MustCompile(regexGoFunctionName),
			FunctionContent: regexp.MustCompile(regexGoFunctionContent),
		},
	}

	pluginDir := filepath.Dir(path)

	// check go sdk in tempDir
	if err := builtin.ExecCommandInDir(exec.Command("go", "version"), pluginDir); err != nil {
		return errors.Wrap(err, "go sdk not installed")
	}

	// parse debugtalk.go in pluginDir
	err := templateContent.parseGoContent(path)
	if err != nil {
		return err
	}

	// generate debugtalk.go in pluginDir
	err = templateContent.genDebugTalk(filepath.Join(pluginDir, PluginGoSourceGenFile), goTemplate)
	if err != nil {
		return err
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
			return err
		}
	}

	// add missing and remove unused modules
	if err := builtin.ExecCommandInDir(exec.Command("go", "mod", "tidy"), pluginDir); err != nil {
		return err
	}

	if output == "" {
		dir, _ := os.Getwd()
		output = filepath.Join(dir, PluginHashicorpGoBuiltFile)
	} else if builtin.IsFolderPathExists(output) {
		output = filepath.Join(output, PluginHashicorpGoBuiltFile)
	}
	outputPath, err := filepath.Abs(output)
	if err != nil {
		return err
	}

	// build plugin debugtalk.bin
	cmd := exec.Command("go", "build", "-o", outputPath, PluginGoSourceGenFile, filepath.Base(path))
	if err := builtin.ExecCommandInDir(cmd, pluginDir); err != nil {
		return err
	}
	log.Info().Str("output", outputPath).Str("plugin", path).Msg("build plugin successfully")
	return nil
}

// buildPy completes funppy information in debugtalk.py
func buildPy(path string, output string) error {
	templateContent := &TemplateContent{
		Version: version.VERSION,
		Fun:     funppy,
		Regexps: &Regexps{
			FunctionName: regexp.MustCompile(regexPythonFunctionName),
		},
	}
	err := templateContent.parsePyContent(path)
	if err != nil {
		return err
	}

	// generate .debugtalk_gen.py
	if output == "" {
		dir, _ := os.Getwd()
		output = filepath.Join(dir, PluginPySourceGenFile)
	} else if builtin.IsFolderPathExists(output) {
		output = filepath.Join(output, PluginPySourceGenFile)
	}
	err = templateContent.genDebugTalk(output, pyTemplate)
	return err
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
		log.Error().Err(err).Str("arg", path).Msg("build plugin failed")
		os.Exit(1)
	}
	return nil
}
