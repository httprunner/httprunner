import collections
import csv
import importlib
import io
import json
import os

import yaml
from httprunner import exceptions, logger, parser, validator
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
    """ load folder path, return all files in list format.
    @param
        folder_path: specified folder path to load
        recursive: if True, will load files recursively
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


def load_dot_env_file(path):
    """ load .env file
    """
    if not path:
        path = os.path.join(os.getcwd(), ".env")
        if not os.path.isfile(path):
            logger.log_debug(".env file not exist: {}".format(path))
            return {}
    else:
        if not os.path.isfile(path):
            raise exceptions.FileNotFound("env file not exist: {}".format(path))

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
        if os.path.isabs(file_path):
            file_path = file_path[len(os.getcwd())+1:]

        return file_path

    # current working directory
    if os.path.abspath(start_dir_path) == os.getcwd():
        raise exceptions.FileNotFound("{} not found in {}".format(file_name, start_path))

    # locate recursive upward
    return locate_file(os.path.dirname(start_dir_path), file_name)


###############################################################################
##   debugtalk.py module loader
###############################################################################

def convert_module_name(python_file_path):
    """ convert python file relative path to module name.

    Args:
        python_file_path (str): python file relative path

    Returns:
        str: module name

    Examples:
        >>> convert_module_name("debugtalk.py")
        debugtalk

        >>> convert_module_name("tests/debugtalk.py")
        tests.debugtalk

        >>> convert_module_name("tests/data/debugtalk.py")
        tests.data.debugtalk

    """
    module_name = python_file_path.replace("/", ".").rstrip(".py")
    return module_name


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


def load_debugtalk_module(start_path=None):
    """ load debugtalk.py module.

    Args:
        start_path (str, optional): start locating path, maybe file path or directory path.
            Defaults to current working directory.

    Returns:
        dict: variables and functions mapping for debugtalk.py

            {
                "variables": {},
                "functions": {}
            }

    """
    start_path = start_path or os.getcwd()

    try:
        module_path = locate_file(start_path, "debugtalk.py")
        module_name = convert_module_name(module_path)
    except exceptions.FileNotFound:
        return {
            "variables": {},
            "functions": {}
        }

    imported_module = importlib.import_module(module_name)
    return load_python_module(imported_module)


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
##   suite loader
###############################################################################


overall_def_dict = {
    "api": {},
    "suite": {}
}
testcases_cache_mapping = {}


def _load_test_dependencies():
    """ load all api and suite definitions.
        default api folder is "$CWD/tests/api/".
        default suite folder is "$CWD/tests/suite/".
    """
    # TODO: cache api and suite loading
    # load api definitions
    api_def_folder = os.path.join(os.getcwd(), "tests", "api")
    for test_file in load_folder_files(api_def_folder):
        _load_api_file(test_file)

    # load suite definitions
    suite_def_folder = os.path.join(os.getcwd(), "tests", "suite")
    for suite_file in load_folder_files(suite_def_folder):
        suite = _load_test_file(suite_file)
        if "def" not in suite["config"]:
            raise exceptions.ParamsError("def missed in suite file: {}!".format(suite_file))

        call_func = suite["config"]["def"]
        function_meta = parser.parse_function(call_func)
        suite["function_meta"] = function_meta
        overall_def_dict["suite"][function_meta["func_name"]] = suite


def _load_api_file(file_path):
    """ load api definition from file and store in overall_def_dict["api"]
        api file should be in format below:
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
    """
    api_items = load_file(file_path)
    if not isinstance(api_items, list):
        raise exceptions.FileFormatError("API format error: {}".format(file_path))

    for api_item in api_items:
        if not isinstance(api_item, dict) or len(api_item) != 1:
            raise exceptions.FileFormatError("API format error: {}".format(file_path))

        key, api_dict = api_item.popitem()
        if key != "api" or not isinstance(api_dict, dict) or "def" not in api_dict:
            raise exceptions.FileFormatError("API format error: {}".format(file_path))

        api_def = api_dict.pop("def")
        function_meta = parser.parse_function(api_def)
        func_name = function_meta["func_name"]

        if func_name in overall_def_dict["api"]:
            logger.log_warning("API definition duplicated: {}".format(func_name))

        api_dict["function_meta"] = function_meta
        overall_def_dict["api"][func_name] = api_dict


def _load_test_file(file_path):
    """ load testcase file or testsuite file
    @param file_path: absolute valid file path
        file_path should be in format below:
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
                        "name": "checkout cart",
                        "request": {},
                        "validate": []
                    }
                }
            ]
    @return testset dict
        {
            "config": {},
            "testcases": [testcase11, testcase12]
        }
    """
    testset = {
        "config": {
            "path": file_path
        },
        "testcases": []     # TODO: rename to tests
    }
    for item in load_file(file_path):
        if not isinstance(item, dict) or len(item) != 1:
            raise exceptions.FileFormatError("Testcase format error: {}".format(file_path))

        key, test_block = item.popitem()
        if not isinstance(test_block, dict):
            raise exceptions.FileFormatError("Testcase format error: {}".format(file_path))

        if key == "config":
            testset["config"].update(test_block)

        elif key == "test":
            if "api" in test_block:
                ref_call = test_block["api"]
                def_block = _get_block_by_name(ref_call, "api")
                _override_block(def_block, test_block)
                testset["testcases"].append(test_block)
            elif "suite" in test_block:
                ref_call = test_block["suite"]
                block = _get_block_by_name(ref_call, "suite")
                testset["testcases"].extend(block["testcases"])
            else:
                testset["testcases"].append(test_block)

        else:
            logger.log_warning(
                "unexpected block key: {}. block key should only be 'config' or 'test'.".format(key)
            )

    return testset


def _get_block_by_name(ref_call, ref_type):
    """ get test content by reference name
    @params:
        ref_call: e.g. api_v1_Account_Login_POST($UserName, $Password)
        ref_type: "api" or "suite"
    """
    function_meta = parser.parse_function(ref_call)
    func_name = function_meta["func_name"]
    call_args = function_meta["args"]
    block = _get_test_definition(func_name, ref_type)
    def_args = block.get("function_meta").get("args", [])

    if len(call_args) != len(def_args):
        raise exceptions.ParamsError("call args mismatch defined args!")

    args_mapping = {}
    for index, item in enumerate(def_args):
        if call_args[index] == item:
            continue

        args_mapping[item] = call_args[index]

    if args_mapping:
        block = parser.parse_data(block, args_mapping)

    return block


def _get_test_definition(name, ref_type):
    """ get expected api or testcase.
    @params:
        name: api or testcase name
        ref_type: "api" or "suite"
    @return
        expected api info if found, otherwise raise ApiNotFound exception
    """
    block = overall_def_dict.get(ref_type, {}).get(name)

    if not block:
        err_msg = "{} not found!".format(name)
        if ref_type == "api":
            raise exceptions.ApiNotFound(err_msg)
        else:
            # ref_type == "suite":
            raise exceptions.TestcaseNotFound(err_msg)

    return block


def _override_block(def_block, current_block):
    """ override def_block with current_block
    @param def_block:
        {
            "name": "get token",
            "request": {...},
            "validate": [{'eq': ['status_code', 200]}]
        }
    @param current_block:
        {
            "name": "get token",
            "extract": [{"token": "content.token"}],
            "validate": [{'eq': ['status_code', 201]}, {'len_eq': ['content.token', 16]}]
        }
    @return
        {
            "name": "get token",
            "request": {...},
            "extract": [{"token": "content.token"}],
            "validate": [{'eq': ['status_code', 201]}, {'len_eq': ['content.token', 16]}]
        }
    """
    def_validators = def_block.get("validate") or def_block.get("validators", [])
    current_validators = current_block.get("validate") or current_block.get("validators", [])

    def_extrators = def_block.get("extract") \
        or def_block.get("extractors") \
        or def_block.get("extract_binds", [])
    current_extractors = current_block.get("extract") \
        or current_block.get("extractors") \
        or current_block.get("extract_binds", [])

    current_block.update(def_block)
    current_block["validate"] = _merge_validator(
        def_validators,
        current_validators
    )
    current_block["extract"] = _merge_extractor(
        def_extrators,
        current_extractors
    )


def _get_validators_mapping(validators):
    """ get validators mapping from api or test validators
    @param (list) validators:
        [
            {"check": "v1", "expect": 201, "comparator": "eq"},
            {"check": {"b": 1}, "expect": 200, "comparator": "eq"}
        ]
    @return
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


