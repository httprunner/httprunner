## Installation

`HttpRunner` is available on [`PyPI`][PyPI] and can be installed through pip or easy_install.

```bash
$ pip install HttpRunner
```

or

```bash
$ easy_install HttpRunner
```

If you want to keep up with the latest version, you can install with github repository url.

```bash
$ pip install git+https://github.com/HttpRunner/HttpRunner.git#egg=HttpRunner
```

## Upgrade

If you have installed `HttpRunner` before and want to upgrade to the latest version, you can use the `-U` option.

This option works on each installation method described above.

```bash
$ pip install -U HttpRunner
$ easy_install -U HttpRunner
$ pip install -U git+https://github.com/HttpRunner/HttpRunner.git#egg=HttpRunner
```

## Check Installation

When HttpRunner is installed, a **httprunner** (**hrun** for short) command should be available in your shell (if you're not using
virtualenv—which you should—make sure your python script directory is on your path).

To see `HttpRunner` version:

```bash
$ httprunner -V     # same as: hrun -V
HttpRunner version: 0.8.1b
PyUnitReport version: 0.1.3b
```

To see available options, run:

```bash
$ httprunner -h     # same as: hrun -h
usage: main-debug.py [-h] [-V] [--no-html-report]
                     [--html-report-name HTML_REPORT_NAME]
                     [--html-report-template HTML_REPORT_TEMPLATE]
                     [--log-level LOG_LEVEL] [--log-file LOG_FILE]
                     [--dot-env-path DOT_ENV_PATH] [--failfast]
                     [--startproject STARTPROJECT]
                     [--validate [VALIDATE [VALIDATE ...]]]
                     [--prettify [PRETTIFY [PRETTIFY ...]]]
                     [testcase_paths [testcase_paths ...]]

One-stop solution for HTTP(S) testing.

positional arguments:
  testcase_paths        testcase file path

optional arguments:
  -h, --help            show this help message and exit
  -V, --version         show version
  --no-html-report      do not generate html report.
  --html-report-name HTML_REPORT_NAME
                        specify html report name, only effective when
                        generating html report.
  --html-report-template HTML_REPORT_TEMPLATE
                        specify html report template path.
  --log-level LOG_LEVEL
                        Specify logging level, default is INFO.
  --log-file LOG_FILE   Write logs to specified file path.
  --dot-env-path DOT_ENV_PATH
                        Specify .env file path, which is useful for keeping
                        sensitive data.
  --failfast            Stop the test run on the first error or failure.
  --startproject STARTPROJECT
                        Specify new project name.
  --validate [VALIDATE [VALIDATE ...]]
                        Validate JSON testcase format.
  --prettify [PRETTIFY [PRETTIFY ...]]
                        Prettify JSON testcase format.
```

## Supported Python Versions

HttpRunner supports Python 2.7, 3.4, 3.5, and 3.6. And we strongly recommend you to use `Python 3.6`.

`HttpRunner` has been tested on `macOS`, `Linux` and `Windows` platforms.


[PyPI]: https://pypi.python.org/pypi
