package main

import (
	"fmt"
	"log"
)

func init() {
	log.Println("plugin init function called")
}

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

func SumTwoString(a, b string) string {
	return a + b
}

func SumStrings(s ...string) string {
	var sum string
	for _, arg := range s {
		sum += arg
	}
	return sum
}

func Concatenate(args ...interface{}) (interface{}, error) {
	var result string
	for _, arg := range args {
		result += fmt.Sprintf("%v", arg)
	}
	return result, nil
}