def _merge_validator(def_validators, current_validators):
    """ merge def_validators with current_validators
    @params:
        def_validators: [{'eq': ['v1', 200]}, {"check": "s2", "expect": 16, "comparator": "len_eq"}]
        current_validators: [{"check": "v1", "expect": 201}, {'len_eq': ['s3', 12]}]
    @return:
        [
            {"check": "v1", "expect": 201, "comparator": "eq"},
            {"check": "s2", "expect": 16, "comparator": "len_eq"},
            {"check": "s3", "expect": 12, "comparator": "len_eq"}
        ]
    """
    if not def_validators:
        return current_validators

    elif not current_validators:
        return def_validators

    else:
        api_validators_mapping = _get_validators_mapping(def_validators)
        test_validators_mapping = _get_validators_mapping(current_validators)

        api_validators_mapping.update(test_validators_mapping)
        return list(api_validators_mapping.values())


def _merge_extractor(def_extrators, current_extractors):
    """ merge def_extrators with current_extractors
    @params:
        def_extrators: [{"var1": "val1"}, {"var2": "val2"}]
        current_extractors: [{"var1": "val111"}, {"var3": "val3"}]
    @return:
        [
            {"var1": "val111"},
            {"var2": "val2"},
            {"var3": "val3"}
        ]
    """
    if not def_extrators:
        return current_extractors

    elif not current_extractors:
        return def_extrators

    else:
        extractor_dict = OrderedDict()
        for api_extrator in def_extrators:
            if len(api_extrator) != 1:
                logger.log_warning("incorrect extractor: {}".format(api_extrator))
                continue

            var_name = list(api_extrator.keys())[0]
            extractor_dict[var_name] = api_extrator[var_name]

        for test_extrator in current_extractors:
            if len(test_extrator) != 1:
                logger.log_warning("incorrect extractor: {}".format(test_extrator))
                continue

            var_name = list(test_extrator.keys())[0]
            extractor_dict[var_name] = test_extrator[var_name]

        extractor_list = []
        for key, value in extractor_dict.items():
            extractor_list.append({key: value})

        return extractor_list


def load_testcases(path):
    """ load testcases from file path
    @param path: path could be in several type
        - absolute/relative file path
        - absolute/relative folder path
        - list/set container with file(s) and/or folder(s)
    @return testcases list, each testcase is corresponding to a file
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
        files_list = load_folder_files(path)
        testcases_list = load_testcases(files_list)

    elif os.path.isfile(path):
        try:
            testcase = _load_test_file(path)
            if testcase["testcases"]:
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


def load(path):
    """ main interface for loading testcases
    @param (str) path: testcase file/folder path
    @return (list) testcases list
    """
    if validator.is_testcases(path):
        return path

    _load_test_dependencies()
    return load_testcases(path)
