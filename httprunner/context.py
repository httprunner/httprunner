import copy
import os
import re
import sys
from collections import OrderedDict

from httprunner import utils
from httprunner.testcase import TestcaseParser


class Context(object):
    """ Manages context functions and variables.
        context has two levels, testset and testcase.
    """
    def __init__(self):
        self.testset_shared_variables_mapping = OrderedDict()
        self.testcase_variables_mapping = OrderedDict()
        self.testcase_parser = TestcaseParser()
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
            variable_evale_value = self.testcase_parser.parse_content_with_bindings(value)

            if level == "testset":
                self.testset_shared_variables_mapping[variable_name] = variable_evale_value

            self.testcase_variables_mapping[variable_name] = variable_evale_value
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

    def get_parsed_request(self, request_dict, level="testcase"):
        """ get parsed request with bind variables and functions.
        @param request_dict: request config mapping
        @param level: testset or testcase
        """
        if level == "testset":
            request_dict = self.testcase_parser.parse_content_with_bindings(
                request_dict
            )
            self.testset_request_config.update(request_dict)

        testcase_request_config = utils.deep_update_dict(
            copy.deepcopy(self.testset_request_config),
            request_dict
        )
        parsed_request = self.testcase_parser.parse_content_with_bindings(
            testcase_request_config
        )

        return parsed_request

    def get_testcase_variables_mapping(self):
        return self.testcase_variables_mapping

    def exec_content_functions(self, content):
        """ execute functions in content.
        """
        self.testcase_parser.eval_content_functions(content)
