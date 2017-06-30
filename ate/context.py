import importlib
import re
import types

from ate import exception, utils


def is_function(tup):
    """
    Takes (name, object) tuple, returns True if it is a function.
    """
    name, item = tup
    return isinstance(item, types.FunctionType)

class Context(object):
    """ Manages binding of variables
    """
    def __init__(self):
        self.functions = dict()
        self.variables = dict()  # Maps variable name to value

    def import_requires(self, modules):
        """ import required modules dynamicly
        """
        for module_name in modules:
            globals()[module_name] = importlib.import_module(module_name)

    def bind_functions(self, function_binds):
        """ Bind named functions within the context
            This allows for passing in self-defined functions in testing.
            e.g. function_binds:
            {
                "add_one": lambda x: x + 1,
                "add_two_nums": "lambda x, y: x + y"
            }
        """
        for func_name, function in function_binds.items():
            if isinstance(function, str):
                function = eval(function)
            self.functions[func_name] = function

    def import_module_functions(self, modules):
        """ import modules and bind all functions within the context
        """
        for module_name in modules:
            imported = importlib.import_module(module_name)
            imported_functions_dict = dict(filter(is_function, vars(imported).items()))
            self.functions.update(imported_functions_dict)

    def bind_variables(self, variable_binds):
        """ Bind named variables to value within the context.
            This allows for passing in variables or functions.
            e.g. variable_binds:
            [
                {"TOKEN": "debugtalk"},
                {"random": {"func": "gen_random_string", "args": [5]}},
                {"json": {'name': 'user', 'password': '123456'}},
                {"md5": {"func": "gen_md5", "args": ["$TOKEN", "$json", "$random"]}}
            ]
        """
        for variable_bind_map in variable_binds:
            for var_name, var_value in variable_bind_map.items():
                self.variables[var_name] = self.get_eval_value(var_value)

    def update_variables(self, variables_mapping):
        """ update context variables binds with new variables mapping
        """
        self.variables.update(variables_mapping)

    def get_eval_value(self, data):
        """ evaluate data recursively, each variable in data will be evaluated.
            variables marker: ${variable}.
        """
        if isinstance(data, str):
            return utils.parse_content_with_variables(data, self.variables)

        if isinstance(data, list):
            return [self.get_eval_value(item) for item in data]

        if isinstance(data, dict):
            if "func" in data:
                # this is a function, e.g. {"func": "gen_random_string", "args": [5]}
                # function marker: "func" key in dict
                # the function will be called, and its return value will be binded to the variable.
                func_name = data['func']
                args = self.get_eval_value(data.get('args', []))
                kargs = self.get_eval_value(data.get('kargs', {}))
                return self.functions[func_name](*args, **kargs)
            else:
                evaluated_data = {}
                for key, value in data.items():
                    evaluated_data[key] = self.get_eval_value(value)

                return evaluated_data

        return data
