# encoding: utf-8

import ast
import collections
import io
import itertools
import json
import os
import random
import re

from httprunner import exception, logger, utils
from httprunner.compat import OrderedDict, basestring, numeric_types
from httprunner.utils import FileUtils

variable_regexp = r"\$([\w_]+)"
function_regexp = r"\$\{([\w_]+\([\$\w\.\-_ =,]*\))\}"
function_regexp_compile = re.compile(r"^([\w_]+)\(([\$\w\.\-_ =,]*)\)$")


def extract_variables(content):
    """ extract all variable names from content, which is in format $variable
    @param (str) content
    @return (list) variable name list

    e.g. $variable => ["variable"]
         /blog/$postid => ["postid"]
         /$var1/$var2 => ["var1", "var2"]
         abc => []
    """
    try:
        return re.findall(variable_regexp, content)
    except TypeError:
        return []

def extract_functions(content):
    """ extract all functions from string content, which are in format ${fun()}
    @param (str) content
    @return (list) functions list

    e.g. ${func(5)} => ["func(5)"]
         ${func(a=1, b=2)} => ["func(a=1, b=2)"]
         /api/1000?_t=${get_timestamp()} => ["get_timestamp()"]
         /api/${add(1, 2)} => ["add(1, 2)"]
         "/api/${add(1, 2)}?_t=${get_timestamp()}" => ["add(1, 2)", "get_timestamp()"]
    """
    try:
        return re.findall(function_regexp, content)
    except TypeError:
        return []

def parse_string_value(str_value):
    """ parse string to number if possible
    e.g. "123" => 123
         "12.2" => 12.3
         "abc" => "abc"
         "$var" => "$var"
    """
    try:
        return ast.literal_eval(str_value)
    except ValueError:
        return str_value
    except SyntaxError:
        # e.g. $var, ${func}
        return str_value

def parse_function(content):
    """ parse function name and args from string content.
    @param (str) content
    @return (dict) function name and args

    e.g. func() => {'func_name': 'func', 'args': [], 'kwargs': {}}
         func(5) => {'func_name': 'func', 'args': [5], 'kwargs': {}}
         func(1, 2) => {'func_name': 'func', 'args': [1, 2], 'kwargs': {}}
         func(a=1, b=2) => {'func_name': 'func', 'args': [], 'kwargs': {'a': 1, 'b': 2}}
         func(1, 2, a=3, b=4) => {'func_name': 'func', 'args': [1, 2], 'kwargs': {'a':3, 'b':4}}
    """
    matched = function_regexp_compile.match(content)
    if not matched:
        raise exception.FunctionNotFound("{} not found!".format(content))

    function_meta = {
        "func_name": matched.group(1),
        "args": [],
        "kwargs": {}
    }

    args_str = matched.group(2).strip()
    if args_str == "":
        return function_meta

    args_list = args_str.split(',')
    for arg in args_list:
        arg = arg.strip()
        if '=' in arg:
            key, value = arg.split('=')
            function_meta["kwargs"][key.strip()] = parse_string_value(value.strip())
        else:
            function_meta["args"].append(parse_string_value(arg))

    return function_meta


