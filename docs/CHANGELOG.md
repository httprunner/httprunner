# Release History

## v4.3.5 (2023-07-23)

- refactor: send events to Google Analytics 4, replace GA v1
- fix: failure unittests caused by httpbin.org, replace with docker service

**go version**

- fix #1603: ensure path suffix '/' exists

**python version**

- fix: upgrade pyyaml from 5.4.1 to 6.0.1, fix installing error
- refactor: update httprunner dependencies

## v4.3.4 (2023-06-01)

**go version**

- feat: add video crawler for feed and live
- feat: cache screenshot ocr texts
- feat: set testcase and request timeout in seconds
- feat: catch interrupt signal
- feat: add new exit code MobileUILaunchAppError/InterruptError/TimeoutError/MobileUIActivityNotMatchError/MobileUIPopupError/LoopActionNotFoundError
- feat: find text with regex
- feat: add UI ocr tags to summary
- feat: check android device offline when running shell failed
- feat: print hrp runner exit code when running finished
- feat: add screen resolution and step start time in summary
- refactor: replace OCR APIs with image APIs
- refactor: FindText(s) returns OCRText(s)
- refactor: merge ActionOption with DataOption
- change: exit with AndroidShellExecError code for adb shell failure
- change: request vedem ocr with uploading image
- change: remove ping/dns sub commands

## v4.3.3 (2023-04-19)

**go version**

- feat: add `sleep_random` to sleep random seconds, with weight for multiple time ranges
- feat: input text with adb
- feat: add adb `screencap` sub command
- feat: add `AssertAppInForeground` to check if the given package is in foreground
- feat: check if app is in foreground when step failed
- feat: add validator AssertAppInForeground and AssertAppNotInForeground
- feat: save screenshots of all steps including ocr and cv recognition process data
- fix: adb driver for TapFloat
- fix: stop logcat only when enabled
- fix: do not fail case when kill logcat error
- fix: take screenshot after each step
- fix: screencap compatibility for shell v1 and v2 protocol
- fix: display parsed url in html report
- fix: fast fail not closing the websocket connection
- fix #1467: failed to parse parameters with plugin functions
- fix #1549: avoid duplicate creating plugins
- fix #1547: generate html report failed for referenced testcases
- fix: setup hooks compatible with v3

## v4.3.2 (2022-12-26)

**go version**

- feat: run Android UI automation with adb by default, add `uixt.WithUIA2(true)` option to use uiautomator2
- refactor: remove unused APIs in UI automation
- refactor: convert cases by specifying from/to format
- change: remove traceroute/curl sub commands

## v4.3.1 (2022-12-22)

**go version**

- feat: add option WithScreenShot
- feat: run xctest before start ios automation
- feat: run step with specified loop times
- feat: add options for FindTexts
- feat: capture pcap file for iOS, including CLI `hrp ios pcap` and option `uixt.WithIOSPcapOptions(...)`
- feat: add performance monitor for iOS, including CLI `hrp ios perf` and options `uixt.WithIOSPerfOptions(...)`
- refactor: move all UI APIs to uixt pkg
- docs: add examples for UI APIs

## v4.3.0 (2022-10-27)

Release hrp sub package `uixt` to support iOS/Android UI automation testing 🎉

- feat: support iOS UI automation with [WebDriverAgent] and [gwda]
- feat: support Android UI automation with [uiautomator2] and [guia2]
- feat: support UI recognition with [OCR service] and [gwda-ext-opencv]

For iOS/Android device management:

- feat: integrage ios device management with [gidevice]
- feat: integrage android device management with [gadb]
- feat: add simple commands to interact with iOS/Android devices, try `hrp ios` and `hrp adb`

Other improvements:

- feat: exit with specified code for different exceptions
- refactor: make uixt/gadb/gidevice/boomer/httpstat as hrp sub package

## v4.2.1 (2022-09-01)

**go version**

- fix: hrp boom duration still limited without specifying `--run-time`

## v4.2.0 (2022-08-21)

**go version**

- feat: support distributed load testing on multi-machines
- feat: support run/boom/convert curl command(s)
- feat: add ping/dns/traceroute/curl sub commands
- feat: improve builtin uploading feature, support `@` indicator and inferring MIME type
- feat: hrp boom support setting duration of run time
- change: support omitting websocket url if not necessary
- change: support multiple websocket connections for each session
- fix: optimize websocket step initialization
- fix: reuse plugin instance if already initialized
- fix: deep copy api step to avoid data racing

