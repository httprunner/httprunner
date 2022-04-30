# Release History

## 2.5.9 (2022-04-30)

- fix #1217: reload debugtalk.py if loaded
- fix #1246: catch exceptions caused by GA report failure
- refactor: format code with black

## 2.5.8 (2022-03-23)

- change: replace events reporter from sentry to Google Analytics

## 2.5.7 (2020-02-21)

**Changed**

- feat: validate with python script, display print message

**Fixed**

- fix: validate script missing indents in html report
- fix: validate with python script, display lineno error

## 2.5.6 (2020-02-19)

**Added**

- feat: save variables and export data to JSON files (named xx.io.json) when specified `--save-tests`

**Changed**

- change: alter HttpRunner default log_level to WARNING

**Fixed**

- fix: abort test when failed to parse all cases
- fix: log error when parse failed

## 2.5.5 (2020-01-06)

**Fixed**

- fix: HTTP method missed "CONNECT", "TRACE"

**Changed**

- change: remove method validation from runner.Runner

## 2.5.4 (2020-01-03)

**Added**

- doc: add examples in json schema

**Fixed**

- fix #835: UnicodeDecodeError when loading json schema files
- fix: RecursionError caused by checking root dir incorrectly on Windows

## 2.5.3 (2020-01-03)

**Fixed**

- fix json schema: variables maybe in string type, e.g. '${prepare_variables()}'
- fix json schema: post json maybe in string type, e.g. '${prepare_post_data()}', '$post_data'

## 2.5.2 (2020-01-02)

**Fixed**

- fix #826: Windows does not support file name include ":"
- fix #819: maximum recursion error in locusts
- fix #818: request missed url in setup_hooks
- fix #808: project_working_directory is not initialized when running passed in data structure

## 2.5.1 (2020-01-02)

**Fixed**

- fix: RefResolutionError on Windows platform

## 2.5.0 (2020-01-01)

**Added**

- feat: add json schema validation for api
- feat: add json schema validation for testcase v1 & v2
- feat: add json schema validation for testsuite v1 & v2

**Changed**

- refactor: use loader.load_cases to validate test files
- refactor: use is_test_path to check if path is valid json/yaml file or a existed directory
- refactor: use is_test_content to check if data_structure is apis/testcases/testsuites

## 2.4.9 (2019-12-29)

**Added**

- test: add unittest for cli

**Changed**

- change: html report name defaults to be in UTC ISO 8601 format

**Fixed**

- fix: display validators in report when validate raised exception
- fix: eval validator python script before validating
- fix: do not strip string content when preparing lazy data
- fix: catch ApiNotFound exception when loading testcases
- fix: print exception string with exception stage

## 2.4.8 (2019-12-25)

**Added**

- feat: store parse failed api/testcase/testsuite file path in `logs/xxx.parse_failed.json`
- feat: add exception SummaryEmpty

**Fixed**

- fix: display request & response details in report when extraction failed
- fix: include CHANGELOG in package

**Changed**

- change: use sys.exit(code) in hrun main

## 2.4.7 (2019-12-24)

**Added**

- feat: report user id to sentry

**Fixed**

- fix #797: locusts command error

## 2.4.6 (2019-12-23)

**Added**

- feat: report tests start event and running exception to sentry

**Fixed**

- fix: ensure initializing sentry_sdk on startup

**Fixed**

## 2.4.5 (2019-12-20)

**Added**

- feat: integrate sentry sdk

**Fixed**

- fix: catch UnicodeDecodeError when json loads request body
- fix: display indented json for request json body

**Changed**

- change: detect request/response bytes encoding, instead of assuming utf-8
- refactor: make report as submodule

## 2.4.4 (2019-12-17)

**Added**

- feat: add keyword `body` to reference response body

**Changed**

- refactor: dumps request/response headers, display indented json in html report
- refactor: dumps request/response body if it is in json format, display indented json in html report
- change: unify response field(content/json/text) to `body` in html report

## 2.4.3 (2019-12-16)

**Added**

- feat: load api content on demand

**Changed**

- refactor: use poetry>=1.0.0
- test: migrate from travis CI to github actions
- test: migrate from coveralls to codecov
- test: run matrix tests on linux/macos/~~windows~~ and Python 2.7/3.5/3.6/3.7/3.8

## 2.4.2 (2019-12-13)

**Changed**

- refactor: replace with open file handler, avoid reading files into memory
- refactor: rename plugin to extension, httprunner/plugins -> httprunner/ext
- docs: update installation doc for developers

## 2.4.1 (2019-12-12)

**Added**

