import collections
import csv
import importlib
import io
import json
import os
import sys

import yaml
from httprunner import built_in, exceptions, logger, parser, utils, validator
from httprunner.compat import OrderedDict

sys.path.insert(0, os.getcwd())

project_mapping = {
    "debugtalk": {
        "variables": {},
        "functions": {}
    },
    "env": {},
    "def-api": {},
    "def-testcase": {}
}
""" dict: save project loaded api/testcases definitions, environments and debugtalk.py module.
"""

dot_env_path = None

testcases_cache_mapping = {}
project_working_directory = os.getcwd()


###############################################################################
##   file loader
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
    @param
        csv_file: csv file path
        e.g. csv file content:
            username,password
            test1,111111
            test2,222222
            test3,333333
    @return
        list of parameter, each parameter is in dict format
        e.g.
        [
            {'username': 'test1', 'password': '111111'},
            {'username': 'test2', 'password': '222222'},
            {'username': 'test3', 'password': '333333'}
        ]
    """
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


def load_dot_env_file():
    """ load .env file, .env file should be located in project working directory by default.
        If dot_env_path is specified, it will be loaded instead.

    Returns:
        dict: environment variables mapping

            {
                "UserName": "debugtalk",
                "Password": "123456",
                "PROJECT_KEY": "ABCDEFGH"
            }

    Raises:
        exceptions.FileFormatError: If env file format is invalid.

    """
    path = dot_env_path or os.path.join(project_working_directory, ".env")
    if not os.path.isfile(path):
        if dot_env_path:
            logger.log_error(".env file not exist: {}".format(dot_env_path))
            sys.exit(1)
        else:
            logger.log_debug(".env file not exist in: {}".format(project_working_directory))
            return {}

    logger.log_info("Loading environment variables from {}".format(path))
    env_variables_mapping = {}
    with io.open(path, 'r', encoding='utf-8') as fp:
        for line in fp:
            if "=" in line:
                variable, value = line.split("=")
            elif ":" in line:
                variable, value = line.split(":")
            else:
                raise exceptions.FileFormatError(".env format error")

            env_variables_mapping[variable.strip()] = value.strip()

    project_mapping["env"] = env_variables_mapping
    utils.set_os_environ(env_variables_mapping)

    return env_variables_mapping


def locate_file(start_path, file_name):
    """ locate filename and return file path.
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
        return file_path

    # current working directory
    if os.path.abspath(start_dir_path) in [os.getcwd(), os.path.abspath(os.sep)]:
        raise exceptions.FileNotFound("{} not found in {}".format(file_name, start_path))

    # locate recursive upward
    return locate_file(os.path.dirname(start_dir_path), file_name)


###############################################################################
##   debugtalk.py module loader
###############################################################################

def load_python_module(module):
    """ load python module.

    Args:
        module: python module

    Returns:
        dict: variables and functions mapping for specified python module

            {
                "variables": {},
                "functions": {}
            }

    """
    debugtalk_module = {
        "variables": {},
        "functions": {}
    }

    for name, item in vars(module).items():
        if validator.is_function((name, item)):
            debugtalk_module["functions"][name] = item
        elif validator.is_variable((name, item)):
            debugtalk_module["variables"][name] = item
        else:
            pass

    return debugtalk_module


def load_builtin_module():
    """ load built_in module
    """
    built_in_module = load_python_module(built_in)
    project_mapping["debugtalk"] = built_in_module


def load_debugtalk_module():
    """ load project debugtalk.py module and merge with builtin module.
        debugtalk.py should be located in project working directory.
        variables and functions mapping for debugtalk.py
            {
                "variables": {},
                "functions": {}
            }

    """
    # load debugtalk.py module
    imported_module = importlib.import_module("debugtalk")
    debugtalk_module = load_python_module(imported_module)

    # override built_in module with debugtalk.py module
    project_mapping["debugtalk"]["variables"].update(debugtalk_module["variables"])
    project_mapping["debugtalk"]["functions"].update(debugtalk_module["functions"])


def get_module_item(module_mapping, item_type, item_name):
    """ get expected function or variable from module mapping.

    Args:
        module_mapping(dict): module mapping with variables and functions.

            {
                "variables": {},
                "functions": {}
            }

        item_type(str): "functions" or "variables"
        item_name(str): function name or variable name

    Returns:
        object: specified variable or function object.

    Raises:
        exceptions.FunctionNotFound: If specified function not found in module mapping
        exceptions.VariableNotFound: If specified variable not found in module mapping

    """
    try:
        return module_mapping[item_type][item_name]
    except KeyError:
        err_msg = "{} not found in debugtalk.py module!\n".format(item_name)
        err_msg += "module mapping: {}".format(module_mapping)
        if item_type == "functions":
            raise exceptions.FunctionNotFound(err_msg)
        else:
            raise exceptions.VariableNotFound(err_msg)


