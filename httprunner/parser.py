# encoding: utf-8

import ast
import os
import random
import re

from httprunner import exceptions, loader, utils
from httprunner.compat import (OrderedDict, basestring, builtin_str,
                               numeric_types, str)

function_regexp = r"\$\{([\w_]+\([\$\w\.\-_ =,]*\))\}"
variable_regexp = r"\$([\w_]+)"
function_regexp_compile = re.compile(r"^([\w_]+)\(([\$\w\.\-_ =,]*)\)$")


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


def extract_variables(content):
    """ extract all variable names from content, which is in format $variable
    @param (str) content
    @return (list) variable name list

    e.g. $variable => ["variable"]
         /blog/$postid => ["postid"]
         /$var1/$var2 => ["var1", "var2"]
         abc => []
    """
    # TODO: change variable notation from $var to {{var}}
    try:
        return re.findall(variable_regexp, content)
    except TypeError:
        return []


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
        raise exceptions.FunctionNotFound("{} not found!".format(content))

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
        raise exceptions.ParamsError("invalid validator: {}".format(validator))

    if "check" in validator and len(validator) > 1:
        # format1
        check_item = validator.get("check")

        if "expect" in validator:
            expect_value = validator.get("expect")
        elif "expected" in validator:
            expect_value = validator.get("expected")
        else:
            raise exceptions.ParamsError("invalid validator: {}".format(validator))

        comparator = validator.get("comparator", "eq")

    elif len(validator) == 1:
        # format2
        comparator = list(validator.keys())[0]
        compare_values = validator[comparator]

        if not isinstance(compare_values, list) or len(compare_values) != 2:
            raise exceptions.ParamsError("invalid validator: {}".format(validator))

        check_item, expect_value = compare_values

    else:
        raise exceptions.ParamsError("invalid validator: {}".format(validator))

    return {
        "check": check_item,
        "expect": expect_value,
        "comparator": comparator
    }


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
                raise exceptions.ParamsError("parameters syntax error!")

            parameter_content_list = [
                # get subset by parameter name
                {key: parameter_item[key] for key in parameter_name_list}
                for parameter_item in parsed_parameter_content
            ]

        parsed_parameters_list.append(parameter_content_list)

    return utils.gen_cartesian_product(*parsed_parameters_list)


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
            raise exceptions.ParamsError("bind item should only be function or variable.")

        try:
            assert self.file_path is not None
            return utils.search_conf_item(self.file_path, item_type, item_name)
        except (AssertionError, exceptions.FunctionNotFound):
            raise exceptions.ParamsError(
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
        csv_content_list = loader.load_file(parameter_file_path)

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
                if not isinstance(variable_value, str):
                    variable_value = builtin_str(variable_value)

                content = content.replace(
                    "${}".format(variable_name),
                    variable_value, 1
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