## v4.1.6 (2022-07-04)

**go version**

- fix: support parameterize for step name
- fix: concurrent map writes error when uploading in boom mode
- fix: record all requests of referenced testcases in boom mode
- fix: failed to record the step error in html report

## v4.1.5 (2022-06-27)

**go version**

- feat: support setting global testcase timeout and step timeout
- feat: support uploading file by multipart/form-data
- change: set http request timeout default to 120s
- fix: insert response cookies into request for redirect requests
- fix: support log debug level for load testing
- fix: failed to load json/data content in api reference
- fix: failed to convert postman collection containing multipart/form-data requests to pytest
- fix: only get the first parameter in referenced testcase
- fix: support variable reference during extraction
- fix: simplify jmespath compatibility conversion
- refactor: simplify testcase converter

**python version**

- fix: failed to parse variable referenced in upload
- refactor: make pytest testcases

## v4.1.4 (2022-06-17)

**go version**

- feat: config pypi index url by setting environment `PYPI_INDEX_URL`
- fix: filter commented out functions when generating plugin file
- fix: failed to use parameters in referenced testcase
- fix: failed to run testcase if python3 is not available on windows
- fix: panic occurred when running API step failed
- fix: step name overrides referenced testcase name

**python version**

- feat: support skip for pytest
- feat: print request and response details in DEBUG level when running API cases
- fix: support None/dict/list format when printing sql response
- fix: omit pseudo header names for HTTP/1, e.g. :authority, :method, :path, :schema

## v4.1.3 (2022-06-14)

**go version**

- feat #1342: support specify custom python3 venv, priority is greater than $HOME/.hrp/venv
- feat: assert python3 package installed and version matched
- refactor: build plugin mechanism, cancel automatic installation of dependencies
- fix #1352: avoid conversion to exponential notation

**python version**

- fix: unexpected changes in step variables

## v4.1.2 (2022-06-09)

- feat: add Dockerfile
- fix #1336: extract package in Windows
- fix: install package on MinGW64 and Windows

**go version**

- fix #1331: use `str_eq` to assert string and digit equality
- fix: load overall `pick_order` strategy in parameters_setting
- fix: ensure all dependencies in debugtalk.py are installed
- fix: select parameters with `random` strategy
- change: remove `hrp har2case`, replace with `hrp convert`

**python version**

- feat #1316: add running log and request & response details in allure report

## v4.1.1 (2022-05-31)

- fix: failed to build debugtalk.go without go.mod
- fix: avoid to escape from html special characters like '&' in converted JSON testcase
- fix: display the full step name when referencing testcase in html report
- fix: failed to regenerate debugtalk_gen.go and .debugtalk_gen.py correctly

## v4.1.0 (2022-05-29)

- feat: add `wiki` sub-command to open httprunner website
- feat: add `build` sub-command for function plugin

**go version**

- feat #1268: convert postman collection to HttpRunner testcase
- feat #1291: run testcases in v2/v3 JSON/YAML format with hrp run/boom command
- feat #1280: support creating empty scaffold project
- fix #1308: load `.env` file as environment variables
- fix #1309: locate plugin file upward recursively until system root dir
- fix #1315: failed to generate a report in failfast mode
- refactor: move base_url to config `environs`
- refactor: implement testcase conversions with `hrp convert`

## v4.1.0-beta (2022-05-21)

- feat: add pre-commit-hook to format go/python code

**go version**

- feat: add boomer mode(standalone/master/worker)
- feat: support load testing with specified `--profile` configuration file
- fix: step request elapsed timing should contain ContentTransfer part
- fix #1288: unable to go get httprunner v4
- fix: panic when config didn't exist in testcase file
- fix: disable keep alive and improve RPS accuracy
- fix: improve RPS accuracy

**python version**

- feat: support new step type with SQL operation
- feat: support new step type with thrift protocol

## v4.0.0 (2022-05-05)

**go version**

- feat: stat HTTP request latencies (DNSLookup, TCP Connection and so on)
- feat: add builtin function `environ`/`ENV`
- fix: demo function compatibility
- fix #1240: losing host port in har2case
- fix: concurrent map write in parameterize
- change: get hrp version from aliyun OSS file when installing
- change: report more load testing metrics to prometheus

