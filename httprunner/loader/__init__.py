"""
HttpRunner loader

- check: validate testcase data structure with JSON schema (TODO)
- locate: locate debugtalk.py, make it's dir as project root path
- load: load testcase files and relevant data, including debugtalk.py, .env, yaml/json api/testcases, csv, etc.
- buildup: assemble loaded content to httprunner testcase/testsuite data structure

"""

from httprunner.loader.check import is_testcase_path, is_testcases, validate_json_file
from httprunner.loader.locate import get_project_working_directory as get_pwd
from httprunner.loader.load import load_csv_file, load_builtin_functions
from httprunner.loader.buildup import load_cases, load_project_data

__all__ = [
    "is_testcase_path",
    "is_testcases",
    "validate_json_file",
    "get_pwd",
    "load_csv_file",
    "load_builtin_functions",
    "load_project_data",
    "load_cases"
]
