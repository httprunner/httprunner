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
        raise exceptions.FileNotFound(".env file path is not exist.")

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
            if isinstance(item, tuple):
                continue
            debugtalk_module["variables"][name] = item
        else:
            pass

    return debugtalk_module


def load_builtin_module():
    """ load built_in module
    """
    built_in_module = load_python_module(built_in)
    return built_in_module


def load_debugtalk_module():
    """ load project debugtalk.py module
        debugtalk.py should be located in project working directory.

    Returns:
        dict: debugtalk module mapping
            {
                "variables": {},
                "functions": {}
            }

    """
    # load debugtalk.py module
    imported_module = importlib.import_module("debugtalk")
    debugtalk_module = load_python_module(imported_module)
    return debugtalk_module


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

def _load_teststeps(test_block, project_mapping):
    """ load teststeps with api/testcase references

    Args:
        test_block (dict): test block content, maybe in 3 formats.
            # api reference
            {
                "name": "add product to cart",
                "api": "api_add_cart()",
                "validate": []
            }
            # testcase reference
            {
                "name": "add product to cart",
                "suite": "create_and_check()",
                "validate": []
            }
            # define directly
            {
                "name": "checkout cart",
                "request": {},
                "validate": []
            }

    Returns:
        list: loaded teststeps list

    """
    def extend_api_definition(block):
        ref_call = block["api"]
        def_block = _get_block_by_name(ref_call, "def-api", project_mapping)
        _extend_block(block, def_block)

    teststeps = []

    # reference api
    if "api" in test_block:
        extend_api_definition(test_block)
        teststeps.append(test_block)

    # reference testcase
    elif "suite" in test_block: # TODO: replace suite with testcase
        ref_call = test_block["suite"]
        block = _get_block_by_name(ref_call, "def-testcase", project_mapping)
        # TODO: bugfix lost block config variables
        for teststep in block["teststeps"]:
            if "api" in teststep:
                extend_api_definition(teststep)
            teststeps.append(teststep)

    # define directly
    else:
        teststeps.append(test_block)

    return teststeps


def _load_testcase(raw_testcase, project_mapping):
    """ load testcase/testsuite with api/testcase references

    Args:
        raw_testcase (list): raw testcase content loaded from JSON/YAML file:
            [
                # config part
                {
                    "config": {
                        "name": "",
                        "def": "suite_order()",
                        "request": {}
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
        project_mapping (dict): project_mapping

    Returns:
        dict: loaded testcase content
            {
                "config": {},
                "teststeps": [teststep11, teststep12]
            }

    """
    loaded_testcase = {
        "config": {},
        "teststeps": []
    }

    for item in raw_testcase:
        # TODO: add json schema validation
        if not isinstance(item, dict) or len(item) != 1:
            raise exceptions.FileFormatError("Testcase format error: {}".format(item))

        key, test_block = item.popitem()
        if not isinstance(test_block, dict):
            raise exceptions.FileFormatError("Testcase format error: {}".format(item))

        if key == "config":
            loaded_testcase["config"].update(test_block)

        elif key == "test":
            loaded_testcase["teststeps"].extend(_load_teststeps(test_block, project_mapping))

        else:
            logger.log_warning(
                "unexpected block key: {}. block key should only be 'config' or 'test'.".format(key)
            )

    return loaded_testcase


def _get_block_by_name(ref_call, ref_type, project_mapping):
    """ get test content by reference name.

    Args:
        ref_call (str): call function.
            e.g. api_v1_Account_Login_POST($UserName, $Password)
        ref_type (enum): "def-api" or "def-testcase"
        project_mapping (dict): project_mapping

    Returns:
        dict: api/testcase definition.

    Raises:
        exceptions.ParamsError: call args number is not equal to defined args number.

    """
    function_meta = parser.parse_function(ref_call)
    func_name = function_meta["func_name"]
    call_args = function_meta["args"]
    block = _get_test_definition(func_name, ref_type, project_mapping)
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


