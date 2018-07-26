# encoding: utf-8

import copy
import os
import re
import sys

from httprunner import exceptions, logger, testcase, utils
from httprunner.compat import OrderedDict


class Context(object):
    """ Manages context functions and variables.
        context has two levels, testset and testcase.
    """
    def __init__(self):
        self.testset_shared_variables_mapping = OrderedDict()
        self.testcase_variables_mapping = OrderedDict()
        self.testcase_parser = testcase.TestcaseParser()
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
            self.import_module_items(["httprunner.built_in"], "testset")

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

    def import_module_items(self, modules, level="testcase"):
        """ import modules and bind all functions within the context
        """
        sys.path.insert(0, os.getcwd())
        for module_name in modules:
            imported_module = utils.get_imported_module(module_name)
            imported_functions_dict = utils.filter_module(imported_module, "function")
            self.__update_context_functions_config(level, imported_functions_dict)

            imported_variables_dict = utils.filter_module(imported_module, "variable")
            self.bind_variables(imported_variables_dict, level)

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
            or testcase.extract_variables(check_item) \
            or testcase.extract_functions(check_item):
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
                testcase.parse_validator(validator),
                resp_obj
            )

            try:
                self.do_validation(evaluated_validator)
            except exceptions.ValidationFailure:
                validate_pass = False

            self.evaluated_validators.append(evaluated_validator)

        if not validate_pass:
            raise exceptions.ValidationFailure
