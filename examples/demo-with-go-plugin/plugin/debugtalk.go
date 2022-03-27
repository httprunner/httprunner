package main

import (
	"fmt"

	"github.com/httprunner/funplugin/fungo"
)

func SumTwoInt(a, b int) int {
	return a + b
}

func SumInts(args ...int) int {
	var sum int
	for _, arg := range args {
		sum += arg
	}
	return sum
}

func Sum(args ...interface{}) (interface{}, error) {
	var sum float64
	for _, arg := range args {
		switch v := arg.(type) {
		case int:
			sum += float64(v)
		case float64:
			sum += v
		default:
			return nil, fmt.Errorf("unexpected type: %T", arg)
		}
	}
	return sum, nil
}

func SetupHookExample(args string) string {
	return fmt.Sprintf("step name: %v, setup...", args)
}

func TeardownHookExample(args string) string {
	return fmt.Sprintf("step name: %v, teardown...", args)
}

func GetVersion() string {
	return "v4.0.0-alpha"
}

func main() {
	fungo.Register("get_httprunner_version", GetVersion)
	fungo.Register("sum_ints", SumInts)
	fungo.Register("sum_two_int", SumTwoInt)
	fungo.Register("sum_two", SumTwoInt)
	fungo.Register("sum", Sum)
	fungo.Register("setup_hook_example", SetupHookExample)
	fungo.Register("teardown_hook_example", TeardownHookExample)
	fungo.Serve()
}
