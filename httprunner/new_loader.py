import importlib
import io
import json
import os
import sys
import types
from typing import Tuple, Dict

import yaml
from loguru import logger
from pydantic import ValidationError

from httprunner import builtin, utils
from httprunner import exceptions
from httprunner.schema import TestCase, ProjectMeta

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


def load_testcase_file(testcase_file) -> Tuple[Dict, TestCase]:
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

    try:
        # validate with pydantic TestCase model
        testcase_obj = TestCase.parse_obj(testcase_content)
    except ValidationError as ex:
        err_msg = f"Invalid testcase format: {testcase_file}"
        logger.error(f"{err_msg}\n{ex}")
        raise exceptions.TestCaseFormatError(err_msg)

    testcase_content["config"]["path"] = testcase_file
    testcase_obj.config.path = testcase_file

    return testcase_content, testcase_obj


def load_dot_env_file(dot_env_path):
    """ load .env file.

    Args:
        dot_env_path (str): .env file path

    Returns:
        dict: environment variables mapping

            {
                "UserName": "debugtalk",
                "Password": "123456",
                "PROJECT_KEY": "ABCDEFGH"
            }

    Raises:
        exceptions.FileFormatError: If .env file format is invalid.

    """
    if not os.path.isfile(dot_env_path):
        return {}

    logger.info(f"Loading environment variables from {dot_env_path}")
    env_variables_mapping = {}

    with io.open(dot_env_path, "r", encoding="utf-8") as fp:
        for line in fp:
            # maxsplit=1
            if "=" in line:
                variable, value = line.split("=", 1)
            elif ":" in line:
                variable, value = line.split(":", 1)
            else:
                raise exceptions.FileFormatError(".env format error")

            env_variables_mapping[variable.strip()] = value.strip()

    utils.set_os_environ(env_variables_mapping)
    return env_variables_mapping


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


def locate_file(start_path, file_name):
    """ locate filename and return absolute file path.
        searching will be recursive upward until current working directory or system root dir.

    Args:
        file_name (str): target locate file name
        start_path (str): start locating path, maybe file path or directory path

    Returns:
        str: located file path. None if file not found.

    Raises:
        exceptions.FileNotFound: If failed to locate file.

    """
    if os.path.isfile(start_path):
        start_dir_path = os.path.dirname(start_path)
    elif os.path.isdir(start_path):
        start_dir_path = start_path
    else:
        raise exceptions.FileNotFound(f"invalid path: {start_path}")

    file_path = os.path.join(start_dir_path, file_name)
    if os.path.isfile(file_path):
        return os.path.abspath(file_path)

    # current working directory
    if os.path.abspath(start_dir_path) == os.getcwd():
        raise exceptions.FileNotFound(f"{file_name} not found in {start_path}")

    # system root dir
    # Windows, e.g. 'E:\\'
    # Linux/Darwin, '/'
    parent_dir = os.path.dirname(start_dir_path)
    if parent_dir == start_dir_path:
        raise exceptions.FileNotFound(f"{file_name} not found in {start_path}")

    # locate recursive upward
    return locate_file(parent_dir, file_name)


def locate_debugtalk_py(start_path):
    """ locate debugtalk.py file

    Args:
        start_path (str): start locating path,
            maybe testcase file path or directory path

    Returns:
        str: debugtalk.py file path, None if not found

    """
    try:
        # locate debugtalk.py file.
        debugtalk_path = locate_file(start_path, "debugtalk.py")
    except exceptions.FileNotFound:
        debugtalk_path = None

    return debugtalk_path


def init_project_working_directory(test_path):
    """ this should be called at startup

        run test file:
            run_path -> load_cases -> load_project_data -> init_project_working_directory
        or run passed in data structure:
            run -> init_project_working_directory

    Args:
        test_path: specified testfile path

    Returns:
        (str, str): debugtalk.py path, project_working_directory

    """

    def prepare_path(path):
        if not os.path.exists(path):
            err_msg = f"path not exist: {path}"
            logger.error(err_msg)
            raise exceptions.FileNotFound(err_msg)

        if not os.path.isabs(path):
            path = os.path.join(os.getcwd(), path)

        return path

    test_path = prepare_path(test_path)

    # locate debugtalk.py file
    debugtalk_path = locate_debugtalk_py(test_path)

    global project_working_directory
    if debugtalk_path:
        # The folder contains debugtalk.py will be treated as PWD.
        project_working_directory = os.path.dirname(debugtalk_path)
    else:
        # debugtalk.py not found, use os.getcwd() as PWD.
        project_working_directory = os.getcwd()

    # add PWD to sys.path
    sys.path.insert(0, project_working_directory)

    return debugtalk_path, project_working_directory


def load_debugtalk_functions():
    """ load project debugtalk.py module functions
        debugtalk.py should be located in project working directory.

    Returns:
        dict: debugtalk module functions mapping
            {
                "func1_name": func1,
                "func2_name": func2
            }

    """
    # load debugtalk.py module
    imported_module = importlib.import_module("debugtalk")
    return load_module_functions(imported_module)


def load_project_data(test_path: str, dot_env_path: str = None) -> ProjectMeta:
    """ load api, testcases, .env, debugtalk.py functions.
        api/testcases folder is relative to project_working_directory

    Args:
        test_path (str): test file/folder path, locate pwd from this path.
        dot_env_path (str): specified .env file path

    Returns:
        project loaded api/testcases definitions,
            environments and debugtalk.py functions.

    """
    debugtalk_path, project_working_directory = init_project_working_directory(
        test_path
    )

    project_meta = {}

    # load .env file
    # NOTICE:
    # environment variable maybe loaded in debugtalk.py
    # thus .env file should be loaded before loading debugtalk.py
    dot_env_path = dot_env_path or os.path.join(project_working_directory, ".env")
    project_meta["env"] = load_dot_env_file(dot_env_path)

    if debugtalk_path:
        # load debugtalk.py functions
        debugtalk_functions = load_debugtalk_functions()
    else:
        debugtalk_functions = {}

    # locate PWD and load debugtalk.py functions
    project_meta["PWD"] = project_working_directory
    project_meta["functions"] = debugtalk_functions
    project_meta["test_path"] = os.path.abspath(test_path)[
        len(project_working_directory) + 1 :
    ]

    return ProjectMeta.parse_obj(project_meta)