###############################################################################
##   testcase loader
###############################################################################

def _load_test_file(file_path):
    """ load testcase file or testsuite file

    Args:
        file_path (str): absolute valid file path. file_path should be in the following format:

            [
                {
                    "config": {
                        "name": "",
                        "def": "suite_order()",
                        "request": {}
                    }
                },
                {
                    "test": {
                        "name": "add product to cart",
                        "api": "api_add_cart()",
                        "validate": []
                    }
                },
                {
                    "test": {
                        "name": "add product to cart",
                        "suite": "create_and_check()",
                        "validate": []
                    }
                },
                {
                    "test": {
                        "name": "checkout cart",
                        "request": {},
                        "validate": []
                    }
                }
            ]

    Returns:
        dict: testcase dict
            {
                "config": {},
                "teststeps": [teststep11, teststep12]
            }

    """
    testcase = {
        "config": {},
        "teststeps": []
    }

    for item in load_file(file_path):
        # TODO: add json schema validation
        if not isinstance(item, dict) or len(item) != 1:
            raise exceptions.FileFormatError("Testcase format error: {}".format(file_path))

        key, test_block = item.popitem()
        if not isinstance(test_block, dict):
            raise exceptions.FileFormatError("Testcase format error: {}".format(file_path))

        if key == "config":
            testcase["config"].update(test_block)

        elif key == "test":

            def extend_api_definition(block):
                ref_call = block["api"]
                def_block = _get_block_by_name(ref_call, "def-api")
                _extend_block(block, def_block)

            # reference api
            if "api" in test_block:
                extend_api_definition(test_block)
                testcase["teststeps"].append(test_block)

            # reference testcase
            elif "suite" in test_block: # TODO: replace suite with testcase
                ref_call = test_block["suite"]
                block = _get_block_by_name(ref_call, "def-testcase")
                # TODO: bugfix lost block config variables
                for teststep in block["teststeps"]:
                    if "api" in teststep:
                        extend_api_definition(teststep)
                    testcase["teststeps"].append(teststep)

            # define directly
            else:
                testcase["teststeps"].append(test_block)

        else:
            logger.log_warning(
                "unexpected block key: {}. block key should only be 'config' or 'test'.".format(key)
            )

    return testcase


def _get_block_by_name(ref_call, ref_type):
    """ get test content by reference name.

    Args:
        ref_call (str): call function.
            e.g. api_v1_Account_Login_POST($UserName, $Password)
        ref_type (enum): "def-api" or "def-testcase"

    Returns:
        dict: api/testcase definition.

    Raises:
        exceptions.ParamsError: call args number is not equal to defined args number.

    """
    function_meta = parser.parse_function(ref_call)
    func_name = function_meta["func_name"]
    call_args = function_meta["args"]
    block = _get_test_definition(func_name, ref_type)
    def_args = block.get("function_meta", {}).get("args", [])

    if len(call_args) != len(def_args):
        err_msg = "{}: call args number is not equal to defined args number!\n".format(func_name)
        err_msg += "defined args: {}\n".format(def_args)
        err_msg += "reference args: {}".format(call_args)
        logger.log_error(err_msg)
        raise exceptions.ParamsError(err_msg)

    args_mapping = {}
    for index, item in enumerate(def_args):
        if call_args[index] == item:
            continue

        args_mapping[item] = call_args[index]

    if args_mapping:
        block = parser.substitute_variables(block, args_mapping)

    return block


def _get_test_definition(name, ref_type):
    """ get expected api or testcase.

    Args:
        name (str): api or testcase name
        ref_type (enum): "def-api" or "def-testcase"

    Returns:
        dict: expected api/testcase info if found.

    Raises:
        exceptions.ApiNotFound: api not found
        exceptions.TestcaseNotFound: testcase not found

    """
    block = project_mapping.get(ref_type, {}).get(name)

    if not block:
        err_msg = "{} not found!".format(name)
        if ref_type == "def-api":
            raise exceptions.ApiNotFound(err_msg)
        else:
            # ref_type == "def-testcase":
            raise exceptions.TestcaseNotFound(err_msg)

    return block


