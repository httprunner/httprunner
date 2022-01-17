package common

import (
	"fmt"
	"os"
	"path/filepath"
	"plugin"
	"reflect"
	"runtime"

	"github.com/rs/zerolog/log"

	pluginHost "github.com/httprunner/hrp/plugin/host"
	pluginShared "github.com/httprunner/hrp/plugin/shared"
)

type pluginFile string

const (
	goPluginFile          pluginFile = pluginShared.Name + ".so"  // built from go plugin
	hashicorpGoPluginFile pluginFile = pluginShared.Name + ".bin" // built from hashicorp go plugin
	hashicorpPyPluginFile pluginFile = pluginShared.Name + ".py"
)

type Plugin interface {
	Init(path string) error                                         // init plugin
	Has(funcName string) bool                                       // check if plugin has function
	Call(funcName string, args ...interface{}) (interface{}, error) // call function
	Quit() error                                                    // quit plugin
}

// GoPlugin implements golang official plugin
type GoPlugin struct {
	*plugin.Plugin
	cachedFunctions map[string]reflect.Value // cache loaded functions to improve performance
}

func (p *GoPlugin) Init(path string) error {
	if runtime.GOOS == "windows" {
		log.Warn().Msg("go plugin does not support windows")
		return fmt.Errorf("go plugin does not support windows")
	}

	var err error
	p.Plugin, err = plugin.Open(path)
	if err != nil {
		log.Error().Err(err).Str("path", path).Msg("load go plugin failed")
		return err
	}

	p.cachedFunctions = make(map[string]reflect.Value)
	log.Info().Str("path", path).Msg("load go plugin success")
	return nil
}

func (p *GoPlugin) Has(funcName string) bool {
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

func (p *GoPlugin) Call(funcName string, args ...interface{}) (interface{}, error) {
	if !p.Has(funcName) {
		return nil, fmt.Errorf("function %s not found", funcName)
	}
	fn := p.cachedFunctions[funcName]
	return CallFunc(fn, args...)
}

func (p *GoPlugin) Quit() error {
	// no need to quit for go plugin
	return nil
}

// HashicorpPlugin implements hashicorp/go-plugin
type HashicorpPlugin struct {
	pluginShared.FuncCaller
	cachedFunctions map[string]bool // cache loaded functions to improve performance
}

func (p *HashicorpPlugin) Init(path string) error {

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

func (p *HashicorpPlugin) Has(funcName string) bool {
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

func (p *HashicorpPlugin) Call(funcName string, args ...interface{}) (interface{}, error) {
	return p.FuncCaller.Call(funcName, args...)
}

func (p *HashicorpPlugin) Quit() error {
	// kill hashicorp plugin process
	pluginHost.Quit()
	return nil
}

func Init(path string) (Plugin, error) {
	if path == "" {
		return nil, nil
	}
	var plugin Plugin

	// priority: hashicorp plugin > go plugin > builtin functions
	// locate hashicorp plugin file
	pluginPath, err := locateFile(path, hashicorpGoPluginFile)
	if err == nil {
		// found hashicorp go plugin file
		plugin = &HashicorpPlugin{}
		err = plugin.Init(pluginPath)
		return plugin, err
	}

	// locate go plugin file
	pluginPath, err = locateFile(path, goPluginFile)
	if err == nil {
		// found go plugin file
		plugin = &GoPlugin{}
		err = plugin.Init(pluginPath)
		return plugin, err
	}

	// plugin not found
	return nil, nil
}

// locateFile searches destFile upward recursively until current
// working directory or system root dir.
func locateFile(startPath string, destFile pluginFile) (string, error) {
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
	pluginPath := filepath.Join(startDir, string(destFile))
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

	return locateFile(parentDir, destFile)
}
