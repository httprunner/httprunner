import copy
import os
import re
import sys
from collections import OrderedDict

from httprunner import exception, testcase, utils


class Context(object):
    """ Manages context functions and variables.
        context has two levels, testset and testcase.
    """
    def __init__(self):
        self.testset_shared_variables_mapping = OrderedDict()
        self.testcase_variables_mapping = OrderedDict()
        self.testcase_parser = testcase.TestcaseParser()
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

        requires = config_dict.get('requires', [])
        self.import_requires(requires)

        function_binds = config_dict.get('function_binds', {})
        self.bind_functions(function_binds, level)

        # import_module_functions will be deprecated soon
        module_items = config_dict.get('import_module_items', []) \
            or config_dict.get('import_module_functions', [])
        self.import_module_items(module_items, level)

        variables = config_dict.get('variables') \
            or config_dict.get('variable_binds', OrderedDict())
        self.bind_variables(variables, level)

    def import_requires(self, modules):
        """ import required modules dynamically
        """
        for module_name in modules:
            globals()[module_name] = utils.get_imported_module(module_name)

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

            self.testcase_variables_mapping[variable_name] = variable_eval_value
            self.testcase_parser.update_binded_variables(self.testcase_variables_mapping)

    def bind_extracted_variables(self, variables):
        """ bind extracted variables to testset context
        @param (OrderDict) variables
            extracted value do not need to evaluate.
        """
        for variable_name, value in variables.items():
            self.testset_shared_variables_mapping[variable_name] = value
            self.testcase_variables_mapping[variable_name] = value
            self.testcase_parser.update_binded_variables(self.testcase_variables_mapping)

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
        # check_item should only be in 3 types:
        # 1, variable reference, e.g. $token
        # 2, string joined by delimiter. e.g. "status_code", "headers.content-type"
        # 3, regex string, e.g. "LB[\d]*(.*)RB[\d]*"
        if testcase.extract_variables(check_item):
            # type 1
            check_value = self.eval_content(check_item)
        else:
            try:
                # type 2 or type 3
                check_value = resp_obj.extract_field(check_item)
            except exception.ParseResponseError:
                raise exception.ParseResponseError("failed to extract check item in response!")

        validator["check_value"] = check_value

        # expect_value should only be in 2 types:
        # 1, variable reference, e.g. $expect_status_code
        # 2, actual value, e.g. 200
        expect_value = self.eval_content(validator["expect"])
        validator["expect"] = expect_value
        return validator

    def do_validation(self, validator_dict):
        """ validate with functions
        """
        comparator = utils.get_uniform_comparator(validator_dict["comparator"])
        validate_func = self.testcase_parser.get_bind_item("function", comparator)

        if not validate_func:
            raise exception.FunctionNotFound("comparator not found: {}".format(comparator))

        check_item = validator_dict["check"]
        check_value = validator_dict["check_value"]
        expect_value = validator_dict["expect"]

        if (check_value is None or expect_value is None) \
            and comparator not in ["is", "eq", "equals", "=="]:
            raise exception.ParamsError("Null value can only be compared with comparator: eq/equals/==")

        try:
            validate_func(validator_dict["check_value"], validator_dict["expect"])
        except (AssertionError, TypeError):
            err_msg = "\n" + "\n".join([
                "\tcheck item name: %s;" % check_item,
                "\tcheck item value: %s (%s);" % (check_value, type(check_value).__name__),
                "\tcomparator: %s;" % comparator,
                "\texpected value: %s (%s)." % (expect_value, type(expect_value).__name__)
            ])
            raise exception.ValidationError(err_msg)

    def validate(self, validators, resp_obj):
        """ check validators with the context variable mapping.
        @param (list) validators
        @param (object) resp_obj
        """
        for validator in validators:
            validator_dict = self.eval_check_item(
                testcase.parse_validator(validator),
                resp_obj
            )
            self.do_validation(validator_dict)

        return True
