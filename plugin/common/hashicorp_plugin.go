package common

import (
	pluginHost "github.com/httprunner/hrp/plugin/host"
	pluginShared "github.com/httprunner/hrp/plugin/shared"
	"github.com/rs/zerolog/log"
)

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
	log.Info().Msg("quit hashicorp plugin process")
	pluginHost.Quit()
	return nil
}
