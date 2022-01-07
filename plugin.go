package hrp

import (
	"fmt"
	"plugin"
	"reflect"
	"runtime"

	"github.com/rs/zerolog/log"

	"github.com/httprunner/hrp/internal/builtin"
	"github.com/httprunner/hrp/internal/ga"
)

func (p *parser) loadPlugin(path string) error {
	if runtime.GOOS == "windows" {
		log.Warn().Msg("go plugin does not support windows")
		return nil
	}

	if path == "" {
		return nil
	}

	// check if loaded before
	if p.pluginLoader != nil {
		return nil
	}

	// locate plugin file
	pluginPath, err := locatePlugin(path)
	if err != nil {
		// plugin not found
		return nil
	}

	// report event for loading go plugin
	go ga.SendEvent(ga.EventTracking{
		Category: "LoadGoPlugin",
		Action:   "plugin.Open",
	})

	// load plugin
	plugins, err := plugin.Open(pluginPath)
	if err != nil {
		log.Error().Err(err).Str("path", path).Msg("load go plugin failed")
		return err
	}
	p.pluginLoader = plugins

	log.Info().Str("path", path).Msg("load go plugin success")
	return nil
}

func getMappingFunction(funcName string, pluginLoader *plugin.Plugin) (reflect.Value, error) {
	var fn reflect.Value
	var err error

	defer func() {
		// check function type
		if err == nil && fn.Kind() != reflect.Func {
			// function not valid
			err = fmt.Errorf("function %s is invalid", funcName)
			return
		}
	}()

	// get function from plugin loader
	if pluginLoader != nil {
		sym, err := pluginLoader.Lookup(funcName)
		if err == nil {
			fn = reflect.ValueOf(sym)
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
	fn, err := getMappingFunction(funcName, p.pluginLoader)
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
