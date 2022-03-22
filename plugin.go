package hrp

import (
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	"github.com/httprunner/funplugin"
	"github.com/httprunner/hrp/internal/ga"
	"github.com/rs/zerolog/log"
)

const (
	goPluginFile          = "debugtalk.so"  // built from go plugin
	hashicorpGoPluginFile = "debugtalk.bin" // built from hashicorp go plugin
	hashicorpPyPluginFile = "debugtalk.py"  // used for hashicorp python plugin
)

func initPlugin(path string, logOn bool) (plugin funplugin.IPlugin, err error) {
	// plugin file not found
	if path == "" {
		return nil, nil
	}
	pluginPath, err := locatePlugin(path)
	if err != nil {
		return nil, nil
	}

	// found plugin file
	plugin, err = funplugin.Init(pluginPath, funplugin.WithLogOn(logOn))
	if err != nil {
		log.Error().Err(err).Msgf("init plugin failed: %s", pluginPath)
		return
	}

	// catch Interrupt and SIGTERM signals to ensure plugin quitted
	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		plugin.Quit()
	}()

	// report event for initializing plugin
	event := ga.EventTracking{
		Category: "InitPlugin",
		Action:   fmt.Sprintf("Init %s plugin", plugin.Type()),
		Value:    0, // success
	}
	if err != nil {
		event.Value = 1 // failed
	}
	go ga.SendEvent(event)

	return
}

func locatePlugin(path string) (pluginPath string, err error) {
	// priority: hashicorp plugin (debugtalk.bin > debugtalk.py) > go plugin (debugtalk.so)

	pluginPath, err = locateFile(path, hashicorpGoPluginFile)
	if err == nil {
		return
	}

	pluginPath, err = locateFile(path, hashicorpPyPluginFile)
	if err == nil {
		return
	}

	pluginPath, err = locateFile(path, goPluginFile)
	if err == nil {
		return
	}

	return "", fmt.Errorf("plugin file not found")
}

// locateFile searches destFile upward recursively until current
// working directory or system root dir.
func locateFile(startPath string, destFile string) (string, error) {
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
	pluginPath := filepath.Join(startDir, destFile)
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
