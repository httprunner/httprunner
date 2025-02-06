package main

import (
	"fmt"
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

func GetUserAgent() string {
	return "hrp/fungo"
}
