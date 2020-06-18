# Run Testcase

Once testcase is ready, you can run testcase with `hrun` command.

Notice, `hrun` is an command alias of `httprunner run`, they have the same effect.

```text
hrun = httprunner run
```

## run testcases in diverse ways

`HttpRunner` can run testcases in diverse ways.

You can run single testcase by specifying testcase file path.

```text
$ hrun path/to/testcase1
```

You can also run several testcases by specifying multiple testcase file paths.

```text
$ hrun path/to/testcase1 path/to/testcase2
```

If you want to run testcases of a whole project, you can achieve this goal by specifying the project folder path.

```text
$ hrun path/to/testcase_folder/
```

## run YAML/JSON testcases

If your testcases are written in YAML/JSON format, `hrun` will firstly convert YAML/JSON testcases to pytest(python) files, and run with `pytest` command.

That is to say,

```text
hrun = make + pytest
```

In most cases, the generated pytest files are in the same folder next to origin YAML/JSON files, with the same file name except adding `_test` suffix and replace extension `.yml/yaml/.json` with `.py`.

```text
/path/to/example.yml => /path/to/example_test.py
```

However, if the testcase folder name or file name contains symbols like dot, hyphen or space, these symbols will be replaced with underscore in order to avoid syntax error in python class importing (testcase reference). Also, folder/file name starts with digit will be adding a prefix `T` because python module and class name can not be started with digit.

```text
path 1/a.b-2/3.yml => path_1/a_b_2/T3_test.py
```

## run pytest testcases

If your testcases are written in pytest format, or you want to run pytest files converted from YAML/JSON testcases, `hrun` and `pytest` commands are both okay. What you need to remember is that `hrun` only wraps `pytest`, thus all the arguments of `pytest` can be used with `hrun`.