## v4.0.0-beta (2022-04-24)

- refactor: merge [hrp] into httprunner v4, which will include golang and python dual engine
- refactor: redesign `IStep` to make step extensible to support implementing new protocols and test types
- feat: disable GA events report by setting environment `DISABLE_GA=true`
- feat: disable sentry reports by setting environment `DISABLE_SENTRY=true`
- feat: prepare python3 venv in `~/.hrp/venv` before running

**go version**

- feat: add `--profile` flag for har2case to support overwrite headers/cookies with specified yaml/json profile file
- feat: support run testcases in specified folder path, including testcases in sub folders
- feat: support HTTP/2 protocol
- feat: support WebSocket protocol
- feat: convert YAML/JSON testcases to pytest scripts with `hrp convert`
- change: integrate [sentry sdk][sentry sdk] for panic reporting and analysis
- change: lock funplugin version when creating scaffold project
- fix: call referenced api/testcase with relative path

**python version**

- feat: support retry when test step failed
- feat: add `pytest` sub-command to run pytest scripts
- change: remove startproject, move all features to go version, replace with `hrp startproject`
- change: remove har2case, move all features to go version, replace with `hrp run`
- change: remove locust, you should run load tests with go version, replace with `hrp boom`
- change: remove fastapi and uvicorn dependencies
- change: add pytest.ini to make log colorful
- fix: ignore exceptions when reporting GA events
- fix: remove misuse of NoReturn in Python typing

## hrp-v0.8.0 (2022-03-22)

- feat: support hashicorp python plugin over gRPC
- feat: create scaffold with plugin option, `--py`(default), `--go`, `--ignore-plugin`
- feat: print statistics summary after load testing finished
- feat: support think time for api/load testing
- fix: update prometheus state to stopped on quit

## hrp-v0.7.0 (2022-03-15)

- feat: support API layer for testcase
- feat: support global headers for testcase
- feat: support call referenced testcase by path in YAML/JSON testcases
- fix: decode failure when content-encoding is deflate
- fix: unstable RPS when load testing in high concurrency

## hrp-v0.6.4 (2022-03-10)

- feat: both support gRPC(default) and net/rpc mode in hashicorp plugin, switch with environment `HRP_PLUGIN_TYPE`
- refactor: move submodule `plugin` to separate repo `github.com/httprunner/funplugin`
- refactor: replace builtin json library with `json-iterator/go` to improve performance

## hrp-v0.6.3 (2022-03-04)

- feat: support customized setup/teardown hooks (variable assignment not supported)
- feat: add flag `--log-plugin` to turn on plugin logging
- change: add short flag `-c` for `--continue-on-failure`
- change: use `--log-requests-off` flag to turn off request & response details logging
- fix: support posting body in json array format
- fix: testcase format compatibility with HttpRunner

## hrp-v0.6.2 (2022-02-22)

- feat: support text/html extraction with regex
- change: json unmarshal to json.Number when parsing data
- fix: omit pseudo header names for HTTP/1, e.g. :authority
- fix: generate `headers.\"Content-Type\"` in har2case
- fix: incorrect data type when extracting data using jmespath
- fix: decode response body in brotli/gzip/deflate formats
- fix: omit print request/response body for non-text content
- fix: parse data for request cookie value

## hrp-v0.6.1 (2022-02-17)

- change: json unmarshal to float64 when parsing data
- fix: set request Content-Type for posting json only when not specified
- fix: failed to generate API test report when data is null
- fix: panic when assertion function not exists
- fix: broadcast to all rendezvous at once when spawn done

## hrp-v0.6.0 (2022-02-08)

- feat: implement `rendezvous` mechanism for data driven
- feat: upload release artifacts to aliyun oss
- feat: dump tests summary for execution results
- feat: generate html report for API testing
- change: remove sentry sdk

## hrp-v0.5.3 (2022-01-25)

- change: download package assets from aliyun OSS
- fix: disable color logging on Windows
- fix: print stderr when exec command failed
- fix: build hashicorp plugin failed when creating scaffold

## hrp-v0.5.2 (2022-01-19)

