import io
import json
import os
import types

import yaml
from loguru import logger

from httprunner import builtin
from httprunner import exceptions
from httprunner.schema import TestCase

try:
    # PyYAML version >= 5.1
    # ref: https://github.com/yaml/pyyaml/wiki/PyYAML-yaml.load(input)-Deprecation
    yaml.warnings({"YAMLLoadWarning": False})
except AttributeError:
    pass


def _load_yaml_file(yaml_file):
    """ load yaml file and check file content format
    """
    with io.open(yaml_file, "r", encoding="utf-8") as stream:
        try:
            yaml_content = yaml.load(stream)
        except yaml.YAMLError as ex:
            logger.error(str(ex))
            raise exceptions.FileFormatError

        return yaml_content


def _load_json_file(json_file):
    """ load json file and check file content format
    """
    with io.open(json_file, encoding="utf-8") as data_file:
        try:
            json_content = json.load(data_file)
        except json.JSONDecodeError:
            err_msg = f"JSONDecodeError: JSON file format error: {json_file}"
            logger.error(err_msg)
            raise exceptions.FileFormatError(err_msg)

        return json_content


def load_testcase_file(testcase_file):
    """load testcase file and validate with pydantic model"""
    file_suffix = os.path.splitext(testcase_file)[1].lower()
    if file_suffix == ".json":
        testcase_content = _load_json_file(testcase_file)
    elif file_suffix in [".yaml", ".yml"]:
        testcase_content = _load_yaml_file(testcase_file)
    else:
        # '' or other suffix
        raise exceptions.FileFormatError(
            f"testcase file should be YAML/JSON format, invalid testcase file: {testcase_file}"
        )

    # validate with pydantic TestCase model
    TestCase.parse_obj(testcase_content)

    return testcase_content


def load_folder_files(folder_path, recursive=True):
    """ load folder path, return all files endswith yml/yaml/json in list.

    Args:
        folder_path (str): specified folder path to load
        recursive (bool): load files recursively if True

    Returns:
        list: files endswith yml/yaml/json
    """
    if isinstance(folder_path, (list, set)):
        files = []
        for path in set(folder_path):
            files.extend(load_folder_files(path, recursive))

        return files

    if not os.path.exists(folder_path):
        return []

    file_list = []

    for dirpath, dirnames, filenames in os.walk(folder_path):
        filenames_list = []

        for filename in filenames:
            if not filename.endswith((".yml", ".yaml", ".json")):
                continue

            filenames_list.append(filename)

        for filename in filenames_list:
            file_path = os.path.join(dirpath, filename)
            file_list.append(file_path)

        if not recursive:
            break

    return file_list


def load_module_functions(module):
    """ load python module functions.

    Args:
        module: python module

    Returns:
        dict: functions mapping for specified python module

            {
                "func1_name": func1,
                "func2_name": func2
            }

    """
    module_functions = {}

    for name, item in vars(module).items():
        if isinstance(item, types.FunctionType):
            module_functions[name] = item

    return module_functions


def load_builtin_functions():
    """ load builtin module functions
    """
    return load_module_functions(builtin)