```text
$ hrun -h
usage: hrun [options] [file_or_dir] [file_or_dir] [...]

positional arguments:
  file_or_dir

general:
  -k EXPRESSION         only run tests which match the given substring expression. An expression is a python evaluatable expression where all names are
                        substring-matched against test names and their parent classes. Example: -k 'test_method or test_other' matches all test functions and
                        classes whose name contains 'test_method' or 'test_other', while -k 'not test_method' matches those that don't contain 'test_method' in
                        their names. -k 'not test_method and not test_other' will eliminate the matches. Additionally keywords are matched to classes and
                        functions containing extra names in their 'extra_keyword_matches' set, as well as functions which have names assigned directly to them.
                        The matching is case-insensitive.
  -m MARKEXPR           only run tests matching given mark expression. example: -m 'mark1 and not mark2'.
  --markers             show markers (builtin, plugin and per-project ones).
  -x, --exitfirst       exit instantly on first error or failed test.
  --maxfail=num         exit after first num failures or errors.
  --strict-markers, --strict
                        markers not registered in the `markers` section of the configuration file raise errors.
  -c file               load configuration from `file` instead of trying to locate one of the implicit configuration files.
  --continue-on-collection-errors
                        Force test execution even if collection errors occur.
  --rootdir=ROOTDIR     Define root directory for tests. Can be relative path: 'root_dir', './root_dir', 'root_dir/another_dir/'; absolute path:
                        '/home/user/root_dir'; path with variables: '$HOME/root_dir'.
  --fixtures, --funcargs
                        show available fixtures, sorted by plugin appearance (fixtures with leading '_' are only shown with '-v')
  --fixtures-per-test   show fixtures per test
  --import-mode={prepend,append}
                        prepend/append to sys.path when importing test modules, default is to prepend.
  --pdb                 start the interactive Python debugger on errors or KeyboardInterrupt.
  --pdbcls=modulename:classname
                        start a custom interactive Python debugger on errors. For example: --pdbcls=IPython.terminal.debugger:TerminalPdb
  --trace               Immediately break when running each test.
  --capture=method      per-test capturing method: one of fd|sys|no|tee-sys.
  -s                    shortcut for --capture=no.
  --runxfail            report the results of xfail tests as if they were not marked
  --lf, --last-failed   rerun only the tests that failed at the last run (or all if none failed)
  --ff, --failed-first  run all tests but run the last failures first. This may re-order tests and thus lead to repeated fixture setup/teardown
  --nf, --new-first     run tests from new files first, then the rest of the tests sorted by file mtime
  --cache-show=[CACHESHOW]
                        show cache contents, don't perform collection or tests. Optional argument: glob (default: '*').
  --cache-clear         remove all cache contents at start of test run.
  --lfnf={all,none}, --last-failed-no-failures={all,none}
                        which tests to run with no previously (known) failures.
  --sw, --stepwise      exit on test failure and continue from last failing test next time
  --stepwise-skip       ignore the first failing test but stop on the next failing test
  --allure-severities=SEVERITIES_SET
                        Comma-separated list of severity names. Tests only with these severities will be run. Possible values are: blocker, critical, normal,
                        minor, trivial.
  --allure-epics=EPICS_SET
                        Comma-separated list of epic names. Run tests that have at least one of the specified feature labels.
  --allure-features=FEATURES_SET
                        Comma-separated list of feature names. Run tests that have at least one of the specified feature labels.
  --allure-stories=STORIES_SET
                        Comma-separated list of story names. Run tests that have at least one of the specified story labels.
  --allure-link-pattern=LINK_TYPE:LINK_PATTERN
                        Url pattern for link type. Allows short links in test, like 'issue-1'. Text will be formatted to full url with python str.format().

reporting:
  --durations=N         show N slowest setup/test durations (N=0 for all).
  -v, --verbose         increase verbosity.
  -q, --quiet           decrease verbosity.
  --verbosity=VERBOSE   set verbosity. Default is 0.
  -r chars              show extra test summary info as specified by chars: (f)ailed, (E)rror, (s)kipped, (x)failed, (X)passed, (p)assed, (P)assed with output,
                        (a)ll except passed (p/P), or (A)ll. (w)arnings are enabled by default (see --disable-warnings), 'N' can be used to reset the list.
                        (default: 'fE').
  --disable-warnings, --disable-pytest-warnings
                        disable warnings summary
  -l, --showlocals      show locals in tracebacks (disabled by default).
  --tb=style            traceback print mode (auto/long/short/line/native/no).
  --show-capture={no,stdout,stderr,log,all}
                        Controls how captured stdout/stderr/log is shown on failed tests. Default is 'all'.
  --full-trace          don't cut any tracebacks (default is to cut).
  --color=color         color terminal output (yes/no/auto).
  --pastebin=mode       send failed|all info to bpaste.net pastebin service.
  --junit-xml=path      create junit-xml style report file at given path.
  --junit-prefix=str    prepend prefix to classnames in junit-xml output
  --result-log=path     DEPRECATED path for machine-readable result log.
  --html=path           create html report file at given path.
  --self-contained-html
                        create a self-contained html file containing all necessary styles, scripts, and images - this means that the report may not render or
                        function where CSP restrictions are in place (see https://developer.mozilla.org/docs/Web/Security/CSP)
  --css=path            append given css file content to report style file.

collection:
  --collect-only, --co  only collect tests, don't execute them.
  --pyargs              try to interpret all arguments as python packages.
  --ignore=path         ignore path during collection (multi-allowed).
  --ignore-glob=path    ignore path pattern during collection (multi-allowed).
  --deselect=nodeid_prefix
                        deselect item (via node id prefix) during collection (multi-allowed).
  --confcutdir=dir      only load conftest.py's relative to specified dir.
  --noconftest          Don't load any conftest.py files.
  --keep-duplicates     Keep duplicate tests.
  --collect-in-virtualenv
                        Don't ignore tests in a local virtualenv directory
  --doctest-modules     run doctests in all .py modules
  --doctest-report={none,cdiff,ndiff,udiff,only_first_failure}
                        choose another output format for diffs on doctest failure
  --doctest-glob=pat    doctests file matching pattern, default: test*.txt
  --doctest-ignore-import-errors
                        ignore doctest ImportErrors
  --doctest-continue-on-failure
                        for a given doctest, continue to run after the first failure

test session debugging and configuration:
  --basetemp=dir        base temporary directory for this test run.(warning: this directory is removed if it exists)
  -V, --version         display pytest version and information about plugins.
  -h, --help            show help message and configuration info
  -p name               early-load given plugin module name or entry point (multi-allowed). To avoid loading of plugins, use the `no:` prefix, e.g. `no:doctest`.
  --trace-config        trace considerations of conftest.py files.
  --debug               store internal tracing debug information in 'pytestdebug.log'.
  -o OVERRIDE_INI, --override-ini=OVERRIDE_INI
                        override ini option with "option=value" style, e.g. `-o xfail_strict=True -o cache_dir=cache`.
  --assert=MODE         Control assertion debugging tools. 'plain' performs no assertion debugging. 'rewrite' (the default) rewrites assert statements in test
                        modules on import to provide assert expression information.
  --setup-only          only setup fixtures, do not execute tests.
  --setup-show          show setup of fixtures while executing tests.
  --setup-plan          show what fixtures and tests would be executed but don't execute anything.

pytest-warnings:
  -W PYTHONWARNINGS, --pythonwarnings=PYTHONWARNINGS
                        set which warnings to report, see -W option of python itself.

logging:
  --no-print-logs       disable printing caught logs on failed tests.
  --log-level=LEVEL     level of messages to catch/display. Not set by default, so it depends on the root/parent log handler's effective level, where it is
                        "WARNING" by default.
  --log-format=LOG_FORMAT
                        log format as used by the logging module.
  --log-date-format=LOG_DATE_FORMAT
                        log date format as used by the logging module.
  --log-cli-level=LOG_CLI_LEVEL
                        cli logging level.
  --log-cli-format=LOG_CLI_FORMAT
                        log format as used by the logging module.
  --log-cli-date-format=LOG_CLI_DATE_FORMAT
                        log date format as used by the logging module.
  --log-file=LOG_FILE   path to a file when logging will be written to.
  --log-file-level=LOG_FILE_LEVEL
                        log file logging level.
  --log-file-format=LOG_FILE_FORMAT
                        log format as used by the logging module.
  --log-file-date-format=LOG_FILE_DATE_FORMAT
                        log date format as used by the logging module.
  --log-auto-indent=LOG_AUTO_INDENT
                        Auto-indent multiline messages passed to the logging module. Accepts true|on, false|off or an integer.

reporting:
  --alluredir=DIR       Generate Allure report in the specified directory (may not exist)
  --clean-alluredir     Clean alluredir folder if it exists
  --allure-no-capture   Do not attach pytest captured logging/stdout/stderr to report

custom options:
  --metadata=key value  additional metadata.
  --metadata-from-json=METADATA_FROM_JSON
                        additional metadata from a json string.

[pytest] ini-options in the first pytest.ini|tox.ini|setup.cfg file found:

  markers (linelist):   markers for test functions
  empty_parameter_set_mark (string):
                        default marker for empty parametersets
  norecursedirs (args): directory patterns to avoid for recursion
  testpaths (args):     directories to search for tests when no files or directories are given in the command line.
  usefixtures (args):   list of default fixtures to be used with this project
  python_files (args):  glob-style file patterns for Python test module discovery
  python_classes (args):
                        prefixes or glob names for Python test class discovery
  python_functions (args):
                        prefixes or glob names for Python test function and method discovery
  disable_test_id_escaping_and_forfeit_all_rights_to_community_support (bool):
                        disable string escape non-ascii characters, might cause unwanted side effects(use at your own risk)
  console_output_style (string):
                        console output: "classic", or with additional progress information ("progress" (percentage) | "count").
  xfail_strict (bool):  default for the strict parameter of xfail markers when not given explicitly (default: False)
  enable_assertion_pass_hook (bool):
                        Enables the pytest_assertion_pass hook.Make sure to delete any previously generated pyc cache files.
  junit_suite_name (string):
                        Test suite name for JUnit report
  junit_logging (string):
                        Write captured log messages to JUnit report: one of no|log|system-out|system-err|out-err|all
  junit_log_passing_tests (bool):
                        Capture log information for passing tests to JUnit report:
  junit_duration_report (string):
                        Duration time to report: one of total|call
  junit_family (string):
                        Emit XML for schema: one of legacy|xunit1|xunit2
  doctest_optionflags (args):
                        option flags for doctests
  doctest_encoding (string):
                        encoding used for doctest files
  cache_dir (string):   cache directory path.
  filterwarnings (linelist):
                        Each line specifies a pattern for warnings.filterwarnings. Processed after -W/--pythonwarnings.
  log_print (bool):     default value for --no-print-logs
  log_level (string):   default value for --log-level
  log_format (string):  default value for --log-format
  log_date_format (string):
                        default value for --log-date-format
  log_cli (bool):       enable log display during test run (also known as "live logging").
  log_cli_level (string):
                        default value for --log-cli-level
  log_cli_format (string):
                        default value for --log-cli-format
  log_cli_date_format (string):
                        default value for --log-cli-date-format
  log_file (string):    default value for --log-file
  log_file_level (string):
                        default value for --log-file-level
  log_file_format (string):
                        default value for --log-file-format
  log_file_date_format (string):
                        default value for --log-file-date-format
  log_auto_indent (string):
                        default value for --log-auto-indent
  faulthandler_timeout (string):
                        Dump the traceback of all threads if a test takes more than TIMEOUT seconds to finish. Not available on Windows.
  addopts (args):       extra command line options
  minversion (string):  minimally required pytest version
  render_collapsed (bool):
                        Open the report with all rows collapsed. Useful for very large reports

environment variables:
  PYTEST_ADDOPTS           extra command line options
  PYTEST_PLUGINS           comma-separated plugins to load during startup
  PYTEST_DISABLE_PLUGIN_AUTOLOAD set to disable plugin auto-loading
  PYTEST_DEBUG             set to enable debug tracing of pytest's internals


to see available markers type: pytest --markers
to see available fixtures type: pytest --fixtures
(shown according to specified file_or_dir or current dir if not specified; fixtures with leading '_' are only shown with the '-v' option
```