- feat: support creating and calling custom functions with [hashicorp/go-plugin]
- feat: add scaffold demo with hashicorp plugin
- feat: report events for initializing plugin
- fix: log failures when the assertion failed

## hrp-v0.5.1 (2022-01-13)

- feat: support specifying running cycles for load testing
- fix: ensure last stats reported when stop running

## hrp-v0.5.0 (2022-01-08)

- feat: support creating and calling custom functions with [go plugin]
- feat: install hrp with one shell command
- feat: add `startproject` sub-command for creating scaffold project
- feat: report GA event for loading go plugin

## hrp-v0.4.0 (2022-01-05)

- feat: implement `parameterize` mechanism for data driven
- feat: add multiple builtin assertion methods and builtin functions

## hrp-v0.3.1 (2021-12-30)

- fix: set ulimit to 10240 before load testing
- fix: concurrent map writes in load testing

## hrp-v0.3.0 (2021-12-24)

- feat: implement `transaction` mechanism for load test
- feat: continue running next step when failure occurs with `--continue-on-failure` flag, default to failfast
- feat: report GA events with version
- feat: run load test with the given limit and burst as rate limiter, use `--spawn-count`, `--spawn-rate` and `--request-increase-rate` flag
- feat: report runner state to prometheus
- refactor: fork [boomer] as submodule initially and made a lot of changes
- change: update API models

## hrp-v0.2.2 (2021-12-07)

- refactor: update models to make API more concise
- change: remove mkdocs, move to [docs repo]

## hrp-v0.2.1 (2021-12-02)

- feat: push load testing metrics to [Prometheus Pushgateway][pushgateway]
- feat: report events with Google Analytics

## hrp-v0.2.0 (2021-11-19)

- feat: deploy mkdocs to github pages when PR merged
- feat: release hrp cli binaries automatically with github actions
- feat: add Makefile for running unittest and building hrp cli binary

## hrp-v0.1.0 (2021-11-18)

- feat: full support for HTTP(S)/1.1 methods
- feat: integrate [zerolog] for logging, include json log and pretty color console log
- feat: implement `har2case` for converting HAR to JSON/YAML testcases
- feat: extract and validate json response with [`jmespath`][jmespath]
- feat: run JSON/YAML testcases with builtin functions
- feat: support testcase and teststep level variables mechanism
- feat: integrate [boomer] standalone mode for load testing
- docs: init documentation website with [mkdocs]
- docs: add project badges, including go report card, codecov, github actions, FOSSA, etc.
- test: add CI test with [github actions][github-actions]
- test: integrate [sentry sdk][sentry sdk] for event reporting and analysis

## 3.1.11 (2022-04-24)

- fix #1273: ImportError by cannot import name '_unicodefun' from 'click'

## 3.1.10 (2022-04-18)

- fix #1249: catch exceptions when requesting with disabling allow_redirects
- fix: catch OSError when running subprocess

## 3.1.9 (2022-04-17)

- fix #1174: pydantic validation error when body is None
- fix #1209: only convert jmespath path for some fields in white list
- fix #1233: parse upload info with session variables
- fix #1246: catch exceptions caused by GA report failure
- fix #1247: catch exceptions when getting socket address failed

## 3.1.8 (2022-03-22)

- feat: add `--profile` flag for har2case to support overwrite headers/cookies with specified yaml/json configuration file
- feat: support variable and function in response extract expression
- fix: keep negative index in jmespath unchanged when converting pytest files, e.g. body.users[-1]
- fix: variable should not start with digit
- change: load yaml file with FullLoader

## 3.1.7 (2022-03-22)

- fix #1117: ignore comments and blank lines when parsing .env file
- fix #1141: parameterize failure caused by pydantic version
- fix #1165: ImportError caused by jinja2 version
- fix: failure in getting client and server IP/port when requesting HTTPS
- fix: upgrade dependencies for security
- change: remove support for dead python 3.6, upgrade supported python version to 3.7/3.8/3.9/3.10
- change: replace events reporter from sentry to Google Analytics

## 3.1.6 (2021-07-18)

**Fixed**

- fix #1086: chinese garbled in response
- fix #1068: incorrect variables and variable type hints
- fix #1079: display error in request body if the list inputted from with_json() contains dict
- fix #1056: validation failed when validation-value is in string format

## 3.1.5 (2021-06-27)

**Fixed**

