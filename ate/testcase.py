import ast
import os
import re

from ate import exception, utils

variable_regexp = r"\$([\w_]+)"
function_regexp = r"\$\{([\w_]+\([\$\w_ =,]*\))\}"
function_regexp_compile = re.compile(r"^([\w_]+)\(([\$\w_ =,]*)\)$")
api_overall_dict = {}


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
    function_meta = {
        "args": [],
        "kwargs": {}
    }
    matched = function_regexp_compile.match(content)
    function_meta["func_name"] = matched.group(1)

    args_str = matched.group(2).replace(" ", "")
    if args_str == "":
        return function_meta

    args_list = args_str.split(',')
    for arg in args_list:
        if '=' in arg:
            key, value = arg.split('=')
            function_meta["kwargs"][key] = parse_string_value(value)
        else:
            function_meta["args"].append(parse_string_value(arg))

    return function_meta

def load_testcases_by_path(path):
    """ load testcases from file path
    @param path
        path could be in several type:
            - absolute/relative file path
            - absolute/relative folder path
            - list/set container with file(s) and/or folder(s)
    @return testcase sets list, each testset is corresponding to a file
        [
            {"name": "desc1", "config": {}, "testcases": [testcase11, testcase12]},
            {"name": "desc2", "config": {}, "testcases": [testcase21, testcase22, testcase23]},
        ]
    """
    if isinstance(path, (list, set)):
        testsets_list = []

        for file_path in set(path):
            _testsets_list = load_testcases_by_path(file_path)
            testsets_list.extend(_testsets_list)

        return testsets_list

    if not os.path.isabs(path):
        path = os.path.join(os.getcwd(), path)

    if os.path.isdir(path):
        files_list = utils.load_folder_files(path, file_type="test", recursive=True)
        return load_testcases_by_path(files_list)

    elif os.path.isfile(path):
        testset = {
            "name": "",
            "config": {
                "path": path
            },
            "testcases": []
        }
        testcases_list = utils.load_testcases(path)
        dir_path = os.path.dirname(os.path.abspath(path))

        for item in testcases_list:
            for key in item:
                if key == "config":
                    testset["config"].update(item["config"])
                    testset["name"] = item["config"].get("name", "")
                elif key == "test":
                    test_block_dict = item["test"]
                    if "api" in test_block_dict:
                        testcase_list = load_testcases_by_call(test_block_dict, dir_path, "api")
                    else:
                        testcase_list = [test_block_dict]

                    testset["testcases"].extend(testcase_list)

        return [testset] if testset["testcases"] else []

    else:
        return []

def load_testcases_by_call(test_block_dict, dir_path, call_type):
    api_call = test_block_dict[call_type]
    function_meta = parse_function(api_call)
    func_name = function_meta["func_name"]
    api_call_args = function_meta["args"]
    api_info = get_api_definition(func_name, dir_path)
    api_def_args = api_info.get("function_meta").get("args", [])

    if len(api_call_args) != len(api_def_args):
        raise exception.ParamsError("api call args invalid!")

    args_mapping = {}
    for index, item in enumerate(api_def_args):
        if api_call_args[index] == item:
            continue

        args_mapping[item] = api_call_args[index]

    if args_mapping:
        api_info = substitute_variables_with_mapping(api_info, args_mapping)

    test_block_dict.update(api_info)

    return [test_block_dict]

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
    if isinstance(content, (list, tuple)):
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

    if isinstance(content, (int, utils.long_type, float, complex)):
        return content

    # content is in string format here
    for var, value in mapping.items():
        if content == var:
            # content is a variable
            content = value
        else:
            content = content.replace(var, str(value))

    return content

def get_api_definition(name, dir_path):
    """ get expected api from dir_path upward recursively
    @param
        name: api name
        dir_path: start search dir path
    @return
        expected api info if found, otherwise raise ApiNotFound exception
    """
    api_dir_dict = api_overall_dict.get(dir_path)
    if not api_dir_dict:
        api_dir_dict = load_api_definition(dir_path)
        api_overall_dict[dir_path] = api_dir_dict

    api_info = api_dir_dict.get(name)
    if api_info:
        return api_info

    parent_dir_path = os.path.dirname(dir_path)
    if dir_path == parent_dir_path:
        # system root path
        err_msg = "{} not found in recursive upward path!".format(name)
        raise exception.ApiNotFound(err_msg)

    return get_api_definition(name, parent_dir_path)

