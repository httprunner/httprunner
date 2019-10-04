import csv
import importlib
import io
import json
import os
import sys

import yaml

from httprunner import built_in, exceptions, logger, utils, validator

try:
    # PyYAML version >= 5.1
    # ref: https://github.com/yaml/pyyaml/wiki/PyYAML-yaml.load(input)-Deprecation
    yaml.warnings({'YAMLLoadWarning': False})
except AttributeError:
    pass


###############################################################################
#   file loader
###############################################################################


def _check_format(file_path, content):
    """ check testcase format if valid
    """
    # TODO: replace with JSON schema validation
    if not content:
        # testcase file content is empty
        err_msg = u"Testcase file content is empty: {}".format(file_path)
        logger.log_error(err_msg)
        raise exceptions.FileFormatError(err_msg)

    elif not isinstance(content, (list, dict)):
        # testcase file content does not match testcase format
        err_msg = u"Testcase file content format invalid: {}".format(file_path)
        logger.log_error(err_msg)
        raise exceptions.FileFormatError(err_msg)


def load_yaml_file(yaml_file):
    """ load yaml file and check file content format
    """
    with io.open(yaml_file, 'r', encoding='utf-8') as stream:
        yaml_content = yaml.load(stream)
        _check_format(yaml_file, yaml_content)
        return yaml_content


def load_json_file(json_file):
    """ load json file and check file content format
    """
    with io.open(json_file, encoding='utf-8') as data_file:
        try:
            json_content = json.load(data_file)
        except exceptions.JSONDecodeError:
            err_msg = u"JSONDecodeError: JSON file format error: {}".format(json_file)
            logger.log_error(err_msg)
            raise exceptions.FileFormatError(err_msg)

        _check_format(json_file, json_content)
        return json_content


def load_csv_file(csv_file):
    """ load csv file and check file content format

    Args:
        csv_file (str): csv file path, csv file content is like below:

    Returns:
        list: list of parameters, each parameter is in dict format

    Examples:
        >>> cat csv_file
        username,password
        test1,111111
        test2,222222
        test3,333333

        >>> load_csv_file(csv_file)
        [
            {'username': 'test1', 'password': '111111'},
            {'username': 'test2', 'password': '222222'},
            {'username': 'test3', 'password': '333333'}
        ]

    """
    if not os.path.isabs(csv_file):
        project_working_directory = tests_def_mapping["PWD"] or os.getcwd()
        # make compatible with Windows/Linux
        csv_file = os.path.join(project_working_directory, *csv_file.split("/"))

    if not os.path.isfile(csv_file):
        # file path not exist
        raise exceptions.CSVNotFound(csv_file)

    csv_content_list = []

    with io.open(csv_file, encoding='utf-8') as csvfile:
        reader = csv.DictReader(csvfile)
        for row in reader:
            csv_content_list.append(row)

    return csv_content_list


def load_file(file_path):
    if not os.path.isfile(file_path):
        raise exceptions.FileNotFound("{} does not exist.".format(file_path))

    file_suffix = os.path.splitext(file_path)[1].lower()
    if file_suffix == '.json':
        return load_json_file(file_path)
    elif file_suffix in ['.yaml', '.yml']:
        return load_yaml_file(file_path)
    elif file_suffix == ".csv":
        return load_csv_file(file_path)
    else:
        # '' or other suffix
        err_msg = u"Unsupported file format: {}".format(file_path)
        logger.log_warning(err_msg)
        return []


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
            if not filename.endswith(('.yml', '.yaml', '.json')):
                continue

            filenames_list.append(filename)

        for filename in filenames_list:
            file_path = os.path.join(dirpath, filename)
            file_list.append(file_path)

        if not recursive:
            break

    return file_list


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

    logger.log_info("Loading environment variables from {}".format(dot_env_path))
    env_variables_mapping = {}

    with io.open(dot_env_path, 'r', encoding='utf-8') as fp:
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


