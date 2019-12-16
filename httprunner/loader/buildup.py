import importlib
import os

from httprunner import exceptions, logger, utils
from httprunner.loader.load import load_module_functions, load_file, load_dot_env_file, \
    load_folder_files
from httprunner.loader.locate import init_project_working_directory, get_project_working_directory

tests_def_mapping = {
    "api": {},
    "testcases": {}
}


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
        pwd = get_project_working_directory()
        api_path = os.path.join(pwd, *api_name.split("/"))
        if os.path.isfile(api_path):
            # type 1: api is defined in individual file
            api_name = api_path

    if api_name in tests_def_mapping["api"]:
        block = tests_def_mapping["api"][api_name]
    elif not os.path.isfile(api_name):
        raise exceptions.ApiNotFound("{} not found!".format(api_name))
    else:
        block = load_file(api_name)

    # NOTICE: avoid project_mapping been changed during iteration.
    raw_testinfo["api_def"] = utils.deepcopy_dict(block)
    tests_def_mapping["api"][api_name] = block


def __extend_with_testcase_ref(raw_testinfo):
    """ extend with testcase reference
    """
    testcase_path = raw_testinfo["testcase"]

    if testcase_path not in tests_def_mapping["testcases"]:
        # make compatible with Windows/Linux
        pwd = get_project_working_directory()
        testcase_path = os.path.join(
            pwd,
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


def load_project_data(test_path, dot_env_path=None):
    """ load api, testcases, .env, debugtalk.py functions.
        api/testcases folder is relative to project_working_directory

    Args:
        test_path (str): test file/folder path, locate pwd from this path.
        dot_env_path (str): specified .env file path

    Returns:
        dict: project loaded api/testcases definitions,
            environments and debugtalk.py functions.

    """
    debugtalk_path, project_working_directory = init_project_working_directory(test_path)

    project_mapping = {}

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
    project_mapping["functions"] = debugtalk_functions
    project_mapping["test_path"] = os.path.abspath(test_path)

    return project_mapping


def load_cases(path, dot_env_path=None):
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

    tests_mapping = {
        "project_mapping": load_project_data(path, dot_env_path)
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
