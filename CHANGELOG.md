# Release History

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