def _extend_block(ref_block, def_block):
    """ extend ref_block with def_block.

    Args:
        def_block (dict): api definition dict.
        ref_block (dict): reference block

    Returns:
        dict: extended reference block.

    Examples:
        >>> def_block = {
                "name": "get token 1",
                "request": {...},
                "validate": [{'eq': ['status_code', 200]}]
            }
        >>> ref_block = {
                "name": "get token 2",
                "extract": [{"token": "content.token"}],
                "validate": [{'eq': ['status_code', 201]}, {'len_eq': ['content.token', 16]}]
            }
        >>> _extend_block(def_block, ref_block)
            {
                "name": "get token 2",
                "request": {...},
                "extract": [{"token": "content.token"}],
                "validate": [{'eq': ['status_code', 201]}, {'len_eq': ['content.token', 16]}]
            }

    """
    # TODO: override variables
    def_validators = def_block.get("validate") or def_block.get("validators", [])
    ref_validators = ref_block.get("validate") or ref_block.get("validators", [])

    def_extrators = def_block.get("extract") \
        or def_block.get("extractors") \
        or def_block.get("extract_binds", [])
    ref_extractors = ref_block.get("extract") \
        or ref_block.get("extractors") \
        or ref_block.get("extract_binds", [])

    ref_block.update(def_block)
    ref_block["validate"] = _merge_validator(
        def_validators,
        ref_validators
    )
    ref_block["extract"] = _merge_extractor(
        def_extrators,
        ref_extractors
    )


def _convert_validators_to_mapping(validators):
    """ convert validators list to mapping.

    Args:
        validators (list): validators in list

    Returns:
        dict: validators mapping, use (check, comparator) as key.

    Examples:
        >>> validators = [
                {"check": "v1", "expect": 201, "comparator": "eq"},
                {"check": {"b": 1}, "expect": 200, "comparator": "eq"}
            ]
        >>> _convert_validators_to_mapping(validators)
            {
                ("v1", "eq"): {"check": "v1", "expect": 201, "comparator": "eq"},
                ('{"b": 1}', "eq"): {"check": {"b": 1}, "expect": 200, "comparator": "eq"}
            }

    """
    validators_mapping = {}

    for validator in validators:
        validator = parser.parse_validator(validator)

        if not isinstance(validator["check"], collections.Hashable):
            check = json.dumps(validator["check"])
        else:
            check = validator["check"]

        key = (check, validator["comparator"])
        validators_mapping[key] = validator

    return validators_mapping


def _merge_validator(def_validators, ref_validators):
    """ merge def_validators with ref_validators.

    Args:
        def_validators (list):
        ref_validators (list):

    Returns:
        list: merged validators

    Examples:
        >>> def_validators = [{'eq': ['v1', 200]}, {"check": "s2", "expect": 16, "comparator": "len_eq"}]
        >>> ref_validators = [{"check": "v1", "expect": 201}, {'len_eq': ['s3', 12]}]
        >>> _merge_validator(def_validators, ref_validators)
            [
                {"check": "v1", "expect": 201, "comparator": "eq"},
                {"check": "s2", "expect": 16, "comparator": "len_eq"},
                {"check": "s3", "expect": 12, "comparator": "len_eq"}
            ]

    """
    if not def_validators:
        return ref_validators

    elif not ref_validators:
        return def_validators

    else:
        def_validators_mapping = _convert_validators_to_mapping(def_validators)
        ref_validators_mapping = _convert_validators_to_mapping(ref_validators)

        def_validators_mapping.update(ref_validators_mapping)
        return list(def_validators_mapping.values())


def _merge_extractor(def_extrators, ref_extractors):
    """ merge def_extrators with ref_extractors

    Args:
        def_extrators (list): [{"var1": "val1"}, {"var2": "val2"}]
        ref_extractors (list): [{"var1": "val111"}, {"var3": "val3"}]

    Returns:
        list: merged extractors

    Examples:
        >>> def_extrators = [{"var1": "val1"}, {"var2": "val2"}]
        >>> ref_extractors = [{"var1": "val111"}, {"var3": "val3"}]
        >>> _merge_extractor(def_extrators, ref_extractors)
            [
                {"var1": "val111"},
                {"var2": "val2"},
                {"var3": "val3"}
            ]

    """
    if not def_extrators:
        return ref_extractors

    elif not ref_extractors:
        return def_extrators

    else:
        extractor_dict = OrderedDict()
        for api_extrator in def_extrators:
            if len(api_extrator) != 1:
                logger.log_warning("incorrect extractor: {}".format(api_extrator))
                continue

            var_name = list(api_extrator.keys())[0]
            extractor_dict[var_name] = api_extrator[var_name]

        for test_extrator in ref_extractors:
            if len(test_extrator) != 1:
                logger.log_warning("incorrect extractor: {}".format(test_extrator))
                continue

            var_name = list(test_extrator.keys())[0]
            extractor_dict[var_name] = test_extrator[var_name]

        extractor_list = []
        for key, value in extractor_dict.items():
            extractor_list.append({key: value})

        return extractor_list


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
        for api_item in api_items:
            key, api_dict = api_item.popitem()

            api_def = api_dict.pop("def")
            function_meta = parser.parse_function(api_def)
            func_name = function_meta["func_name"]

            if func_name in api_definition_mapping:
                logger.log_warning("API definition duplicated: {}".format(func_name))

            api_dict["function_meta"] = function_meta
            api_definition_mapping[func_name] = api_dict

    project_mapping["def-api"] = api_definition_mapping
    return api_definition_mapping


