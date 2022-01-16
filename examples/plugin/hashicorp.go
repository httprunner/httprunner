package main

import "github.com/httprunner/hrp/plugin"

// register functions and build to plugin binary
func main() {
	plugin.Register("sum_ints", SumInts)
	plugin.Register("sum_two_int", SumTwoInt)
	plugin.Register("sum", Sum)
	plugin.Register("sum_two_string", SumTwoString)
	plugin.Register("sum_strings", SumStrings)
	plugin.Register("concatenate", Concatenate)
	plugin.Serve()
}
