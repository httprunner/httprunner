package pluginInternal

// FuncCaller is the interface that we're exposing as a plugin.
type FuncCaller interface {
	GetNames() ([]string, error)                                    // get all plugin function names list
	Call(funcName string, args ...interface{}) (interface{}, error) // call plugin function
}

type IPlugin interface {
	Init(path string) error                                         // init plugin
	Has(funcName string) bool                                       // check if plugin has function
	Call(funcName string, args ...interface{}) (interface{}, error) // call function
	Quit() error                                                    // quit plugin
}
