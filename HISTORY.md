# Release History

## 2.0.6 (2019-03-18)

**Features**

- create .gitignore file when initializing new project

**Bugfixes**

- fix CSV relative path detection
- fix current validators displaying the former one when they are empty

## 2.0.5 (2019-03-04)

**Features**

- implement method to get variables and output

**Bugfixes**

- fix xss in response json

## 2.0.4 (2019-02-28)

**Bugfixes**

- fix verify priority with nested testcase
- fix function in config variables called multiple times
- dump loaded tests when running tests_mapping directly

## 2.0.3 (2019-02-24)

**Bugfixes**

- fix verify priority: teststep > config
- fix Chinese charactor in log_file encoding error in Windows
- fix dump file with Chinese charactor in Python 3

## 2.0.2 (2019-01-21)

**Bugfixes**

- each teststeps in one testcase share the same session
- fix duplicate API definition output

**Improvements**

- display result from hook functions in DEBUG level log
- change log level of "Variables & Output" to INFO
- print Invalid testcase path or testcases
- print testcase output in INFO level log

## 2.0.1 (2019-01-18)

**Bugfixes**

- override current teststep variables with former testcase output variables
- make compatible with testcase name is empty
- skip undefined variable when parsing string content
- add back request method in report

## 2.0.0 (2019-01-01)

- Massive Refactor and Simplification
- Redesign testcase structure
- Module pipline
- Start Semantic Versioning
- Switch to Apache 2.0 license
- Change logo
