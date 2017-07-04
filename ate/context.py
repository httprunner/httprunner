import copy
import importlib
import re
import types
from collections import OrderedDict

from ate import exception, testcase, utils


def is_function(tup):
    """ Takes (name, object) tuple, returns True if it is a function.
    """
    name, item = tup
    return isinstance(item, types.FunctionType)

class Context(object):
    """ Manages context functions and variables.
        context has two levels, testset and testcase.
    """
    def __init__(self):
        self.testset_config = {}
        self.testset_shared_variables_mapping = dict()

        self.testcase_config = {}
        self.testcase_variables_mapping = dict()
        self.init_context()

    def init_context(self, level='testset'):
        """
        testset level context initializes when a file is loaded,
        testcase level context initializes when each testcase starts.
        """
        if level == "testset":
            self.testset_config["functions"] = {}
            self.testset_config["variables"] = OrderedDict()
            self.testset_config["request"] = {}
            self.testset_shared_variables_mapping = {}

        self.testcase_config["functions"] = {}
        self.testcase_config["variables"] = OrderedDict()
        self.testcase_config["request"] = {}
        self.testcase_variables_mapping = copy.deepcopy(self.testset_shared_variables_mapping)

    def import_requires(self, modules):
        """ import required modules dynamicly
        """
        for module_name in modules:
            globals()[module_name] = importlib.import_module(module_name)

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

        self.__update_context_config(level, "functions", eval_function_binds)

    def import_module_functions(self, modules, level="testcase"):
        """ import modules and bind all functions within the context
        """
        for module_name in modules:
            imported = importlib.import_module(module_name)
            imported_functions_dict = dict(filter(is_function, vars(imported).items()))
            self.__update_context_config(level, "functions", imported_functions_dict)

    def register_variables_config(self, variable_binds, level="testcase"):
        """ register variable configs
        @param (list) variable_binds, variable can be value or custom function
        e.g.
            [
                {"TOKEN": "debugtalk"},
                {"random": "${gen_random_string(5)}"},
                {"json": {'name': 'user', 'password': '123456'}},
                {"md5": "${gen_md5($TOKEN, $json, $random)}"}
            ]
        """
        if level == "testset":
            for variable_bind in variable_binds:
                self.testset_config["variables"].update(variable_bind)
        elif level == "testcase":
            self.testcase_config["variables"] = copy.deepcopy(self.testset_config["variables"])
            for variable_bind in variable_binds:
                self.testcase_config["variables"].update(variable_bind)

    def register_request(self, request_dict, level="testcase"):
        self.__update_context_config(level, "request", request_dict)

    def __update_context_config(self, level, config_type, config_mapping):
        """
        @param level: testset or testcase
        @param config_type: functions, variables or request
        @param config_mapping: functions config mapping or variables config mapping
        """
        if level == "testset":
            self.testset_config[config_type].update(config_mapping)
        elif level == "testcase":
            self.testcase_config[config_type].update(config_mapping)

    def get_parsed_request(self):
        """ get parsed request, with each variable replaced by bind value.
            testcase request shall inherit from testset request configs,
            but can not change testset configs, that's why we use copy.deepcopy here.
        """
        testcase_request_config = utils.deep_update_dict(
            copy.deepcopy(self.testset_config["request"]),
            self.testcase_config["request"]
        )

        parsed_request = testcase.parse_template(
            testcase_request_config,
            self._get_evaluated_testcase_variables()
        )

        return parsed_request

    def bind_extracted_variables(self, variables_mapping):
        """ bind extracted variable to current testcase context and testset context.
            since extracted variable maybe used in current testcase and next testcases.
        """
        self.testset_shared_variables_mapping.update(variables_mapping)
        self.testcase_variables_mapping.update(variables_mapping)

    def get_testcase_variables_mapping(self):
        return self.testcase_variables_mapping

    def _get_evaluated_testcase_variables(self):
        """ variables in variables_config will be evaluated each time
        """
        testcase_functions_config = copy.deepcopy(self.testset_config["functions"])
        testcase_functions_config.update(self.testcase_config["functions"])
        self.testcase_config["functions"] = testcase_functions_config

        testcase_variables_config = copy.deepcopy(self.testset_config["variables"])
        testcase_variables_config.update(self.testcase_config["variables"])
        self.testcase_config["variables"] = testcase_variables_config

        for var_name, var_value in self.testcase_config["variables"].items():
            self.testcase_variables_mapping[var_name] = self.get_eval_value(var_value)

        return self.testcase_variables_mapping

    def get_eval_value(self, data):
        """ evaluate data recursively, each variable in data will be evaluated.
        """
        if isinstance(data, (list, tuple)):
            return [self.get_eval_value(item) for item in data]

        if isinstance(data, dict):
            evaluated_data = {}
            for key, value in data.items():
                evaluated_data[key] = self.get_eval_value(value)

            return evaluated_data

        if isinstance(data, (int, float)):
            return data

        # data is in string format here
        data = data.strip()
        if utils.is_variable(data):
            # variable marker: $var
            variable_name = utils.parse_variable(data)
            value = self.testcase_variables_mapping.get(variable_name)
            if value is None:
                raise exception.ParamsError(
                    "%s is not defined in bind variables!" % variable_name)
            return value

        elif utils.is_functon(data):
            # function marker: ${func(1, 2, a=3, b=4)}
            fuction_meta = utils.parse_function(data)
            func_name = fuction_meta['func_name']
            args = fuction_meta.get('args', [])
            kwargs = fuction_meta.get('kwargs', {})
            args = self.get_eval_value(args)
            kwargs = self.get_eval_value(kwargs)
            return self.testcase_config["functions"][func_name](*args, **kwargs)
        else:
            return data