def load_test_folder(test_folder_path):
    """ load testcases definitions from folder.

    Args:
        test_folder_path (str): testcases files folder.

            testcase file should be in the following format:
            [
                {
                    "config": {
                        "def": "create_and_check",
                        "request": {},
                        "validate": []
                    }
                },
                {
                    "test": {
                        "api": "get_user",
                        "validate": []
                    }
                }
            ]

    Returns:
        dict: testcases definition mapping.

            {
                "create_and_check": [
                    {"config": {}},
                    {"test": {}},
                    {"test": {}}
                ],
                "tests/testcases/create_and_get.yml": [
                    {"config": {}},
                    {"test": {}},
                    {"test": {}}
                ]
            }

    """
    test_definition_mapping = {}

    test_items_mapping = load_folder_content(test_folder_path)

    for test_file_path, items in test_items_mapping.items():
        # TODO: add JSON schema validation

        testcase = {
            "config": {},
            "teststeps": []
        }
        for item in items:
            key, block = item.popitem()

            if key == "config":
                testcase["config"].update(block)

                if "def" not in block:
                    test_definition_mapping[test_file_path] = testcase
                    continue

                testcase_def = block.pop("def")
                function_meta = parser.parse_function(testcase_def)
                func_name = function_meta["func_name"]

                if func_name in test_definition_mapping:
                    logger.log_warning("API definition duplicated: {}".format(func_name))

                testcase["function_meta"] = function_meta
                test_definition_mapping[func_name] = testcase
            else:
                # key == "test":
                testcase["teststeps"].append(block)

    project_mapping["def-testcase"] = test_definition_mapping
    return test_definition_mapping


def reset_loader():
    """ reset project mapping.
    """
    global project_working_directory
    project_working_directory = os.getcwd()

    project_mapping["debugtalk"] = {
        "variables": {},
        "functions": {}
    }
    project_mapping["env"] = {}
    project_mapping["def-api"] = {}
    project_mapping["def-testcase"] = {}
    testcases_cache_mapping.clear()


def locate_debugtalk_py(start_path):
    """ locate debugtalk.py file.

    Args:
        start_path (str): start locating path, maybe testcase file path or directory path

    """
    try:
        debugtalk_path = locate_file(start_path, "debugtalk.py")
        return os.path.abspath(debugtalk_path)
    except exceptions.FileNotFound:
        return None


def load_project_tests(test_path):
    """ load api, testcases, .env, builtin module and debugtalk.py.
        api/testcases folder is relative to project_working_directory

    Args:
        test_path (str): test file/folder path, locate pwd from this path.

    """
    global project_working_directory

    reset_loader()
    load_builtin_module()

    debugtalk_path = locate_debugtalk_py(test_path)
    if debugtalk_path:
        # The folder contains debugtalk.py will be treated as PWD.
        # add PWD to sys.path
        project_working_directory = os.path.dirname(debugtalk_path)

        # load debugtalk.py
        sys.path.insert(0, project_working_directory)
        load_debugtalk_module()
    else:
        # debugtalk.py not found, use os.getcwd() as PWD.
        project_working_directory = os.getcwd()

    load_dot_env_file()
    load_api_folder(os.path.join(project_working_directory, "api"))
    # TODO: replace suite with testcases
    load_test_folder(os.path.join(project_working_directory, "suite"))


def load_testcases(path):
    """ load testcases from file path, extend and merge with api/testcase definitions.

    Args:
        path (str): testcase file/foler path.
            path could be in several types:
                - absolute/relative file path
                - absolute/relative folder path
                - list/set container with file(s) and/or folder(s)

    Returns:
        list: testcases list, each testcase is corresponding to a file
        [
            testcase_dict_1,
            testcase_dict_2
        ]

    """
    if isinstance(path, (list, set)):
        testcases_list = []

        for file_path in set(path):
            testcases = load_testcases(file_path)
            if not testcases:
                continue
            testcases_list.extend(testcases)

        return testcases_list

    if not os.path.isabs(path):
        path = os.path.join(os.getcwd(), path)

    if path in testcases_cache_mapping:
        return testcases_cache_mapping[path]

    if os.path.isdir(path):
        load_project_tests(path)
        files_list = load_folder_files(path)
        testcases_list = load_testcases(files_list)

    elif os.path.isfile(path):
        try:
            load_project_tests(path)
            testcase = _load_test_file(path)
            if testcase["teststeps"]:
                testcases_list = [testcase]
            else:
                testcases_list = []
        except exceptions.FileFormatError:
            testcases_list = []

    else:
        err_msg = "path not exist: {}".format(path)
        logger.log_error(err_msg)
        raise exceptions.FileNotFound(err_msg)

    testcases_cache_mapping[path] = testcases_list
    return testcases_list