- feat: add `upload` keyword for upload test, see [doc](https://docs.httprunner.org/prepare/upload-case/)
- test: pip install package
- test: hrun command

**Fixed**

- fix: typo testfile_paths
- fix: check if locustio installed
- fix: dump json file name is empty when running relative testfile

## 2.4.0 (2019-12-11)

**Added**

- feat: validate with python script, ref #773
- feat: rearrange html report, failed testcases will be displayed on top.

**Changed**

- refactor: make loader as submodule, split to check/locate/load/buildup
- refactor: make built_in as submodule, split to comparators and functions
- refactor: adjust code for context and validator
- docs: update cli argument help
- adjust format code, remove unused import

**Fixed**

- fix: keep setup/teardown hooks original order when merge & override.
- fix: length comparator exceptions when running in CSV data-driven mode.

## 2.3.3 (2019-12-04)

**Fixed**

- fix #768: dump json file path error when folder name contains dot, such as `a.b.c`

**Changed**

- change: rename builtin function, sleep_N_secs => sleep

## 2.3.2 (2019-11-01)

**Added**

- docs: add docs content to repo, visit at `https://docs.httprunner.org`
- docs: update developer interface docs

**Changed**

- rename `render_html_report` to `gen_html_report`
- make gen_html_report separate with HttpRunner().run_tests()
- `--report-file`: specify report file path, this has higher priority than specifying report dir.
- remove `summary` property from HttpRunner

## 2.3.1 (2019-10-28)

**Fixed**

- fix locusts entry configuration

**Changed**

- update PyPi classifiers

## 2.3.0 (2019-10-27)

**Added**

- feat: implement plugin system prototype, make locusts as plugin
- test: add Python 3.8 to Travis-CI
- feat: add `__main__.py`, `python -m httprunner` can be used to hrun tests

**Changed**

- update dependency versions in pyproject.toml
- rename folder, httprunner/templates => httprunner/static
- log httprunner version before running tests
- remove unused import & code

**Fixed**

- fix #707: duration stat error in multiple testsuites

## 2.2.6 (2019-09-18)

**Added**

- feat: config variables support parsing from function
- feat: support [jsonpath](https://goessner.net/articles/JsonPath/) to parse json response [#679](https://github.com/httprunner/httprunner/pull/679)
- feat: generate html report with specified report file [#704](https://github.com/httprunner/httprunner/pull/704)

**Changed**

- remove unused import
- adjust code format

**Fixed**

- fix: dev-rules link 404

## 2.2.5 (2019-07-28)

**Added**

- log HttpRunner version when initializing

**Fixed**

- fix #658: sys.exit 1 if any testcase failed
- fix ModuleNotFoundError in debugging mode if httprunner uninstalled

## 2.2.4 (2019-07-18)

**Changed**

- replace pipenv & setup.py with poetry
- drop support for Python 3.4 as it was EOL on 2019-03-16
- relocate debugging scripts, move from main-debug.py to httprunner.cli

**Fixed**

- fix #574: delete unnecessary code
- fix #551: raise if times is not digit
- fix #572: tests_def_mapping["testcases"] typo error

## 2.2.3 (2019-06-30)

**Fixed**

- fix yaml FullLoader AttributeError when PyYAML version < 5.1

## 2.2.2 (2019-06-26)

**Changed**

- `extract` is used to replace `output` when passing former teststep's (as a testcase) export value to next teststep
- `export` is used to replace `output` in testcase config

## 2.2.1 (2019-06-25)

**Added**

- add demo api/testcase/testsuite to new created scaffold project
- update default `.gitignore` of new created scaffold project
- add demo content to `debugtalk.py`/`.env` of new created scaffold project

**Fixed**

- fix extend with testcase reference in format version 2
- fix ImportError when locustio is not installed
- fix YAMLLoadWarning by specify yaml loader

## 2.2.0 (2019-06-24)

**Added**

- support testcase/testsuite in format version 2

**Fixed**

- add wheel in dev packages
- fix exception when teststep name reference former extracted variable

## 2.1.3 (2019-04-24)

**Fixed**

- replace eval mechanism with builtins to prevent security vulnerabilities
- ImportError for builtins in Python2.7

## 2.1.2 (2019-04-17)

**Added**

- support new variable notation ${var}
- use \$\$ to escape \$ notation
- add Python 3.7 for travis CI

**Fixed**

- match duplicate variable/function in single raw string
- escape '{' and '}' notation in raw string
- print_info: TypeError when value is None
- display api name when running api as testcase

## 2.1.1 (2019-04-11)

**Changed**

refactor upload files mechanism with [requests-toolbelt](https://toolbelt.readthedocs.io/en/latest/user.html#multipart-form-data-encoder):

- simplify usage syntax, detect mimetype with [filetype](https://github.com/h2non/filetype.py).
- support upload multiple fields.

## 2.1.0 (2019-04-10)

**Added**

- implement json dump Python objects when save tests
- implement lazy parser
- remove project_mapping from parse_tests result

**Fixed**

- reference output variables
- pass output variables between testcases

## 2.0.6 (2019-03-18)

**Added**

- create .gitignore file when initializing new project

**Fixed**

- fix CSV relative path detection
- fix current validators displaying the former one when they are empty

## 2.0.5 (2019-03-04)

**Added**

- implement method to get variables and output

**Fixed**

- fix xss in response json

## 2.0.4 (2019-02-28)

**Fixed**

- fix verify priority with nested testcase
- fix function in config variables called multiple times
- dump loaded tests when running tests_mapping directly

## 2.0.3 (2019-02-24)

**Fixed**

- fix verify priority: teststep > config
- fix Chinese charactor in log_file encoding error in Windows
- fix dump file with Chinese charactor in Python 3

## 2.0.2 (2019-01-21)

**Fixed**

- each teststeps in one testcase share the same session
- fix duplicate API definition output

**Changed**

- display result from hook functions in DEBUG level log
- change log level of "Variables & Output" to INFO
- print Invalid testcase path or testcases
- print testcase output in INFO level log

## 2.0.1 (2019-01-18)

**Fixed**

- override current teststep variables with former testcase output variables
- Fixed compatibility when testcase name is empty
- skip undefined variable when parsing string content

**Changed**

- add back request method in report

## 2.0.0 (2019-01-01)

**Changed**

- Massive Refactor and Simplification
- Redesign testcase structure
- Module pipline
- Start Semantic Versioning
- Switch to Apache 2.0 license
- Change logo
