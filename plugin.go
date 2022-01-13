package hrp

import (
	"fmt"
	"os"
	"path/filepath"
	"plugin"
	"reflect"
	"runtime"

	"github.com/rs/zerolog/log"

	"github.com/httprunner/hrp/internal/builtin"
	"github.com/httprunner/hrp/internal/ga"
)

type pluginFile string

const (
	goPluginFile          pluginFile = "debugtalk.so" // built from go plugin
	hashicorpGoPluginFile pluginFile = "debugtalk"    // built from hashicorp go plugin
	hashicorpPyPluginFile pluginFile = "debugtalk.py"
)

type hrpPlugin interface {
	init(path string) error
	lookup(funcName string) (reflect.Value, error) // lookup function
	// call(funcName string, args ...interface{}) (interface{}, error)
	quit() error
}

// goPlugin implements golang official plugin
type goPlugin struct {
	*plugin.Plugin
}

func (p *goPlugin) init(path string) error {
	if runtime.GOOS == "windows" {
		log.Warn().Msg("go plugin does not support windows")
		return fmt.Errorf("go plugin does not support windows")
	}

	var err error
	// report event for loading go plugin
	defer func() {
		event := ga.EventTracking{
			Category: "LoadGoPlugin",
			Action:   "plugin.Open",
		}
		if err != nil {
			event.Value = 1 // failed
		}
		go ga.SendEvent(event)
	}()

	p.Plugin, err = plugin.Open(path)
	if err != nil {
		log.Error().Err(err).Str("path", path).Msg("load go plugin failed")
		return err
	}

	log.Info().Str("path", path).Msg("load go plugin success")
	return nil
}

func (p *goPlugin) lookup(funcName string) (reflect.Value, error) {
	if p.Plugin == nil {
		return reflect.Value{}, fmt.Errorf("go plugin is not loaded")
	}

	sym, err := p.Plugin.Lookup(funcName)
	if err != nil {
		return reflect.Value{}, fmt.Errorf("function %s is not found", funcName)
	}
	fn := reflect.ValueOf(sym)

	// check function type
	if fn.Kind() != reflect.Func {
		return reflect.Value{}, fmt.Errorf("function %s is invalid", funcName)
	}

	return fn, nil
}

func (p *goPlugin) call(funcName string, args ...interface{}) (interface{}, error) {
	if p.Plugin == nil {
		return nil, fmt.Errorf("go plugin is not loaded")
	}
	return nil, nil
}

func (p *goPlugin) quit() error {
	return nil
}

// hashicorpPlugin implements hashicorp/go-plugin
type hashicorpPlugin struct {
	cachedFunctions map[string]reflect.Value
}

func (p *hashicorpPlugin) init(path string) error {
	log.Info().Str("path", path).Msg("load hashicorp go plugin success")
	return nil
}

func (p *hashicorpPlugin) lookup(funcName string) (reflect.Value, error) {
	return reflect.Value{}, nil
}

func (p *hashicorpPlugin) call(funcName string, args ...interface{}) (interface{}, error) {
	return nil, nil
}

func (p *hashicorpPlugin) quit() error {
	return nil
}

func (p *parser) initPlugin(path string) error {
	if path == "" {
		return nil
	}

	// locate go plugin file
	pluginPath, err := locatePlugin(path, goPluginFile)
	if err == nil {
		// found go plugin file
		p.plugin = &goPlugin{}
		return p.plugin.init(pluginPath)
	}

	// locate hashicorp plugin file
	pluginPath, err = locatePlugin(path, hashicorpGoPluginFile)
	if err == nil {
		// found hashicorp go plugin file
		p.plugin = &hashicorpPlugin{}
		return p.plugin.init(pluginPath)
	}

	// plugin not found
	return nil
}

// locatePlugin searches destPluginFile upward recursively until current
// working directory or system root dir.
func locatePlugin(startPath string, destPluginFile pluginFile) (string, error) {
	stat, err := os.Stat(startPath)
	if os.IsNotExist(err) {
		return "", err
	}

	var startDir string
	if stat.IsDir() {
		startDir = startPath
	} else {
		startDir = filepath.Dir(startPath)
	}
	startDir, _ = filepath.Abs(startDir)

	// convention over configuration
	pluginPath := filepath.Join(startDir, string(destPluginFile))
	if _, err := os.Stat(pluginPath); err == nil {
		return pluginPath, nil
	}

	// current working directory
	cwd, _ := os.Getwd()
	if startDir == cwd {
		return "", fmt.Errorf("searched to CWD, plugin file not found")
	}

	// system root dir
	parentDir, _ := filepath.Abs(filepath.Dir(startDir))
	if parentDir == startDir {
		return "", fmt.Errorf("searched to system root dir, plugin file not found")
	}

	return locatePlugin(parentDir, destPluginFile)
}

func (p *parser) getMappingFunction(funcName string) (reflect.Value, error) {
	var fn reflect.Value

	// get function from plugin
	if p.plugin != nil {
		fn, err := p.plugin.lookup(funcName)
		if err == nil {
			return fn, nil
		}
	}

	// get builtin function
	if function, ok := builtin.Functions[funcName]; ok {
		fn = reflect.ValueOf(function)
		return fn, nil
	}

	// function not found
	return reflect.Value{}, fmt.Errorf("function %s is not found", funcName)
}

// callFunc calls function with arguments
// only support return at most one result value
func (p *parser) callFunc(funcName string, arguments ...interface{}) (interface{}, error) {
	fn, err := p.getMappingFunction(funcName)
	if err != nil {
		return nil, err
	}

	if fn.Type().NumIn() != len(arguments) {
		// function arguments not match
		return nil, fmt.Errorf("function arguments number not match")
	}

	argumentsValue := make([]reflect.Value, len(arguments))
	for index, argument := range arguments {
		argumentValue := reflect.ValueOf(argument)
		expectArgumentType := fn.Type().In(index)
		actualArgumentType := reflect.TypeOf(argument)

		// type match
		if expectArgumentType == actualArgumentType {
			argumentsValue[index] = argumentValue
			continue
		}

		// type not match, check if convertible
		if !actualArgumentType.ConvertibleTo(expectArgumentType) {
			// function argument type not match and not convertible
			err := fmt.Errorf("function argument %d's type is neither match nor convertible, expect %v, actual %v",
				index, expectArgumentType, actualArgumentType)
			return nil, err
		}
		// convert argument to expect type
		argumentsValue[index] = argumentValue.Convert(expectArgumentType)
	}

	resultValues := fn.Call(argumentsValue)
	if len(resultValues) > 1 {
		// function should return at most one value
		err := fmt.Errorf("function should return at most one value")
		return nil, err
	}

	// no return value
	if len(resultValues) == 0 {
		return nil, nil
	}

	// return one value
	// convert reflect.Value to interface{}
	result := resultValues[0].Interface()
	return result, nil
}
