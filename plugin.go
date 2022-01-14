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
	pluginHost "github.com/httprunner/hrp/plugin/host"
	pluginShared "github.com/httprunner/hrp/plugin/shared"
)

type pluginFile string

const (
	goPluginFile          pluginFile = pluginShared.Name + ".so"  // built from go plugin
	hashicorpGoPluginFile pluginFile = pluginShared.Name + ".bin" // built from hashicorp go plugin
	hashicorpPyPluginFile pluginFile = pluginShared.Name + ".py"
)

type hrpPlugin interface {
	init(path string) error                                         // init plugin
	has(funcName string) bool                                       // check if plugin has function
	call(funcName string, args ...interface{}) (interface{}, error) // call function
	quit() error                                                    // quit plugin
}

// goPlugin implements golang official plugin
type goPlugin struct {
	*plugin.Plugin
	cachedFunctions map[string]reflect.Value // cache loaded functions to improve performance
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

	p.cachedFunctions = make(map[string]reflect.Value)
	log.Info().Str("path", path).Msg("load go plugin success")
	return nil
}

func (p *goPlugin) has(funcName string) bool {
	fn, ok := p.cachedFunctions[funcName]
	if ok {
		return fn.IsValid()
	}

	sym, err := p.Plugin.Lookup(funcName)
	if err != nil {
		p.cachedFunctions[funcName] = reflect.Value{} // mark as invalid
		return false
	}
	fn = reflect.ValueOf(sym)

	// check function type
	if fn.Kind() != reflect.Func {
		p.cachedFunctions[funcName] = reflect.Value{} // mark as invalid
		return false
	}

	p.cachedFunctions[funcName] = fn
	return true
}

func (p *goPlugin) call(funcName string, args ...interface{}) (interface{}, error) {
	fn := p.cachedFunctions[funcName]
	return pluginShared.CallFunc(fn, args...)
}

func (p *goPlugin) quit() error {
	// no need to quit for go plugin
	return nil
}

// hashicorpPlugin implements hashicorp/go-plugin
type hashicorpPlugin struct {
	pluginShared.FuncCaller
	cachedFunctions map[string]bool // cache loaded functions to improve performance
}

func (p *hashicorpPlugin) init(path string) error {

	f, err := pluginHost.Init(path)
	if err != nil {
		log.Error().Err(err).Str("path", path).Msg("load go hashicorp plugin failed")
		return err
	}
	p.FuncCaller = f

	p.cachedFunctions = make(map[string]bool)
	log.Info().Str("path", path).Msg("load hashicorp go plugin success")
	return nil
}

func (p *hashicorpPlugin) has(funcName string) bool {
	flag, ok := p.cachedFunctions[funcName]
	if ok {
		return flag
	}

	funcNames, err := p.GetNames()
	if err != nil {
		return false
	}

	for _, name := range funcNames {
		if name == funcName {
			p.cachedFunctions[funcName] = true // cache as exists
			return true
		}
	}

	p.cachedFunctions[funcName] = false // cache as not exists
	return false
}

func (p *hashicorpPlugin) call(funcName string, args ...interface{}) (interface{}, error) {
	return p.FuncCaller.Call(funcName, args...)
}

func (p *hashicorpPlugin) quit() error {
	// kill hashicorp plugin process
	pluginHost.Quit()
	return nil
}

func (p *parser) initPlugin(path string) error {
	if path == "" {
		return nil
	}

	// priority: hashicorp plugin > go plugin > builtin functions
	// locate hashicorp plugin file
	pluginPath, err := locatePlugin(path, hashicorpGoPluginFile)
	if err == nil {
		// found hashicorp go plugin file
		p.plugin = &hashicorpPlugin{}
		return p.plugin.init(pluginPath)
	}

	// locate go plugin file
	pluginPath, err = locatePlugin(path, goPluginFile)
	if err == nil {
		// found go plugin file
		p.plugin = &goPlugin{}
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

// callFunc calls function with arguments
// only support return at most one result value
func (p *parser) callFunc(funcName string, arguments ...interface{}) (interface{}, error) {
	// call with plugin function
	if p.plugin != nil && p.plugin.has(funcName) {
		return p.plugin.call(funcName, arguments...)
	}

	// get builtin function
	function, ok := builtin.Functions[funcName]
	if !ok {
		return nil, fmt.Errorf("function %s is not found", funcName)
	}
	fn := reflect.ValueOf(function)

	// call with builtin function
	return pluginShared.CallFunc(fn, arguments...)
}
