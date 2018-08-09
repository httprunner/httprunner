# encoding: utf-8

import copy
import os
import random
import re
import sys

from httprunner import built_in, exceptions, loader, logger, parser, utils
from httprunner.compat import OrderedDict, basestring, builtin_str, str


def parse_parameters(parameters, testset_path=None):
    """ parse parameters and generate cartesian product.

    Args:
        parameters (list) parameters: parameter name and value in list
            parameter value may be in three types:
                (1) data list, e.g. ["iOS/10.1", "iOS/10.2", "iOS/10.3"]
                (2) call built-in parameterize function, "${parameterize(account.csv)}"
                (3) call custom function in debugtalk.py, "${gen_app_version()}"

        testset_path (str): testset file path, used for locating csv file and debugtalk.py

    Returns:
        list: cartesian product list

    Examples:
        >>> parameters = [
            {"user_agent": ["iOS/10.1", "iOS/10.2", "iOS/10.3"]},
            {"username-password": "${parameterize(account.csv)}"},
            {"app_version": "${gen_app_version()}"}
        ]
        >>> parse_parameters(parameters)

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
        """ get specified function or variable.

        Args:
            item_type(str): functions or variables
            item_name(str): function name or variable name

        Returns:
            object: specified function or variable object.
        """
        if item_type == "functions":
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
        else:
            # item_type == "variables":
            if item_name in self.variables:
                return self.variables[item_name]

        debugtalk_module = loader.load_debugtalk_module(self.file_path)
        return loader.get_module_item(debugtalk_module, item_type, item_name)

    def get_bind_function(self, func_name):
        return self._get_bind_item("functions", func_name)

    def get_bind_variable(self, variable_name):
        return self._get_bind_item("variables", variable_name)

    def load_csv_list(self, csv_file_name, fetch_method="Sequential"):
        """ locate csv file and load csv content.

        Args:
            csv_file_name (str): csv file name
            fetch_method (str): fetch data method, defaults to Sequential.
                If set to "random", csv data list will be reordered in random.

        Returns:
            list: csv data list
        """
        csv_file_path = loader.locate_file(self.file_path, csv_file_name)
        csv_content_list = loader.load_file(csv_file_path)

        if fetch_method.lower() == "random":
            random.shuffle(csv_content_list)

        return csv_content_list

    def _eval_content_functions(self, content):
        functions_list = parser.extract_functions(content)
        for func_content in functions_list:
            function_meta = parser.parse_function(func_content)
            func_name = function_meta['func_name']

            args = function_meta.get('args', [])
            kwargs = function_meta.get('kwargs', {})
            args = self.eval_content_with_bindings(args)
            kwargs = self.eval_content_with_bindings(kwargs)

            if func_name in ["parameterize", "P"]:
                eval_value = self.load_csv_list(*args, **kwargs)
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
        variables_list = parser.extract_variables(content)
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


class Context(object):
    """ Manages context functions and variables.
        context has two levels, testset and testcase.
    """
    def __init__(self):
        self.testset_shared_variables_mapping = OrderedDict()
        self.testcase_variables_mapping = OrderedDict()
        self.testcase_parser = TestcaseParser()
        self.evaluated_validators = []
        self.init_context()

    def init_context(self, level='testset'):
        """
        testset level context initializes when a file is loaded,
        testcase level context initializes when each testcase starts.
        """
        if level == "testset":
            self.testset_functions_config = {}
            self.testset_request_config = {}
            self.testset_shared_variables_mapping = OrderedDict()

        # testcase config shall inherit from testset configs,
        # but can not change testset configs, that's why we use copy.deepcopy here.
        self.testcase_functions_config = copy.deepcopy(self.testset_functions_config)
        self.testcase_variables_mapping = copy.deepcopy(self.testset_shared_variables_mapping)

        self.testcase_parser.bind_functions(self.testcase_functions_config)
        self.testcase_parser.update_binded_variables(self.testcase_variables_mapping)

        if level == "testset":
            self.import_module_items(built_in)

    def config_context(self, config_dict, level):
        if level == "testset":
            self.testcase_parser.file_path = config_dict.get("path", None)

        variables = config_dict.get('variables') \
            or config_dict.get('variable_binds', OrderedDict())
        self.bind_variables(variables, level)

    def bind_functions(self, function_binds, level="testcase"):
        """ Bind named functions within the context
            This allows for passing in self-defined functions in testing.
            e.g. function_binds:
            {
                "add_one": lambda x: x + 1,             # lambda function
                "add_two_nums": "lambda x, y: x + y"    # lambda function in string
            }
        """
        eval_function_binds = {}
        for func_name, function in function_binds.items():
            if isinstance(function, str):
                function = eval(function)
            eval_function_binds[func_name] = function

        self.__update_context_functions_config(level, eval_function_binds)

    def import_module_items(self, imported_module):
        """ import module functions and variables and bind to testset context
        """
        module_mapping = loader.load_python_module(imported_module)
        self.__update_context_functions_config("testset", module_mapping["functions"])
        self.bind_variables(module_mapping["variables"], "testset")

    def bind_variables(self, variables, level="testcase"):
        """ bind variables to testset context or current testcase context.
            variables in testset context can be used in all testcases of current test suite.

        @param (list or OrderDict) variables, variable can be value or custom function.
            if value is function, it will be called and bind result to variable.
        e.g.
            OrderDict({
                "TOKEN": "debugtalk",
                "random": "${gen_random_string(5)}",
                "json": {'name': 'user', 'password': '123456'},
                "md5": "${gen_md5($TOKEN, $json, $random)}"
            })
        """
        if isinstance(variables, list):
            variables = utils.convert_to_order_dict(variables)

        for variable_name, value in variables.items():
            variable_eval_value = self.eval_content(value)

            if level == "testset":
                self.testset_shared_variables_mapping[variable_name] = variable_eval_value

            self.bind_testcase_variable(variable_name, variable_eval_value)

    def bind_testcase_variable(self, variable_name, variable_value):
        """ bind and update testcase variables mapping
        """
        self.testcase_variables_mapping[variable_name] = variable_value
        self.testcase_parser.update_binded_variables(self.testcase_variables_mapping)

    def bind_extracted_variables(self, variables):
        """ bind extracted variables to testset context
        @param (OrderDict) variables
            extracted value do not need to evaluate.
        """
        for variable_name, value in variables.items():
            self.testset_shared_variables_mapping[variable_name] = value
            self.bind_testcase_variable(variable_name, value)

    def __update_context_functions_config(self, level, config_mapping):
        """
        @param level: testset or testcase
        @param config_type: functions
        @param config_mapping: functions config mapping
        """
        if level == "testset":
            self.testset_functions_config.update(config_mapping)

        self.testcase_functions_config.update(config_mapping)
        self.testcase_parser.bind_functions(self.testcase_functions_config)

    def eval_content(self, content):
        """ evaluate content recursively, take effect on each variable and function in content.
            content may be in any data structure, include dict, list, tuple, number, string, etc.
        """
        return self.testcase_parser.eval_content_with_bindings(content)

    def get_parsed_request(self, request_dict, level="testcase"):
        """ get parsed request with bind variables and functions.
        @param request_dict: request config mapping
        @param level: testset or testcase
        """
        if level == "testset":
            request_dict = self.eval_content(
                request_dict
            )
            self.testset_request_config.update(request_dict)

        testcase_request_config = utils.deep_update_dict(
            copy.deepcopy(self.testset_request_config),
            request_dict
        )
        parsed_request = self.eval_content(
            testcase_request_config
        )

        return parsed_request

    def eval_check_item(self, validator, resp_obj):
        """ evaluate check item in validator
        @param (dict) validator
            {"check": "status_code", "comparator": "eq", "expect": 201}
            {"check": "$resp_body_success", "comparator": "eq", "expect": True}
        @param (object) resp_obj
        @return (dict) validator info
            {
                "check": "status_code",
                "check_value": 200,
                "expect": 201,
                "comparator": "eq"
            }
        """
        check_item = validator["check"]
        # check_item should only be the following 5 formats:
        # 1, variable reference, e.g. $token
        # 2, function reference, e.g. ${is_status_code_200($status_code)}
        # 3, dict or list, maybe containing variable/function reference, e.g. {"var": "$abc"}
        # 4, string joined by delimiter. e.g. "status_code", "headers.content-type"
        # 5, regex string, e.g. "LB[\d]*(.*)RB[\d]*"

        if isinstance(check_item, (dict, list)) \
            or parser.extract_variables(check_item) \
            or parser.extract_functions(check_item):
            # format 1/2/3
            check_value = self.eval_content(check_item)
        else:
            # format 4/5
            check_value = resp_obj.extract_field(check_item)

        validator["check_value"] = check_value

        # expect_value should only be in 2 types:
        # 1, variable reference, e.g. $expect_status_code
        # 2, actual value, e.g. 200
        expect_value = self.eval_content(validator["expect"])
        validator["expect"] = expect_value
        validator["check_result"] = "unchecked"
        return validator

    def do_validation(self, validator_dict):
        """ validate with functions
        """
        # TODO: move comparator uniform to init_test_suites
        comparator = utils.get_uniform_comparator(validator_dict["comparator"])
        validate_func = self.testcase_parser.get_bind_function(comparator)

        if not validate_func:
            raise exceptions.FunctionNotFound("comparator not found: {}".format(comparator))

        check_item = validator_dict["check"]
        check_value = validator_dict["check_value"]
        expect_value = validator_dict["expect"]

        if (check_value is None or expect_value is None) \
            and comparator not in ["is", "eq", "equals", "=="]:
            raise exceptions.ParamsError("Null value can only be compared with comparator: eq/equals/==")

        validate_msg = "validate: {} {} {}({})".format(
            check_item,
            comparator,
            expect_value,
            type(expect_value).__name__
        )

        try:
            validator_dict["check_result"] = "pass"
            validate_func(check_value, expect_value)
            validate_msg += "\t==> pass"
            logger.log_debug(validate_msg)
        except (AssertionError, TypeError):
            validate_msg += "\t==> fail"
            validate_msg += "\n{}({}) {} {}({})".format(
                check_value,
                type(check_value).__name__,
                comparator,
                expect_value,
                type(expect_value).__name__
            )
            logger.log_error(validate_msg)
            validator_dict["check_result"] = "fail"
            raise exceptions.ValidationFailure(validate_msg)

    def validate(self, validators, resp_obj):
        """ make validations
        """
        if not validators:
            return

        logger.log_info("start to validate.")
        self.evaluated_validators = []
        validate_pass = True

        for validator in validators:
            # evaluate validators with context variable mapping.
            evaluated_validator = self.eval_check_item(
                parser.parse_validator(validator),
                resp_obj
            )

            try:
                self.do_validation(evaluated_validator)
            except exceptions.ValidationFailure:
                validate_pass = False

            self.evaluated_validators.append(evaluated_validator)

        if not validate_pass:
            raise exceptions.ValidationFailure
