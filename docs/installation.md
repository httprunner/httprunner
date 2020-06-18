
`HttpRunner` is developed with Python, it supports Python `3.6+` and most operating systems. Combination of Python `3.6/3.7/3.8` and `macOS/Linux/Windows` are tested continuously on [GitHub-Actions][github-actions].

## Installation

`HttpRunner` is available on [`PyPI`][PyPI] and can be installed through `pip`.

```bash
$ pip3 install httprunner
```

If you want to keep up with the latest version, you can install with github repository url.

```bash
$ pip3 install git+https://github.com/httprunner/httprunner.git@master
```

If you have installed `HttpRunner` before and want to upgrade to the latest version, you can use the `-U` option.

```bash
$ pip3 install -U httprunner
$ pip3 install -U git+https://github.com/httprunner/httprunner.git@master
```

## Check Installation

When HttpRunner is installed, 5 commands will be added in your system.

- `httprunner`: main command, used for all functions
- `hrun`: alias for `httprunner run`, used to run YAML/JSON/pytest testcases
- `hmake`: alias for `httprunner make`, used to convert YAML/JSON testcases to pytest files
- `har2case`: alias for `httprunner har2case`, used to convert HAR to YAML/JSON testcases
- `locusts`: used to run load test with [locust][locust]

To see `HttpRunner` version:

```text
$ httprunner -V  # hrun -V
3.1.0
```

To see available options, run:

```text
$ httprunner -h
usage: httprunner [-h] [-V] {run,startproject,har2case,make} ...

One-stop solution for HTTP(S) testing.

positional arguments:
  {run,startproject,har2case,make}
                        sub-command help
    run                 Make HttpRunner testcases and run with pytest.
    startproject        Create a new project with template structure.
    har2case            Convert HAR(HTTP Archive) to YAML/JSON testcases for
                        HttpRunner.
    make                Convert YAML/JSON testcases to pytest cases.

optional arguments:
  -h, --help            show this help message and exit
  -V, --version         show version
```

> Notice: `locusts` is an individual command, for the reason to monkey patch ssl at beginning to avoid RecursionError when running locust.

[PyPI]: https://pypi.python.org/pypi
[github-actions]: https://github.com/httprunner/httprunner/actions
[locust]: http://locust.io/