- fix: decode brotli encoding

## 3.1.4 (2020-07-30)

**Changed**

- change: override variables strategy, step variables > extracted variables from previous steps

**Fixed**

- fix: parameters feature with custom functions
- fix: request json field with variable reference
- fix: pickle BufferedReader TypeError in upload feature

## 3.1.3 (2020-07-06)

**Added**

- feat: implement `parameters` feature

**Fixed**

- fix: validate with variable or function whose evaluation result is "" or not text
- fix: raise TestCaseFormatError if teststep validate invalid
- fix: raise TestCaseFormatError if ref testcase is invalid

## 3.1.2 (2020-06-29)

**Fixed**

- fix: missing setup/teardown hooks for referenced testcase
- fix: compatibility for `black` on Android termux that does not support multiprocessing well
- fix: mishandling of request header `Content-Length` for GET method
- fix: validate with jmespath containing variable or function, e.g. `body.locations[$index].name`

**Changed**

- change: import locust at beginning to monkey patch all modules
- change: open file in binary mode

## 3.1.1 (2020-06-23)

**Added**

- feat: add optional message for assertion

**Fixed**

- fix: ValueError when type_match None
- fix: override referenced testcase export in teststep
- fix: avoid duplicate import
- fix: override locust weight

## 3.1.0 (2020-06-21)

**Added**

- feat: integrate [locust] v1.0

**Changed**

- change: make converted referenced pytest files always relative to ProjectRootDir
- change: log function details when call function failed
- change: do not raise error if failed to get client/server address info

**Fixed**

- fix: path handling error when har2case har file and cwd != ProjectRootDir
- fix: missing list type for request body

## 3.0.13 (2020-06-17)

**Added**

- feat: log client/server IP and port

**Fixed**

- fix: avoid '.csv' been converted to '_csv'
- fix: convert har to JSON format testcase
- fix: missing ${var} handling in overriding config variables
- fix: SyntaxError caused by quote in case of headers."Set-Cookie"
- fix: FileExistsError when specified project name conflicts with existed file
- fix: testcase path handling error when path startswith "./" or ".\\"

## 3.0.12 (2020-06-14)

**Fixed**

- fix: compatibility with different path separators of Linux and Windows
- fix: IndexError in ensure_file_path_valid when file_path=os.getcwd()
- fix: ensure step referenced api, convert to v3 testcase
- fix: several other compatibility issues

**Changed**

- change: skip reporting sentry for errors occurred in debugtalk.py

## 3.0.11 (2020-06-08)

**Changed**

- change: override variables
  (1) testcase: session variables > step variables > config variables
  (2) testsuite: testcase variables > config variables
  (3) testsuite testcase variables > testcase config variables

**Fixed**

- fix: incorrect summary success when testcase failed
- fix: reload to refresh previously loaded debugtalk module
- fix: escape $$ in variable value

## 3.0.10 (2020-06-07)

**Added**

- feat: implement step setup/teardown hooks
- feat: support alter response in teardown hooks

**Fixed**

- fix: ensure upload ready
- fix: add ExtendJSONEncoder to safely dump json data with python object, such as MultipartEncoder

## 3.0.9 (2020-06-07)

**Fixed**

- fix: miss formatting referenced testcase
- fix: handle cases when parent directory name includes dot/hyphen/space

**Changed**

- change: add `export` keyword in TStep to export session variables from referenced testcase
- change: rename TestCaseInOut field, config_vars and export_vars
- change: rename StepData field, export_vars
- change: add `--tb=short` for `hrun` command to use shorter traceback format by default
- change: search debugtalk.py upward recursively until system root dir

## 3.0.8 (2020-06-04)

**Added**

- feat: add sentry sdk
- feat: extract session variable from referenced testcase step

**Fixed**

- fix: missing request json
- fix: override testsuite/testcase config verify
- fix: only strip whitespaces and tabs, \n\r are left because they maybe used in changeset
- fix: log testcase duration before raise ValidationFailure

**Changed**

- change: add httprunner version in generated pytest file

## 3.0.7 (2020-06-03)

**Added**

- feat: make pytest files in chain style
- feat: `hrun` supports run pytest files
- feat: get raw testcase model from pytest file

**Fixed**

