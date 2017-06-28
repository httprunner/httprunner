import re
import importlib
from ate import exception, utils

class Context(object):
    """ Manages binding of variables
    """
    def __init__(self):
        self.functions = dict()
        self.variables = dict()  # Maps variable name to value
        self.extractors = dict()

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

    def bind_extractors(self, extract_binds):
        """ Bind named extractors to value within the context.
            value => parsed from requests.Response object
            key => extractor name, can be used as variable in next testcases
        @param (dict) extract_binds
            {
                "resp_status_code": "status_code",
                "resp_headers": "headers",
                "resp_headers_content_type": "headers.content-type",
                "resp_content": "content"
            }
        """
        self.extractors.update(extract_binds)

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
