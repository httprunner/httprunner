package hrp

import (
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"

	"github.com/httprunner/funplugin"
	"github.com/rs/zerolog/log"

	"github.com/httprunner/httprunner/v4/hrp/internal/sdk"
)

const (
	PluginGoBuiltFile          = "debugtalk.so"      // built from go official plugin
	PluginHashicorpGoBuiltFile = "debugtalk.bin"     // built from hashicorp go plugin
	PluginGoSourceFile         = "debugtalk.go"      // golang function plugin source file
	PluginGoSourceGenFile      = "debugtalk_gen.go"  // generated for hashicorp go plugin
	PluginPySourceFile         = "debugtalk.py"      // python function plugin source file
	PluginPySourceGenFile      = ".debugtalk_gen.py" // generated for hashicorp python plugin
)

const projectInfoFile = "proj.json" // used for ensuring root project

func initPlugin(path string, logOn bool) (plugin funplugin.IPlugin, err error) {
	// plugin file not found
	if path == "" {
		return nil, nil
	}
	pluginPath, err := locatePlugin(path)
	if err != nil {
		return nil, nil
	}

	if strings.HasSuffix(pluginPath, ".py") {
		// register funppy plugin
		genPyPluginPath := filepath.Join(filepath.Dir(pluginPath), PluginPySourceGenFile)
		err = BuildPlugin(pluginPath, genPyPluginPath)
		if err != nil {
			log.Error().Err(err).Str("path", pluginPath).Msg("build plugin failed")
			return nil, nil
		}
		pluginPath = genPyPluginPath
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
	event := sdk.EventTracking{
		Category: "InitPlugin",
		Action:   fmt.Sprintf("Init %s plugin", plugin.Type()),
		Value:    0, // success
	}
	if err != nil {
		event.Value = 1 // failed
	}
	go sdk.SendEvent(event)

	return
}

func locatePlugin(path string) (pluginPath string, err error) {
	// priority: hashicorp plugin (debugtalk.bin > debugtalk.py) > go plugin (debugtalk.so)

	pluginPath, err = locateFile(path, PluginHashicorpGoBuiltFile)
	if err == nil {
		return
	}

	pluginPath, err = locateFile(path, PluginPySourceFile)
	if err == nil {
		return
	}

	pluginPath, err = locateFile(path, PluginGoBuiltFile)
	if err == nil {
		return
	}

	log.Warn().Err(err).Str("path", path).Msg("plugin file not found")
	return "", fmt.Errorf("plugin file not found")
}

// locateFile searches destFile upward recursively until system root dir
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

	// system root dir
	parentDir, _ := filepath.Abs(filepath.Dir(startDir))
	if parentDir == startDir {
		return "", fmt.Errorf("searched to system root dir, plugin file not found")
	}

	return locateFile(parentDir, destFile)
}

func GetProjectRootDirPath(path string) (rootDir string, err error) {
	pluginPath, err := locatePlugin(path)
	if err == nil {
		rootDir = filepath.Dir(pluginPath)
		return
	}
	// fix: no debugtalk file in project but having proj.json created by startpeoject
	projPath, err := locateFile(path, projectInfoFile)
	if err == nil {
		rootDir = filepath.Dir(projPath)
		return
	}

	// failed to locate project root dir
	// maybe project plugin debugtalk.xx and proj.json are not exist
	// use current dir instead
	return os.Getwd()
}