def _get_test_definition(name, ref_type, project_mapping):
    """ get expected api or testcase.

    Args:
        name (str): api or testcase name
        ref_type (enum): "def-api" or "def-testcase"
        project_mapping (dict): project_mapping

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

    return test_definition_mapping


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


def load_project_tests(test_path, dot_env_path=None):
    """ load api, testcases, .env, builtin module and debugtalk.py.
        api/testcases folder is relative to project_working_directory

    Args:
        test_path (str): test file/folder path, locate pwd from this path.
        dot_env_path (str): specified .env file path

    Returns:
        dict: project loaded api/testcases definitions, environments and debugtalk.py module.

    """
    project_mapping = {}

    debugtalk_path = locate_debugtalk_py(test_path)
    # locate PWD with debugtalk.py path
    if debugtalk_path:
        # The folder contains debugtalk.py will be treated as PWD.
        project_working_directory = os.path.dirname(debugtalk_path)
    else:
        # debugtalk.py is not found, use os.getcwd() as PWD.
        project_working_directory = os.getcwd()

    # add PWD to sys.path
    sys.path.insert(0, project_working_directory)

    # load .env
    dot_env_path = dot_env_path or os.path.join(project_working_directory, ".env")
    if os.path.isfile(dot_env_path):
        project_mapping["env"] = load_dot_env_file(dot_env_path)
    else:
        project_mapping["env"] = {}

    # load debugtalk.py
    if debugtalk_path:
        project_mapping["debugtalk"] = load_debugtalk_module()
    else:
        project_mapping["debugtalk"] = {
            "variables": {},
            "functions": {}
        }

    project_mapping["def-api"] = load_api_folder(os.path.join(project_working_directory, "api"))
    # TODO: replace suite with testcases
    project_mapping["def-testcase"] = load_test_folder(os.path.join(project_working_directory, "suite"))

    return project_mapping


def load_tests(path, dot_env_path=None):
    """ load testcases from file path, extend and merge with api/testcase definitions.

    Args:
        path (str/list): testcase file/foler path.
            path could be in several types:
                - absolute/relative file path
                - absolute/relative folder path
                - list/set container with file(s) and/or folder(s)
        dot_env_path (str): specified .env file path

    Returns:
        list: testcases list, each testcase is corresponding to a file
        [
            {   # testcase data structure
                "config": {
                    "name": "desc1",
                    "path": "testcase1_path",
                    "variables": [],                    # optional
                    "request": {}                       # optional
                    "refs": {
                        "debugtalk": {
                            "variables": {},
                            "functions": {}
                        },
                        "env": {},
                        "def-api": {},
                        "def-testcase": {}
                    }
                },
                "teststeps": [
                    # teststep data structure
                    {
                        'name': 'test step desc1',
                        'variables': [],    # optional
                        'extract': [],      # optional
                        'validate': [],
                        'request': {},
                        'function_meta': {}
                    },
                    teststep2   # another teststep dict
                ]
            },
            testcase_dict_2     # another testcase dict
        ]

    """
    if isinstance(path, (list, set)):
        testcases_list = []

        for file_path in set(path):
            testcases = load_tests(file_path, dot_env_path)
            if not testcases:
                continue
            testcases_list.extend(testcases)

        return testcases_list

    if not os.path.exists(path):
        err_msg = "path not exist: {}".format(path)
        logger.log_error(err_msg)
        raise exceptions.FileNotFound(err_msg)

    if not os.path.isabs(path):
        path = os.path.join(os.getcwd(), path)

    if os.path.isdir(path):
        files_list = load_folder_files(path)
        testcases_list = load_tests(files_list, dot_env_path)

    elif os.path.isfile(path):
        try:
            raw_testcase = load_file(path)
            project_mapping = load_project_tests(path, dot_env_path)
            testcase = _load_testcase(raw_testcase, project_mapping)
            testcase["config"]["path"] = path
            testcase["config"]["refs"] = project_mapping
            testcases_list = [testcase]
        except exceptions.FileFormatError:
            testcases_list = []

    return testcases_list


def load_locust_tests(path, dot_env_path=None):
    """ load locust testcases

    Args:
        path (str): testcase/testsuite file path.
        dot_env_path (str): specified .env file path

    Returns:
        dict: locust testcases with weight
        {
            "config": {...},
            "tests": [
                # weight 3
                [teststep11],
                [teststep11],
                [teststep11],
                # weight 2
                [teststep21, teststep22],
                [teststep21, teststep22]
            ]
        }

    """
    raw_testcase = load_file(path)
    project_mapping = load_project_tests(path, dot_env_path)

    config = {
        "refs": project_mapping
    }
    tests = []
    for item in raw_testcase:
        key, test_block = item.popitem()

        if key == "config":
            config.update(test_block)
        elif key == "test":
            teststeps = _load_teststeps(test_block, project_mapping)
            weight = test_block.pop("weight", 1)
            for _ in range(weight):
                tests.append(teststeps)

    return {
        "config": config,
        "tests": tests
    }