class TestcaseLoader(object):

    overall_def_dict = {
        "api": {},
        "suite": {}
    }
    testcases_cache_mapping = {}

    @staticmethod
    def load_test_dependencies():
        """ load all api and suite definitions.
            default api folder is "$CWD/tests/api/".
            default suite folder is "$CWD/tests/suite/".
        """
        # TODO: cache api and suite loading
        # load api definitions
        api_def_folder = os.path.join(os.getcwd(), "tests", "api")
        for test_file in FileUtils.load_folder_files(api_def_folder):
            TestcaseLoader.load_api_file(test_file)

        # load suite definitions
        suite_def_folder = os.path.join(os.getcwd(), "tests", "suite")
        for suite_file in FileUtils.load_folder_files(suite_def_folder):
            suite = TestcaseLoader.load_test_file(suite_file)
            if "def" not in suite["config"]:
                raise exception.ParamsError("def missed in suite file: {}!".format(suite_file))

            call_func = suite["config"]["def"]
            function_meta = parse_function(call_func)
            suite["function_meta"] = function_meta
            TestcaseLoader.overall_def_dict["suite"][function_meta["func_name"]] = suite

    @staticmethod
    def load_api_file(file_path):
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
        api_items = FileUtils.load_file(file_path)
        if not isinstance(api_items, list):
            raise exception.FileFormatError("API format error: {}".format(file_path))

        for api_item in api_items:
            if not isinstance(api_item, dict) or len(api_item) != 1:
                raise exception.FileFormatError("API format error: {}".format(file_path))

            key, api_dict = api_item.popitem()
            if key != "api" or not isinstance(api_dict, dict) or "def" not in api_dict:
                raise exception.FileFormatError("API format error: {}".format(file_path))

            api_def = api_dict.pop("def")
            function_meta = parse_function(api_def)
            func_name = function_meta["func_name"]

            if func_name in TestcaseLoader.overall_def_dict["api"]:
                logger.log_warning("API definition duplicated: {}".format(func_name))

            api_dict["function_meta"] = function_meta
            TestcaseLoader.overall_def_dict["api"][func_name] = api_dict

    @staticmethod
    def load_test_file(file_path):
        """ load testcase file or suite file
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
        for item in FileUtils.load_file(file_path):
            if not isinstance(item, dict) or len(item) != 1:
                raise exception.FileFormatError("Testcase format error: {}".format(file_path))

            key, test_block = item.popitem()
            if not isinstance(test_block, dict):
                raise exception.FileFormatError("Testcase format error: {}".format(file_path))

            if key == "config":
                testset["config"].update(test_block)

            elif key == "test":
                if "api" in test_block:
                    ref_call = test_block["api"]
                    def_block = TestcaseLoader._get_block_by_name(ref_call, "api")
                    TestcaseLoader._override_block(def_block, test_block)
                    testset["testcases"].append(test_block)
                elif "suite" in test_block:
                    ref_call = test_block["suite"]
                    block = TestcaseLoader._get_block_by_name(ref_call, "suite")
                    testset["testcases"].extend(block["testcases"])
                else:
                    testset["testcases"].append(test_block)

            else:
                logger.log_warning(
                    "unexpected block key: {}. block key should only be 'config' or 'test'.".format(key)
                )

        return testset

    @staticmethod
    def _get_block_by_name(ref_call, ref_type):
        """ get test content by reference name
        @params:
            ref_call: e.g. api_v1_Account_Login_POST($UserName, $Password)
            ref_type: "api" or "suite"
        """
        function_meta = parse_function(ref_call)
        func_name = function_meta["func_name"]
        call_args = function_meta["args"]
        block = TestcaseLoader._get_test_definition(func_name, ref_type)
        def_args = block.get("function_meta").get("args", [])

        if len(call_args) != len(def_args):
            raise exception.ParamsError("call args mismatch defined args!")

        args_mapping = {}
        for index, item in enumerate(def_args):
            if call_args[index] == item:
                continue

            args_mapping[item] = call_args[index]

        if args_mapping:
            block = substitute_variables_with_mapping(block, args_mapping)

        return block

    @staticmethod
    def _get_test_definition(name, ref_type):
        """ get expected api or suite.
        @params:
            name: api or suite name
            ref_type: "api" or "suite"
        @return
            expected api info if found, otherwise raise ApiNotFound exception
        """
        block = TestcaseLoader.overall_def_dict.get(ref_type, {}).get(name)

        if not block:
            err_msg = "{} not found!".format(name)
            if ref_type == "api":
                raise exception.ApiNotFound(err_msg)
            else:
                # ref_type == "suite":
                raise exception.SuiteNotFound(err_msg)

        return block

    @staticmethod
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

    @staticmethod
    def load_testsets_by_path(path):
        """ load testcases from file path
        @param path: path could be in several type
            - absolute/relative file path
            - absolute/relative folder path
            - list/set container with file(s) and/or folder(s)
        @return testcase sets list, each testset is corresponding to a file
            [
                testset_dict_1,
                testset_dict_2
            ]
        """
        if isinstance(path, (list, set)):
            testsets = []

            for file_path in set(path):
                testset = TestcaseLoader.load_testsets_by_path(file_path)
                if not testset:
                    continue
                testsets.extend(testset)

            return testsets

        if not os.path.isabs(path):
            path = os.path.join(os.getcwd(), path)

        if path in TestcaseLoader.testcases_cache_mapping:
            return TestcaseLoader.testcases_cache_mapping[path]

        if os.path.isdir(path):
            files_list = FileUtils.load_folder_files(path)
            testcases_list = TestcaseLoader.load_testsets_by_path(files_list)

        elif os.path.isfile(path):
            try:
                testset = TestcaseLoader.load_test_file(path)
                if testset["testcases"] or testset["api"]:
                    testcases_list = [testset]
                else:
                    testcases_list = []
            except exception.FileFormatError:
                testcases_list = []

        else:
            logger.log_error(u"file not found: {}".format(path))
            testcases_list = []

        TestcaseLoader.testcases_cache_mapping[path] = testcases_list
        return testcases_list

def parse_validator(validator):
    """ parse validator, validator maybe in two format
    @param (dict) validator
        format1: this is kept for compatiblity with the previous versions.
            {"check": "status_code", "comparator": "eq", "expect": 201}
            {"check": "$resp_body_success", "comparator": "eq", "expect": True}
        format2: recommended new version
            {'eq': ['status_code', 201]}
            {'eq': ['$resp_body_success', True]}
    @return (dict) validator info
        {
            "check": "status_code",
            "expect": 201,
            "comparator": "eq"
        }
    """
    if not isinstance(validator, dict):
        raise exception.ParamsError("invalid validator: {}".format(validator))

    if "check" in validator and len(validator) > 1:
        # format1
        check_item = validator.get("check")

        if "expect" in validator:
            expect_value = validator.get("expect")
        elif "expected" in validator:
            expect_value = validator.get("expected")
        else:
            raise exception.ParamsError("invalid validator: {}".format(validator))

        comparator = validator.get("comparator", "eq")

    elif len(validator) == 1:
        # format2
        comparator = list(validator.keys())[0]
        compare_values = validator[comparator]

        if not isinstance(compare_values, list) or len(compare_values) != 2:
            raise exception.ParamsError("invalid validator: {}".format(validator))

        check_item, expect_value = compare_values

    else:
        raise exception.ParamsError("invalid validator: {}".format(validator))

    return {
        "check": check_item,
        "expect": expect_value,
        "comparator": comparator
    }

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
        validator = parse_validator(validator)

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


def is_testset(data_structure):
    """ check if data_structure is a testset
    testset should always be in the following data structure:
        {
            "name": "desc1",
            "config": {},
            "api": {},
            "testcases": [testcase11, testcase12]
        }
    """
    if not isinstance(data_structure, dict):
        return False

    if "name" not in data_structure or "testcases" not in data_structure:
        return False

    if not isinstance(data_structure["testcases"], list):
        return False

    return True

def is_testsets(data_structure):
    """ check if data_structure is testset or testsets
    testsets should always be in the following data structure:
        testset_dict
        or
        [
            testset_dict_1,
            testset_dict_2
        ]
    """
    if not isinstance(data_structure, list):
        return is_testset(data_structure)

    for item in data_structure:
        if not is_testset(item):
            return False

    return True

def substitute_variables_with_mapping(content, mapping):
    """ substitute variables in content with mapping
    e.g.
    @params
        content = {
            'request': {
                'url': '/api/users/$uid',
                'headers': {'token': '$token'}
            }
        }
        mapping = {"$uid": 1000}
    @return
        {
            'request': {
                'url': '/api/users/1000',
                'headers': {'token': '$token'}
            }
        }
    """
    # TODO: refactor type check
    if isinstance(content, bool):
        return content

    if isinstance(content, (numeric_types, type)):
        return content

    if not content:
        return content

    if isinstance(content, (list, set, tuple)):
        return [
            substitute_variables_with_mapping(item, mapping)
            for item in content
        ]

    if isinstance(content, dict):
        substituted_data = {}
        for key, value in content.items():
            eval_key = substitute_variables_with_mapping(key, mapping)
            eval_value = substitute_variables_with_mapping(value, mapping)
            substituted_data[eval_key] = eval_value

        return substituted_data

    # content is in string format here
    for var, value in mapping.items():
        if content == var:
            # content is a variable
            content = value
        else:
            content = content.replace(var, str(value))

    return content

def gen_cartesian_product(*args):
    """ generate cartesian product for lists
    @param
        (list) args
            [{"a": 1}, {"a": 2}],
            [
                {"x": 111, "y": 112},
                {"x": 121, "y": 122}
            ]
    @return
        cartesian product in list
        [
            {'a': 1, 'x': 111, 'y': 112},
            {'a': 1, 'x': 121, 'y': 122},
            {'a': 2, 'x': 111, 'y': 112},
            {'a': 2, 'x': 121, 'y': 122}
        ]
    """
    if not args:
        return []
    elif len(args) == 1:
        return args[0]

    product_list = []
    for product_item_tuple in itertools.product(*args):
        product_item_dict = {}
        for item in product_item_tuple:
            product_item_dict.update(item)

        product_list.append(product_item_dict)

    return product_list

def parse_parameters(parameters, testset_path=None):
    """ parse parameters and generate cartesian product
    @params
        (list) parameters: parameter name and value in list
            parameter value may be in three types:
                (1) data list
                (2) call built-in parameterize function
                (3) call custom function in debugtalk.py
            e.g.
                [
                    {"user_agent": ["iOS/10.1", "iOS/10.2", "iOS/10.3"]},
                    {"username-password": "${parameterize(account.csv)}"},
                    {"app_version": "${gen_app_version()}"}
                ]
        (str) testset_path: testset file path, used for locating csv file and debugtalk.py
    @return cartesian product in list
    """
    testcase_parser = TestcaseParser(file_path=testset_path)

    parsed_parameters_list = []
    for parameter in parameters:
        parameter_name, parameter_content = list(parameter.items())[0]
        parameter_name_list = parameter_name.split("-")

        if isinstance(parameter_content, list):
            # (1) data list
            # e.g. {"app_version": ["2.8.5", "2.8.6"]}
            #       => [{"app_version": "2.8.5", "app_version": "2.8.6"}]
            # e.g. {"username-password": [["user1", "111111"], ["test2", "222222"]}
            #       => [{"username": "user1", "password": "111111"}, {"username": "user2", "password": "222222"}]
            parameter_content_list = []
            for parameter_item in parameter_content:
                if not isinstance(parameter_item, (list, tuple)):
                    # "2.8.5" => ["2.8.5"]
                    parameter_item = [parameter_item]

                # ["app_version"], ["2.8.5"] => {"app_version": "2.8.5"}
                # ["username", "password"], ["user1", "111111"] => {"username": "user1", "password": "111111"}
                parameter_content_dict = dict(zip(parameter_name_list, parameter_item))

                parameter_content_list.append(parameter_content_dict)
        else:
            # (2) & (3)
            parsed_parameter_content = testcase_parser.eval_content_with_bindings(parameter_content)
            # e.g. [{'app_version': '2.8.5'}, {'app_version': '2.8.6'}]
            # e.g. [{"username": "user1", "password": "111111"}, {"username": "user2", "password": "222222"}]
            if not isinstance(parsed_parameter_content, list):
                raise exception.ParamsError("parameters syntax error!")

            parameter_content_list = [
                # get subset by parameter name
                {key: parameter_item[key] for key in parameter_name_list}
                for parameter_item in parsed_parameter_content
            ]

        parsed_parameters_list.append(parameter_content_list)

    return gen_cartesian_product(*parsed_parameters_list)

class TestcaseParser(object):

    def __init__(self, variables={}, functions={}, file_path=None):
        self.update_binded_variables(variables)
        self.bind_functions(functions)
        self.file_path = file_path

    def update_binded_variables(self, variables):
        """ bind variables to current testcase parser
        @param (dict) variables, variables binds mapping
            {
                "authorization": "a83de0ff8d2e896dbd8efb81ba14e17d",
                "random": "A2dEx",
                "data": {"name": "user", "password": "123456"},
                "uuid": 1000
            }
        """
        self.variables = variables

    def bind_functions(self, functions):
        """ bind functions to current testcase parser
        @param (dict) functions, functions binds mapping
            {
                "add_two_nums": lambda a, b=1: a + b
            }
        """
        self.functions = functions

    def _get_bind_item(self, item_type, item_name):
        if item_type == "function":
            if item_name in self.functions:
                return self.functions[item_name]

            try:
                # check if builtin functions
                item_func = eval(item_name)
                if callable(item_func):
                    # is builtin function
                    return item_func
            except (NameError, TypeError):
                # is not builtin function, continue to search
                pass
        elif item_type == "variable":
            if item_name in self.variables:
                return self.variables[item_name]
        else:
            raise exception.ParamsError("bind item should only be function or variable.")

        try:
            assert self.file_path is not None
            return utils.search_conf_item(self.file_path, item_type, item_name)
        except (AssertionError, exception.FunctionNotFound):
            raise exception.ParamsError(
                "{} is not defined in bind {}s!".format(item_name, item_type))

    def get_bind_function(self, func_name):
        return self._get_bind_item("function", func_name)

    def get_bind_variable(self, variable_name):
        return self._get_bind_item("variable", variable_name)

    def parameterize(self, csv_file_name, fetch_method="Sequential"):
        parameter_file_path = os.path.join(
            os.path.dirname(self.file_path),
            "{}".format(csv_file_name)
        )
        csv_content_list = FileUtils.load_file(parameter_file_path)

        if fetch_method.lower() == "random":
            random.shuffle(csv_content_list)

        return csv_content_list

    def _eval_content_functions(self, content):
        functions_list = extract_functions(content)
        for func_content in functions_list:
            function_meta = parse_function(func_content)
            func_name = function_meta['func_name']

            args = function_meta.get('args', [])
            kwargs = function_meta.get('kwargs', {})
            args = self.eval_content_with_bindings(args)
            kwargs = self.eval_content_with_bindings(kwargs)

            if func_name in ["parameterize", "P"]:
                eval_value = self.parameterize(*args, **kwargs)
            else:
                func = self.get_bind_function(func_name)
                eval_value = func(*args, **kwargs)

            func_content = "${" + func_content + "}"
            if func_content == content:
                # content is a variable
                content = eval_value
            else:
                # content contains one or many variables
                content = content.replace(
                    func_content,
                    str(eval_value), 1
                )

        return content

    def _eval_content_variables(self, content):
        """ replace all variables of string content with mapping value.
        @param (str) content
        @return (str) parsed content

        e.g.
            variable_mapping = {
                "var_1": "abc",
                "var_2": "def"
            }
            $var_1 => "abc"
            $var_1#XYZ => "abc#XYZ"
            /$var_1/$var_2/var3 => "/abc/def/var3"
            ${func($var_1, $var_2, xyz)} => "${func(abc, def, xyz)}"
        """
        variables_list = extract_variables(content)
        for variable_name in variables_list:
            variable_value = self.get_bind_variable(variable_name)

            if "${}".format(variable_name) == content:
                # content is a variable
                content = variable_value
            else:
                # content contains one or several variables
                content = content.replace(
                    "${}".format(variable_name),
                    str(variable_value), 1
                )

        return content

    def eval_content_with_bindings(self, content):
        """ parse content recursively, each variable and function in content will be evaluated.

        @param (dict) content in any data structure
            {
                "url": "http://127.0.0.1:5000/api/users/$uid/${add_two_nums(1, 1)}",
                "method": "POST",
                "headers": {
                    "Content-Type": "application/json",
                    "authorization": "$authorization",
                    "random": "$random",
                    "sum": "${add_two_nums(1, 2)}"
                },
                "body": "$data"
            }
        @return (dict) parsed content with evaluated bind values
            {
                "url": "http://127.0.0.1:5000/api/users/1000/2",
                "method": "POST",
                "headers": {
                    "Content-Type": "application/json",
                    "authorization": "a83de0ff8d2e896dbd8efb81ba14e17d",
                    "random": "A2dEx",
                    "sum": 3
                },
                "body": {"name": "user", "password": "123456"}
            }
        """
        if content is None:
            return None

        if isinstance(content, (list, tuple)):
            return [
                self.eval_content_with_bindings(item)
                for item in content
            ]

        if isinstance(content, dict):
            evaluated_data = {}
            for key, value in content.items():
                eval_key = self.eval_content_with_bindings(key)
                eval_value = self.eval_content_with_bindings(value)
                evaluated_data[eval_key] = eval_value

            return evaluated_data

        if isinstance(content, basestring):

            # content is in string format here
            content = content.strip()

            # replace functions with evaluated value
            # Notice: _eval_content_functions must be called before _eval_content_variables
            content = self._eval_content_functions(content)

            # replace variables with binding value
            content = self._eval_content_variables(content)

        return content