def load_api_definition(dir_path):
    """ load all api definitions in specified dir path
    @param (str) dir_path
    @return (dict) all api definitions in dir_path merged in one dict
    """
    api_files = utils.load_folder_files(dir_path, file_type="api", recursive=False)

    api_def_list = []
    for api_file in api_files:
        api_def_list.extend(utils.load_testcases(api_file))

    api_dir_dict = {}

    for item in api_def_list:
        for key in item:
            if key == "api":
                api_def = item["api"].pop("def")
                function_meta = parse_function(api_def)
                func_name = function_meta["func_name"]

                api_info = {}
                api_info["function_meta"] = function_meta
                api_info.update(item["api"])
                api_dir_dict[func_name] = api_info

    return api_dir_dict


class TestcaseParser(object):

    def __init__(self, variables_binds={}, functions_binds={}, file_path=None):
        self.bind_variables(variables_binds)
        self.bind_functions(functions_binds)
        self.file_path = file_path

    def bind_variables(self, variables_binds):
        """ bind variables to current testcase parser
        @param (dict) variables_binds, variables binds mapping
            {
                "authorization": "a83de0ff8d2e896dbd8efb81ba14e17d",
                "random": "A2dEx",
                "data": {"name": "user", "password": "123456"},
                "uuid": 1000
            }
        """
        self.variables_binds = variables_binds

    def bind_functions(self, functions_binds):
        """ bind functions to current testcase parser
        @param (dict) functions_binds, functions binds mapping
            {
                "add_two_nums": lambda a, b=1: a + b
            }
        """
        self.functions_binds = functions_binds

    def get_bind_item(self, item_type, item_name):
        if item_type == "function":
            if item_name in self.functions_binds:
                return self.functions_binds[item_name]
        elif item_type == "variable":
            if item_name in self.variables_binds:
                return self.variables_binds[item_name]
        else:
            raise exception.ParamsError("bind item should only be function or variable.")

        try:
            assert self.file_path is not None
            return utils.search_conf_item(self.file_path, item_type, item_name)
        except (AssertionError, exception.FunctionNotFound):
            raise exception.ParamsError(
                "{} is not defined in bind {}s!".format(item_name, item_type))

    def eval_content_functions(self, content):
        functions_list = extract_functions(content)
        for func_content in functions_list:
            function_meta = parse_function(func_content)
            func_name = function_meta['func_name']

            func = self.get_bind_item("function", func_name)

            args = function_meta.get('args', [])
            kwargs = function_meta.get('kwargs', {})
            args = self.parse_content_with_bindings(args)
            kwargs = self.parse_content_with_bindings(kwargs)
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

    def eval_content_variables(self, content):
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
            variable_value = self.get_bind_item("variable", variable_name)

            if "${}".format(variable_name) == content:
                # content is a variable
                content = variable_value
            else:
                # content contains one or many variables
                content = content.replace(
                    "${}".format(variable_name),
                    str(variable_value), 1
                )

        return content

    def parse_content_with_bindings(self, content):
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

        if isinstance(content, (list, tuple)):
            return [
                self.parse_content_with_bindings(item)
                for item in content
            ]

        if isinstance(content, dict):
            evaluated_data = {}
            for key, value in content.items():
                eval_key = self.parse_content_with_bindings(key)
                eval_value = self.parse_content_with_bindings(value)
                evaluated_data[eval_key] = eval_value

            return evaluated_data

        if isinstance(content, (int, utils.long_type, float, complex)):
            return content

        # content is in string format here
        content = "" if content is None else content.strip()

        # replace functions with evaluated value
        # Notice: eval_content_functions must be called before eval_content_variables
        content = self.eval_content_functions(content)

        # replace variables with binding value
        content = self.eval_content_variables(content)

        return content
