# Installation

`HttpRunner+` is developed with Golang, it supports Go `1.13+` and most operating systems. Combination of Go `1.13/1.14/1.15/1.16/1.17` and `macOS/Linux/Windows` are tested continuously on [GitHub-Actions][github-actions].

## install as CLI tool

```bash
$ go get -u github.com/httprunner/hrp/hrp
```

Since installed, you will get a `hrp` command with multiple sub-commands.

```text
$ hrp -h
hrp (HttpRunner+) is the next generation for HttpRunner. Enjoy! ✨ 🚀 ✨

License: Apache-2.0
Github: https://github.com/httprunner/hrp
Copyright 2021 debugtalk

Usage:
  hrp [command]

Available Commands:
  boom        run load test with boomer
  completion  generate the autocompletion script for the specified shell
  har2case    Convert HAR to json/yaml testcase files
  help        Help about any command
  run         run API test

Flags:
  -h, --help               help for hrp
      --log-json           set log to json format
  -l, --log-level string   set log level (default "INFO")
  -v, --version            version for hrp

Use "hrp [command] --help" for more information about a command.
```

## install as library

Beside using `hrp` as a CLI tool, you can also use it as golang library.

```bash
$ go get -u github.com/httprunner/hrp
```

Then you can import `github.com/httprunner/hrp` and write testcases in Golang.

[github-actions]: https://github.com/httprunner/hrp/actions
