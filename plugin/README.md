# plugin

When you need to do some dynamic calculations or custom logic processing in testcases, you need to use the plugin function mechanism.

HttpRunner+ supports both [hashicorp/plugin] and [go plugin] to create and call custom functions.

## hashicorp/plugin

It is recommended to use [hashicorp/plugin] in most cases.

### create plugin functions

Firstly, you need to define your plugin functions. The functions can be very flexible, only the following restrictions should be complied with.

- function should return at most one value and one error.
- `Register()` and `Serve()` must be called to register plugin functions and start a plugin server process in `main()`.

Here is some plugin functions as example.

```go
package main

import (
	"fmt"

	"github.com/httprunner/hrp/plugin"
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

func main() {
	plugin.Register("sum_ints", SumInts)
	plugin.Register("sum_two_int", SumTwoInt)
	plugin.Register("sum", Sum)
	plugin.Serve()
}
```

You can get more examples at [examples/plugin/]

### build plugin

Secondly, you can build your hashicorp plugin to binary file `debugtalk.bin`. The name of `debugtalk.bin` is by convention and should not be changed.

```bash
$ go build -o examples/debugtalk.bin examples/plugin/hashicorp.go examples/plugin/debugtalk.go
```

It is recommended to place the `debugtalk.bin` file in your project root folder, or you can put it in the parent folder of the target testcase file. HttpRunner+ will search `debugtalk.bin` upward recursively until current working directory or system root dir.

### use plugin functions

Then, you can call your defined plugin function in your `YAML/JSON` testcase at any position.

```json
{
    "name": "get with params",
    "variables": {
        "a": "${sum_two_int(1,6)}",
        "b": "${sum_ints(1,2,3)}",
        "c": "${sum(1, 2.3, 4)}",
    },
    "request": {
        "method": "GET",
        "url": "/get",
        "params": {
            "foo1": "$c",
            "foo2": "${max($a, $b)}"
        },
        "headers": {
            "User-Agent": "HttpRunnerPlus"
        }
    }
}
```

### rpc vs. gRPC

HttpRunner+ has both supported `net/rpc` and `gRPC` in [hashicorp/plugin]. It is recommended to use `gRPC` and this is the default choice.

If you want to run plugin in `net/rpc` mode, you can set an environment variable `HRP_PLUGIN_TYPE=rpc`.

```bash
$ export HRP_PLUGIN_TYPE=rpc
$ hrp run examples/demo.json
$ hrp boom examples/demo.json
```

## go plugin

The golang official plugin is only supported on Linux, FreeBSD, and macOS. And this solution also has many drawbacks.

### create plugin functions

Firstly, you need to define your plugin functions. The functions can be very flexible, only the following restrictions should be complied with.

- plugin package name must be `main`.
- function names must be capitalized.
- function should return at most one value and one error.

Here is some plugin functions as example.

```go
package main

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
```

You can get more examples at [examples/plugin/debugtalk.go]

### build plugin

Then you can build your go plugin with `-buildmode=plugin` flag to binary file `debugtalk.so`. The name of `debugtalk.so` is by convention and should not be changed.

```bash
$ go build -buildmode=plugin -o=examples/debugtalk.so examples/plugin/debugtalk.go
```

It is recommended to place the `debugtalk.so` file in your project root folder, or you can put it in the parent folder of the target testcase file. HttpRunner+ will search `debugtalk.so` upward recursively until current working directory or system root dir.

### use plugin functions

Then, you can call your defined plugin function in your `YAML/JSON` testcase at any position.

```json
{
    "name": "get with params",
    "variables": {
        "a": "${SumTwoInt(1,6)}",
        "b": "${SumInts(1,2,3)}",
        "c": "${Sum(1, 2.3, 4)}",
    },
    "request": {
        "method": "GET",
        "url": "/get",
        "params": {
            "foo1": "$c",
            "foo2": "${max($a, $b)}"
        },
        "headers": {
            "User-Agent": "HttpRunnerPlus"
        }
    }
}
```

Notice: you should use the original function name.

[hashicorp/plugin]: https://github.com/hashicorp/go-plugin
[go plugin]: https://pkg.go.dev/plugin
[examples/plugin/]: ../examples/plugin/
[examples/plugin/debugtalk.go]: ../examples/plugin/debugtalk.go
