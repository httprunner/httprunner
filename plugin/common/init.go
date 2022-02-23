package common

import (
	"fmt"
	"os"
	"path/filepath"

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

func Init(path string) (Plugin, error) {
	if path == "" {
		return nil, nil
	}
	var plugin Plugin

	// priority: hashicorp plugin > go plugin
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