def locate_file(start_path, file_name):
    """ locate filename and return absolute file path.
        searching will be recursive upward until current working directory.

    Args:
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
        raise exceptions.FileNotFound("invalid path: {}".format(start_path))

    file_path = os.path.join(start_dir_path, file_name)
    if os.path.isfile(file_path):
        return os.path.abspath(file_path)

    # current working directory
    if os.path.abspath(start_dir_path) in [os.getcwd(), os.path.abspath(os.sep)]:
        raise exceptions.FileNotFound("{} not found in {}".format(file_name, start_path))

    # locate recursive upward
    return locate_file(os.path.dirname(start_dir_path), file_name)


###############################################################################
#   debugtalk.py module loader
###############################################################################


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
        if validator.is_function(item):
            module_functions[name] = item

    return module_functions


def load_builtin_functions():
    """ load built_in module functions
    """
    return load_module_functions(built_in)


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


###############################################################################
#   testcase loader
###############################################################################


project_mapping = {}
tests_def_mapping = {
    "PWD": None,
    "api": {},
    "testcases": {}
}


def __extend_with_api_ref(raw_testinfo):
    """ extend with api reference

    Raises:
        exceptions.ApiNotFound: api not found

    """
    api_name = raw_testinfo["api"]

    # api maybe defined in two types:
    # 1, individual file: each file is corresponding to one api definition
    # 2, api sets file: one file contains a list of api definitions
    if not os.path.isabs(api_name):
        # make compatible with Windows/Linux
        api_path = os.path.join(tests_def_mapping["PWD"], *api_name.split("/"))
        if os.path.isfile(api_path):
            # type 1: api is defined in individual file
            api_name = api_path

    try:
        block = tests_def_mapping["api"][api_name]
        # NOTICE: avoid project_mapping been changed during iteration.
        raw_testinfo["api_def"] = utils.deepcopy_dict(block)
    except KeyError:
        raise exceptions.ApiNotFound("{} not found!".format(api_name))


def __extend_with_testcase_ref(raw_testinfo):
    """ extend with testcase reference
    """
    testcase_path = raw_testinfo["testcase"]

    if testcase_path not in tests_def_mapping["testcases"]:
        # make compatible with Windows/Linux
        testcase_path = os.path.join(
            project_mapping["PWD"],
            *testcase_path.split("/")
        )
        loaded_testcase = load_file(testcase_path)

        if isinstance(loaded_testcase, list):
            # make compatible with version < 2.2.0
            testcase_dict = load_testcase(loaded_testcase)
        elif isinstance(loaded_testcase, dict) and "teststeps" in loaded_testcase:
            # format version 2, implemented in 2.2.0
            testcase_dict = load_testcase_v2(loaded_testcase)
        else:
            raise exceptions.FileFormatError(
                "Invalid format testcase: {}".format(testcase_path))

        tests_def_mapping["testcases"][testcase_path] = testcase_dict
    else:
        testcase_dict = tests_def_mapping["testcases"][testcase_path]

    raw_testinfo["testcase_def"] = testcase_dict


def load_teststep(raw_testinfo):
    """ load testcase step content.
        teststep maybe defined directly, or reference api/testcase.

    Args:
        raw_testinfo (dict): test data, maybe in 3 formats.
            # api reference
            {
                "name": "add product to cart",
                "api": "/path/to/api",
                "variables": {},
                "validate": [],
                "extract": {}
            }
            # testcase reference
            {
                "name": "add product to cart",
                "testcase": "/path/to/testcase",
                "variables": {}
            }
            # define directly
            {
                "name": "checkout cart",
                "request": {},
                "variables": {},
                "validate": [],
                "extract": {}
            }

    Returns:
        dict: loaded teststep content

    """
    # reference api
    if "api" in raw_testinfo:
        __extend_with_api_ref(raw_testinfo)

    # TODO: reference proc functions
    # elif "func" in raw_testinfo:
    #     pass

    # reference testcase
    elif "testcase" in raw_testinfo:
        __extend_with_testcase_ref(raw_testinfo)

    # define directly
    else:
        pass

    return raw_testinfo


def load_testcase(raw_testcase):
    """ load testcase with api/testcase references.

    Args:
        raw_testcase (list): raw testcase content loaded from JSON/YAML file:
            [
                # config part
                {
                    "config": {
                        "name": "XXXX",
                        "base_url": "https://debugtalk.com"
                    }
                },
                # teststeps part
                {
                    "test": {...}
                },
                {
                    "test": {...}
                }
            ]

    Returns:
        dict: loaded testcase content
            {
                "config": {},
                "teststeps": [test11, test12]
            }

    """
    config = {}
    tests = []

    for item in raw_testcase:
        key, test_block = item.popitem()
        if key == "config":
            config.update(test_block)
        elif key == "test":
            tests.append(load_teststep(test_block))
        else:
            logger.log_warning(
                "unexpected block key: {}. block key should only be 'config' or 'test'.".format(key)
            )

    return {
        "config": config,
        "teststeps": tests
    }


def load_testcase_v2(raw_testcase):
    """ load testcase in format version 2.

    Args:
        raw_testcase (dict): raw testcase content loaded from JSON/YAML file:
            {
                "config": {
                    "name": "xxx",
                    "variables": {}
                }
                "teststeps": [
                    {
                        "name": "teststep 1",
                        "request" {...}
                    },
                    {
                        "name": "teststep 2",
                        "request" {...}
                    },
                ]
            }

    Returns:
        dict: loaded testcase content
            {
                "config": {},
                "teststeps": [test11, test12]
            }

    """
    raw_teststeps = raw_testcase.pop("teststeps")
    raw_testcase["teststeps"] = [
        load_teststep(teststep)
        for teststep in raw_teststeps
    ]
    return raw_testcase


def load_testsuite(raw_testsuite):
    """ load testsuite with testcase references.
        support two different formats.

    Args:
        raw_testsuite (dict): raw testsuite content loaded from JSON/YAML file:
            # version 1, compatible with version < 2.2.0
            {
                "config": {
                    "name": "xxx",
                    "variables": {}
                }
                "testcases": {
                    "testcase1": {
                        "testcase": "/path/to/testcase",
                        "variables": {...},
                        "parameters": {...}
                    },
                    "testcase2": {}
                }
            }

            # version 2, implemented in 2.2.0
            {
                "config": {
                    "name": "xxx",
                    "variables": {}
                }
                "testcases": [
                    {
                        "name": "testcase1",
                        "testcase": "/path/to/testcase",
                        "variables": {...},
                        "parameters": {...}
                    },
                    {}
                ]
            }

    Returns:
        dict: loaded testsuite content
            {
                "config": {},
                "testcases": [testcase1, testcase2]
            }

    """
    raw_testcases = raw_testsuite.pop("testcases")
    raw_testsuite["testcases"] = {}

    if isinstance(raw_testcases, dict):
        # make compatible with version < 2.2.0
        for name, raw_testcase in raw_testcases.items():
            __extend_with_testcase_ref(raw_testcase)
            raw_testcase.setdefault("name", name)
            raw_testsuite["testcases"][name] = raw_testcase

    elif isinstance(raw_testcases, list):
        # format version 2, implemented in 2.2.0
        for raw_testcase in raw_testcases:
            __extend_with_testcase_ref(raw_testcase)
            testcase_name = raw_testcase["name"]
            raw_testsuite["testcases"][testcase_name] = raw_testcase

    else:
        # invalid format
        raise exceptions.FileFormatError("Invalid testsuite format!")

    return raw_testsuite


def load_test_file(path):
    """ load test file, file maybe testcase/testsuite/api

    Args:
        path (str): test file path

    Returns:
        dict: loaded test content

            # api
            {
                "path": path,
                "type": "api",
                "name": "",
                "request": {}
            }

            # testcase
            {
                "path": path,
                "type": "testcase",
                "config": {},
                "teststeps": []
            }

            # testsuite
            {
                "path": path,
                "type": "testsuite",
                "config": {},
                "testcases": {}
            }

    """
    raw_content = load_file(path)
    loaded_content = None

    if isinstance(raw_content, dict):

        if "testcases" in raw_content:
            # file_type: testsuite
            # TODO: add json schema validation for testsuite
            loaded_content = load_testsuite(raw_content)
            loaded_content["path"] = path
            loaded_content["type"] = "testsuite"

        elif "teststeps" in raw_content:
            # file_type: testcase (format version 2)
            loaded_content = load_testcase_v2(raw_content)
            loaded_content["path"] = path
            loaded_content["type"] = "testcase"

        elif "request" in raw_content:
            # file_type: api
            # TODO: add json schema validation for api
            loaded_content = raw_content
            loaded_content["path"] = path
            loaded_content["type"] = "api"

        else:
            # invalid format
            raise exceptions.FileFormatError("Invalid test file format!")

    elif isinstance(raw_content, list) and len(raw_content) > 0:
        # file_type: testcase
        # make compatible with version < 2.2.0
        # TODO: add json schema validation for testcase
        loaded_content = load_testcase(raw_content)
        loaded_content["path"] = path
        loaded_content["type"] = "testcase"

    else:
        # invalid format
        raise exceptions.FileFormatError("Invalid test file format!")

    return loaded_content


def load_folder_content(folder_path):
    """ load api/testcases/testsuites definitions from folder.

    Args:
        folder_path (str): api/testcases/testsuites files folder.

    Returns:
        dict: api definition mapping.

            {
                "tests/api/basic.yml": [
                    {"api": {"def": "api_login", "request": {}, "validate": []}},
                    {"api": {"def": "api_logout", "request": {}, "validate": []}}
                ]
            }

    """
    items_mapping = {}

    for file_path in load_folder_files(folder_path):
        items_mapping[file_path] = load_file(file_path)

    return items_mapping


def load_api_folder(api_folder_path):
    """ load api definitions from api folder.

    Args:
        api_folder_path (str): api files folder.

            api file should be in the following format:
            [
                {
                    "api": {
                        "def": "api_login",
                        "request": {},
                        "validate": []
                    }
                },
                {
                    "api": {
                        "def": "api_logout",
                        "request": {},
                        "validate": []
                    }
                }
            ]

    Returns:
        dict: api definition mapping.

            {
                "api_login": {
                    "function_meta": {"func_name": "api_login", "args": [], "kwargs": {}}
                    "request": {}
                },
                "api_logout": {
                    "function_meta": {"func_name": "api_logout", "args": [], "kwargs": {}}
                    "request": {}
                }
            }

    """
    api_definition_mapping = {}

    api_items_mapping = load_folder_content(api_folder_path)

    for api_file_path, api_items in api_items_mapping.items():
        # TODO: add JSON schema validation
        if isinstance(api_items, list):
            for api_item in api_items:
                key, api_dict = api_item.popitem()
                api_id = api_dict.get("id") or api_dict.get("def") \
                    or api_dict.get("name")
                if key != "api" or not api_id:
                    raise exceptions.ParamsError(
                        "Invalid API defined in {}".format(api_file_path))

                if api_id in api_definition_mapping:
                    raise exceptions.ParamsError(
                        "Duplicated API ({}) defined in {}".format(
                            api_id, api_file_path))
                else:
                    api_definition_mapping[api_id] = api_dict

        elif isinstance(api_items, dict):
            if api_file_path in api_definition_mapping:
                raise exceptions.ParamsError(
                    "Duplicated API defined: {}".format(api_file_path))
            else:
                api_definition_mapping[api_file_path] = api_items

    return api_definition_mapping


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


def load_project_tests(test_path, dot_env_path=None):
    """ load api, testcases, .env, debugtalk.py functions.
        api/testcases folder is relative to project_working_directory

    Args:
        test_path (str): test file/folder path, locate pwd from this path.
        dot_env_path (str): specified .env file path

    Returns:
        dict: project loaded api/testcases definitions,
            environments and debugtalk.py functions.

    """
    # locate debugtalk.py file
    debugtalk_path = locate_debugtalk_py(test_path)

    if debugtalk_path:
        # The folder contains debugtalk.py will be treated as PWD.
        project_working_directory = os.path.dirname(debugtalk_path)
    else:
        # debugtalk.py not found, use os.getcwd() as PWD.
        project_working_directory = os.getcwd()

    # add PWD to sys.path
    sys.path.insert(0, project_working_directory)

    # load .env file
    # NOTICE:
    # environment variable maybe loaded in debugtalk.py
    # thus .env file should be loaded before loading debugtalk.py
    dot_env_path = dot_env_path or os.path.join(project_working_directory, ".env")
    project_mapping["env"] = load_dot_env_file(dot_env_path)

    if debugtalk_path:
        # load debugtalk.py functions
        debugtalk_functions = load_debugtalk_functions()
    else:
        debugtalk_functions = {}

    # locate PWD and load debugtalk.py functions

    project_mapping["PWD"] = project_working_directory
    built_in.PWD = project_working_directory
    project_mapping["functions"] = debugtalk_functions

    # load api
    tests_def_mapping["api"] = load_api_folder(os.path.join(project_working_directory, "api"))
    tests_def_mapping["PWD"] = project_working_directory


def load_tests(path, dot_env_path=None):
    """ load testcases from file path, extend and merge with api/testcase definitions.

    Args:
        path (str): testcase/testsuite file/foler path.
            path could be in 2 types:
                - absolute/relative file path
                - absolute/relative folder path
        dot_env_path (str): specified .env file path

    Returns:
        dict: tests mapping, include project_mapping and testcases.
              each testcase is corresponding to a file.
            {
                "project_mapping": {
                    "PWD": "XXXXX",
                    "functions": {},
                    "env": {}
                },
                "testcases": [
                    {   # testcase data structure
                        "config": {
                            "name": "desc1",
                            "path": "testcase1_path",
                            "variables": [],                    # optional
                        },
                        "teststeps": [
                            # test data structure
                            {
                                'name': 'test desc1',
                                'variables': [],    # optional
                                'extract': [],      # optional
                                'validate': [],
                                'request': {}
                            },
                            test_dict_2   # another test dict
                        ]
                    },
                    testcase_2_dict     # another testcase dict
                ],
                "testsuites": [
                    {   # testsuite data structure
                        "config": {},
                        "testcases": {
                            "testcase1": {},
                            "testcase2": {},
                        }
                    },
                    testsuite_2_dict
                ]
            }

    """
    if not os.path.exists(path):
        err_msg = "path not exist: {}".format(path)
        logger.log_error(err_msg)
        raise exceptions.FileNotFound(err_msg)

    if not os.path.isabs(path):
        path = os.path.join(os.getcwd(), path)

    load_project_tests(path, dot_env_path)
    tests_mapping = {
        "project_mapping": project_mapping
    }

    def __load_file_content(path):
        loaded_content = None
        try:
            loaded_content = load_test_file(path)
        except exceptions.FileFormatError:
            logger.log_warning("Invalid test file format: {}".format(path))

        if not loaded_content:
            pass
        elif loaded_content["type"] == "testsuite":
            tests_mapping.setdefault("testsuites", []).append(loaded_content)
        elif loaded_content["type"] == "testcase":
            tests_mapping.setdefault("testcases", []).append(loaded_content)
        elif loaded_content["type"] == "api":
            tests_mapping.setdefault("apis", []).append(loaded_content)

    if os.path.isdir(path):
        files_list = load_folder_files(path)
        for path in files_list:
            __load_file_content(path)

    elif os.path.isfile(path):
        __load_file_content(path)

    return tests_mapping