- fix: convert jmespath.search result to int/float unintentionally
- fix: referenced testcase should not be run duplicately
- fix: requests.cookies.CookieConflictError, multiple cookies with name
- fix: missing exit code from pytest
- fix: skip invalid testcase/testsuite yaml/json file

**Changed**

- change: `har2case` generate pytest file by default
- docs: update sponsor info

## 3.0.6 (2020-05-29)

**Added**

- feat: make referenced testcase as pytest class

**Fixed**

- fix: ensure converted python file in utf-8 encoding
- fix: duplicate running referenced testcase
- fix: ensure compatibility issues between testcase format v2 and v3
- fix: ensure compatibility with deprecated cli args in v2, include --failfast/--report-file/--save-tests
- fix: UnicodeDecodeError when request body in protobuf

**Changed**

- change: make `allure-pytest`, `requests-toolbelt`, `filetype` as optional dependencies
- change: move all unittests to tests folder
- change: save testcase log in PWD/logs/ directory

## 3.0.5 (2020-05-22)

**Added**

- feat: each testcase has an unique id in uuid4 format
- feat: add default header `HRUN-Request-ID` for each testcase #721
- feat: builtin allure report
- feat: dump log for each testcase

**Fixed**

- fix: ensure referenced testcase share the same session

**Changed**

- change: remove default added `-s` option for hrun

## 3.0.4 (2020-05-19)

**Added**

- feat: make testsuite and run testsuite
- feat: testcase/testsuite config support getting variables by function
- feat: har2case with request cookies
- feat: log request/response headers and body with indent

**Fixed**

- fix: extract response cookies
- fix: handle errors when no valid testcases generated

**Changed**

- change: har2case do not ignore request headers, except for header startswith :

## 3.0.3 (2020-05-17)

**Fixed**

- fix: compatibility with testcase file path includes dots, space and minus sign
- fix: testcase generator, validate content.xxx => body.xxx
- fix: scaffold for v3

## 3.0.2 (2020-05-16)

**Added**

- feat: add `make` sub-command to generate python testcases from YAML/JSON
- feat: format generated python testcases with [black]
- test: add postman echo & httpbin as testcase examples

**Changed**

- refactor all
- replace jsonschema validation with pydantic
- remove compatibility with testcase/testsuite format v1
- replace unittest with pytest
- remove builtin html report, allure will be used with pytest later
- remove locust support temporarily
- update command line interface

## 3.0.1 (2020-03-24)

**Changed**

- remove sentry sdk

## 3.0.0 (2020-03-10)

**Added**

- feat: dump log for each testcase
- feat: add default header `HRUN-Request-ID` for each testcase #721

**Changed**

- remove support for Python 2.7
- replace logging with [loguru]
- replace string format with f-string
- remove dependency colorama and colorlog
- generate reports/logs folder in current working directory
- remove cli `--validate`
- remove cli `--pretty`

## 2.0 (2019-01-01~2020-02-21)

reference: [v2-changelog]


[hrp]: https://github.com/httprunner/hrp
[hashicorp/go-plugin]: https://github.com/hashicorp/go-plugin
[go plugin]: https://pkg.go.dev/plugin
[docs repo]: https://github.com/httprunner/httprunner.github.io
[zerolog]: https://github.com/rs/zerolog
[jmespath]: https://jmespath.org/
[mkdocs]: https://www.mkdocs.org/
[github-actions]: https://github.com/httprunner/hrp/actions
[boomer]: github.com/myzhan/boomer
[sentry sdk]: https://github.com/getsentry/sentry-go
[pushgateway]: https://github.com/prometheus/pushgateway
[locust]: https://locust.io/
[black]: https://github.com/psf/black
[loguru]: https://github.com/Delgan/loguru
[v2-changelog]: https://github.com/httprunner/httprunner/blob/v2/docs/CHANGELOG.md
[WebDriverAgent]: https://github.com/appium/WebDriverAgent
[uiautomator2]: https://github.com/appium/appium-uiautomator2-server
[gidevice]: https://github.com/electricbubble/gidevice
[gwda]: https://github.com/electricbubble/gwda
[guia2]: https://github.com/electricbubble/guia2
[gadb]: https://github.com/electricbubble/gadb
[OCR service]: https://www.volcengine.com/product/text-recognition
[gwda-ext-opencv]: https://github.com/electricbubble/gwda-ext-opencv