## execution logs

By default, `hrun` will not print details of request and response data.

```text
$ hrun examples/postman_echo/request_methods/request_with_functions.yml 
2020-06-17 15:39:41.041 | INFO     | httprunner.make:make_testcase:317 - start to make testcase: /Users/debugtalk/MyProjects/HttpRunner-dev/HttpRunner/examples/postman_echo/request_methods/request_with_functions.yml
2020-06-17 15:39:41.042 | INFO     | httprunner.make:make_testcase:390 - generated testcase: /Users/debugtalk/MyProjects/HttpRunner-dev/HttpRunner/examples/postman_echo/request_methods/request_with_functions_test.py
2020-06-17 15:39:41.042 | INFO     | httprunner.make:format_pytest_with_black:154 - format pytest cases with black ...
reformatted /Users/debugtalk/MyProjects/HttpRunner-dev/HttpRunner/examples/postman_echo/request_methods/request_with_functions_test.py
All done! ✨ 🍰 ✨
1 file reformatted, 1 file left unchanged.
2020-06-17 15:39:41.315 | INFO     | httprunner.cli:main_run:56 - start to run tests with pytest. HttpRunner version: 3.0.13
====================================================================== test session starts ======================================================================
platform darwin -- Python 3.7.5, pytest-5.4.2, py-1.8.1, pluggy-0.13.1
rootdir: /Users/debugtalk/MyProjects/HttpRunner-dev/HttpRunner
plugins: metadata-1.9.0, allure-pytest-2.8.16, html-2.1.1
collected 1 item                                                                                                                                                

examples/postman_echo/request_methods/request_with_functions_test.py .                                                                                    [100%]

======================================================================= 1 passed in 2.98s =======================================================================
```

