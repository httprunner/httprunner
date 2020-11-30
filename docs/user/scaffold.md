# Scaffold

If you want to create a new project, you can use the scaffold to startup quickly.

## help

```text
$ httprunner startproject -h
usage: httprunner startproject [-h] [project_name]

positional arguments:
  project_name  Specify new project name.

optional arguments:
  -h, --help    show this help message and exit
```

## create new project

The only argument you need to specify is the project name.

```text
$ httprunner startproject demo
2020-06-15 11:53:25.498 | INFO     | httprunner.scaffold:create_scaffold:37 - Create new project: demo
Project Root Dir: /Users/debugtalk/MyProjects/HttpRunner-dev/HttpRunner/demo

created folder: demo
created folder: demo/har
created folder: demo/testcases
created folder: demo/reports
created file: demo/testcases/demo_testcase_request.yml
created file: demo/testcases/demo_testcase_ref.yml
created file: demo/debugtalk.py
created file: demo/.env
created file: demo/.gitignore

$ tree demo -a
demo
├── .env
├── .gitignore
├── debugtalk.py
├── har
├── reports
└── testcases
    ├── demo_testcase_ref.yml
    └── demo_testcase_request.yml

3 directories, 5 files
```

If you specify a project name that already exists, you will get a warning.

```text
$  httprunner startproject demo
2020-06-15 11:55:03.192 | WARNING  | httprunner.scaffold:create_scaffold:32 - Project demo exists, please specify a new project name.

$ tree demo -a
demo
├── .env
├── .gitignore
├── debugtalk.py
├── har
├── reports
└── testcases
    ├── demo_testcase_ref.yml
    └── demo_testcase_request.yml

3 directories, 5 files
```

## run scaffold project

The scaffold project has several valid testcases, so you can run tests without any edit.

```text
$ hrun demo
2020-06-15 11:57:15.883 | INFO     | httprunner.loader:load_dot_env_file:130 - Loading environment variables from /Users/debugtalk/MyProjects/HttpRunner-dev/HttpRunner/demo/.env
2020-06-15 11:57:15.883 | DEBUG    | httprunner.utils:set_os_environ:32 - Set OS environment variable: USERNAME
2020-06-15 11:57:15.884 | DEBUG    | httprunner.utils:set_os_environ:32 - Set OS environment variable: PASSWORD
2020-06-15 11:57:15.885 | INFO     | httprunner.make:make_testcase:310 - start to make testcase: /Users/debugtalk/MyProjects/HttpRunner-dev/HttpRunner/demo/testcases/demo_testcase_ref.yml
2020-06-15 11:57:15.898 | INFO     | httprunner.make:make_testcase:310 - start to make testcase: /Users/debugtalk/MyProjects/HttpRunner-dev/HttpRunner/demo/testcases/demo_testcase_request.yml
2020-06-15 11:57:15.899 | INFO     | httprunner.make:make_testcase:383 - generated testcase: /Users/debugtalk/MyProjects/HttpRunner-dev/HttpRunner/demo/testcases/demo_testcase_request_test.py
2020-06-15 11:57:15.900 | INFO     | httprunner.make:make_testcase:383 - generated testcase: /Users/debugtalk/MyProjects/HttpRunner-dev/HttpRunner/demo/testcases/demo_testcase_ref_test.py
2020-06-15 11:57:15.911 | INFO     | httprunner.make:make_testcase:310 - start to make testcase: /Users/debugtalk/MyProjects/HttpRunner-dev/HttpRunner/demo/testcases/demo_testcase_request.yml
2020-06-15 11:57:15.912 | INFO     | httprunner.make:__ensure_project_meta_files:128 - copy .env to /Users/debugtalk/MyProjects/HttpRunner-dev/HttpRunner/demo/_env
2020-06-15 11:57:15.912 | INFO     | httprunner.make:format_pytest_with_black:147 - format pytest cases with black ...
reformatted /Users/debugtalk/MyProjects/HttpRunner-dev/HttpRunner/demo/testcases/demo_testcase_ref_test.py
reformatted /Users/debugtalk/MyProjects/HttpRunner-dev/HttpRunner/demo/testcases/demo_testcase_request_test.py
All done! ✨ 🍰 ✨
2 files reformatted, 1 file left unchanged.
2020-06-15 11:57:16.299 | INFO     | httprunner.cli:main_run:56 - start to run tests with pytest. HttpRunner version: 3.0.12
====================================================================== test session starts ======================================================================
platform darwin -- Python 3.7.5, pytest-5.4.2, py-1.8.1, pluggy-0.13.1
rootdir: /Users/debugtalk/MyProjects/HttpRunner-dev/HttpRunner
plugins: metadata-1.9.0, allure-pytest-2.8.16, html-2.1.1
collected 2 items

demo/testcases/demo_testcase_request_test.py .                                                                                                            [ 50%]
demo/testcases/demo_testcase_ref_test.py .                                                                                                                [100%]

======================================================================= 2 passed in 6.87s =======================================================================
```
