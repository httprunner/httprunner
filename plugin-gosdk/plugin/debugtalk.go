package main

import (
	"fmt"

	plugin "github.com/httprunner/hrp/plugin-gosdk"
)

func SumInt(args ...interface{}) (interface{}, error) {
	var sum int
	for _, arg := range args {
		v, ok := arg.(int)
		if !ok {
			return nil, fmt.Errorf("unexpected type: %T", arg)
		}
		sum += v
	}
	return sum, nil
}

func ConcatenateString(args ...interface{}) (interface{}, error) {
	var result string
	for _, arg := range args {
		v, ok := arg.(string)
		if !ok {
			return nil, fmt.Errorf("unexpected type: %T", arg)
		}
		result += v
	}
	return result, nil
}

// register functions and build to plugin binary
func main() {
	plugin.Register("sum_int", SumInt)
	plugin.Register("concatenate_string", ConcatenateString)
	plugin.Serve()
}