If you want to view details of request & response data, extraction and validation, you can add an argument `-s` (shortcut for `--capture=no`). 

```text
$ hrun -s examples/postman_echo/request_methods/request_with_functions.yml
2020-06-17 15:42:54.369 | INFO     | httprunner.make:make_testcase:317 - start to make testcase: /Users/debugtalk/MyProjects/HttpRunner-dev/HttpRunner/examples/postman_echo/request_methods/request_with_functions.yml
2020-06-17 15:42:54.369 | INFO     | httprunner.make:make_testcase:390 - generated testcase: /Users/debugtalk/MyProjects/HttpRunner-dev/HttpRunner/examples/postman_echo/request_methods/request_with_functions_test.py
2020-06-17 15:42:54.370 | INFO     | httprunner.make:format_pytest_with_black:154 - format pytest cases with black ...
reformatted /Users/debugtalk/MyProjects/HttpRunner-dev/HttpRunner/examples/postman_echo/request_methods/request_with_functions_test.py
All done! ✨ 🍰 ✨
1 file reformatted, 1 file left unchanged.
2020-06-17 15:42:54.699 | INFO     | httprunner.cli:main_run:56 - start to run tests with pytest. HttpRunner version: 3.0.13
====================================================================== test session starts ======================================================================
platform darwin -- Python 3.7.5, pytest-5.4.2, py-1.8.1, pluggy-0.13.1
rootdir: /Users/debugtalk/MyProjects/HttpRunner-dev/HttpRunner
plugins: metadata-1.9.0, allure-pytest-2.8.16, html-2.1.1
collected 1 item                                                                                                                                                

examples/postman_echo/request_methods/request_with_functions_test.py 2020-06-17 15:42:55.017 | INFO     | httprunner.runner:test_start:435 - Start to run testcase: request methods testcase with functions, TestCase ID: cc404c49-000f-485c-b4c1-ac3367a053fe
2020-06-17 15:42:55.018 | INFO     | httprunner.runner:__run_step:278 - run step begin: get with params >>>>>>
2020-06-17 15:42:56.326 | DEBUG    | httprunner.client:log_print:40 - 
================== request details ==================
method   : GET
url      : https://postman-echo.com/get?foo1=bar11&foo2=bar21&sum_v=3
headers  : {
    "User-Agent": "HttpRunner/3.0.13",
    "Accept-Encoding": "gzip, deflate",
    "Accept": "*/*",
    "Connection": "keep-alive",
    "HRUN-Request-ID": "HRUN-cc404c49-000f-485c-b4c1-ac3367a053fe-775018",
    "Content-Length": "2",
    "Content-Type": "application/json"
}
cookies  : {}
body     : {}

2020-06-17 15:42:56.327 | DEBUG    | httprunner.client:log_print:40 - 
================== response details ==================
status_code : 200
headers  : {
    "Date": "Wed, 17 Jun 2020 07:42:56 GMT",
    "Content-Type": "application/json; charset=utf-8",
    "Content-Length": "477",
    "Connection": "keep-alive",
    "ETag": "W/\"1dd-2JtBYPcnh8D6fqLz8KFn16Oq1R0\"",
    "Vary": "Accept-Encoding",
    "set-cookie": "sails.sid=s%3A6J_EtUk3nkL_C2xtx-NtAXrlA5wPxEgk.gIO2yBbtvGWIIgQ%2F2mZhMkU669G3F60cvLAPWbwyoGM; Path=/; HttpOnly"
}
cookies  : {
    "sails.sid": "s%3A6J_EtUk3nkL_C2xtx-NtAXrlA5wPxEgk.gIO2yBbtvGWIIgQ%2F2mZhMkU669G3F60cvLAPWbwyoGM"
}
encoding : utf-8
content_type : application/json; charset=utf-8
body     : {
    "args": {
        "foo1": "bar11",
        "foo2": "bar21",
        "sum_v": "3"
    },
    "headers": {
        "x-forwarded-proto": "https",
        "x-forwarded-port": "443",
        "host": "postman-echo.com",
        "x-amzn-trace-id": "Root=1-5ee9c980-d8e98cc72a26ef24f5819ce3",
        "content-length": "2",
        "user-agent": "HttpRunner/3.0.13",
        "accept-encoding": "gzip, deflate",
        "accept": "*/*",
        "hrun-request-id": "HRUN-cc404c49-000f-485c-b4c1-ac3367a053fe-775018",
        "content-type": "application/json"
    },
    "url": "https://postman-echo.com/get?foo1=bar11&foo2=bar21&sum_v=3"
}

2020-06-17 15:42:56.328 | INFO     | httprunner.client:request:203 - status_code: 200, response_time(ms): 1307.33 ms, response_length: 477 bytes
2020-06-17 15:42:56.328 | INFO     | httprunner.response:extract:152 - extract mapping: {'foo3': 'bar21'}
2020-06-17 15:42:56.328 | INFO     | httprunner.response:validate:209 - assert status_code equal 200(int)       ==> pass
2020-06-17 15:42:56.329 | INFO     | httprunner.response:validate:209 - assert body.args.foo1 equal bar11(str)  ==> pass
2020-06-17 15:42:56.329 | INFO     | httprunner.response:validate:209 - assert body.args.sum_v equal 3(str)     ==> pass
2020-06-17 15:42:56.329 | INFO     | httprunner.response:validate:209 - assert body.args.foo2 equal bar21(str)  ==> pass
2020-06-17 15:42:56.330 | INFO     | httprunner.runner:__run_step:290 - run step end: get with params <<<<<<

<Omit>

2020-06-17 15:42:57.019 | INFO     | httprunner.runner:test_start:444 - generate testcase log: /Users/debugtalk/MyProjects/HttpRunner-dev/HttpRunner/examples/postman_echo/logs/cc404c49-000f-485c-b4c1-ac3367a053fe.run.log
.

======================================================================= 1 passed in 2.13s =======================================================================
```

Also, an execution log file will be generated for each testcase, located in `<ProjectRootDir>/logs/TestCaseID.run.log`.

## TestCase ID & Request ID

For the sake of troubleshooting, each testcase will generate a unique ID (uuid4), and each request headers will be added a `HRUN-Request-ID` field with testcase ID automatically.

```text
HRUN-Request-ID = "HRUN-<TestCase ID>-<timestamp_six_digits>"
timestamp_six_digits = str(int(time.time() * 1000))[-6:])
```

In other words, all requests in one testcase will have the same `HRUN-Request-ID` prefix, and each request will have a unique `HRUN-Request-ID` suffix.

## Client & Server IP:PORT

Sometimes, logging remote server IP and port can be great helpful for trouble shooting, especially when there are multiple servers and we want to checkout which one returns error.

Since version `3.0.13`, HttpRunner will log client & server IP:Port in debug level. 

```text
2020-06-18 11:32:38.366 | INFO     | httprunner.runner:__run_step:278 - run step begin: post raw text >>>>>>
2020-06-18 11:32:38.687 | DEBUG    | httprunner.client:request:187 - client IP: 10.90.205.63, Port: 62802
2020-06-18 11:32:38.687 | DEBUG    | httprunner.client:request:195 - server IP: 34.233.204.163, Port: 443
```

as well as testcase summary.

```text
"address": {
    "client_ip": "10.90.205.63",
    "client_port": 62802,
    "server_ip": "34.233.204.163",
    "server_port": 443
},
```

## arguments for v2.x compatibility

Besides all the arguments of `pytest`, `hrun` also has several other arguments to keep compatibility with HttpRunner v2.x.

- `--failfast`: has no effect, this argument will be removed automatically
- `--report-file`: specify html report file path, this argument will be replaced with `--html --self-contained-html` and generate html report with `pytest-html` plugin
- `--save-tests`: if set, HttpRunner v3.x will create a pytest conftest.py file containing session fixture to aggregate each testcase's summary and dumps to summary.json
