import ast
import re

from ate import utils
from ate.exception import ParamsError

variable_regexp = r"\$([\w_]+)"
function_regexp = re.compile(r"^\$\{([\w_]+)\(([\$\w_ =,]*)\)\}$")


def get_contain_variables(content):
    """ extract all variable names from content, which is in format $variable
    @param (str) content
    @return (list) variable name list

    e.g. $variable => ["variable"]
         /blog/$postid => ["postid"]
         /$var1/$var2 => ["var1", "var2"]
         abc => []
    """
    return re.findall(variable_regexp, content)

def parse_variables(content, variable_mapping):
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
    variable_name_list = get_contain_variables(content)
    for variable_name in variable_name_list:
        if variable_name not in variable_mapping:
            raise ParamsError(
                "%s is not defined in bind variables!" % variable_name)

        variable_value = variable_mapping.get(variable_name)
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

def is_functon(content):
    """ check if content is a function, which is in format ${func()}
    @param (str) content
    @return (bool) True or False

    e.g. ${func()} => True
         ${func(5)} => True
         ${func(1, 2)} => True
         ${func(a=1, b=2)} => True
         $abc => False
         abc => False
    """
    matched = function_regexp.match(content)
    return True if matched else False

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

    e.g. ${func()} => {'func_name': 'func', 'args': [], 'kwargs': {}}
         ${func(5)} => {'func_name': 'func', 'args': [5], 'kwargs': {}}
         ${func(1, 2)} => {'func_name': 'func', 'args': [1, 2], 'kwargs': {}}
         ${func(a=1, b=2)} => {'func_name': 'func', 'args': [], 'kwargs': {'a': 1, 'b': 2}}
         ${func(1, 2, a=3, b=4)} => {'func_name': 'func', 'args': [1, 2], 'kwargs': {'a':3, 'b':4}}
    """
    function_meta = {
        "args": [],
        "kwargs": {}
    }
    matched = function_regexp.match(content)
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

def parse_content_with_bindings(content, variables_binds, functions_binds):
    """ evaluate content recursively, each variable in content will be
        evaluated with bind variables and functions.

    variables marker: $variable.
    @param (dict) content in any data structure
        {
            "url": "http://127.0.0.1:5000/api/users/$uid",
            "method": "POST",
            "headers": {
                "Content-Type": "application/json",
                "authorization": "$authorization",
                "random": "$random",
                "sum": "${add_two_nums(1, 2)}"
            },
            "body": "$data"
        }
    @param (dict) variables_binds, variables binds mapping
        {
            "authorization": "a83de0ff8d2e896dbd8efb81ba14e17d",
            "random": "A2dEx",
            "data": {"name": "user", "password": "123456"}
        }
    @param (dict) functions_binds, functions binds mapping
        {
            "add_two_nums": lambda a, b=1: a + b
        }
    @return (dict) parsed content with evaluated bind values
        {
            "url": "http://127.0.0.1:5000/api/users/1000",
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
            parse_content_with_bindings(item, variables_binds, functions_binds)
            for item in content
        ]

    if isinstance(content, dict):
        evaluated_data = {}
        for key, value in content.items():
            evaluated_data[key] = parse_content_with_bindings(
                value, variables_binds, functions_binds)

        return evaluated_data

    if isinstance(content, (int, float)):
        return content

    # content is in string format here
    content = "" if content is None else content.strip()

    if is_functon(content):
        # function marker: ${func(1, 2, a=3, b=4)}
        fuction_meta = parse_function(content)
        func_name = fuction_meta['func_name']

        func = functions_binds.get(func_name)
        if func is None:
            raise ParamsError(
                "%s is not defined in bind functions!" % func_name)

        args = fuction_meta.get('args', [])
        kwargs = fuction_meta.get('kwargs', {})
        args = parse_content_with_bindings(args, variables_binds, functions_binds)
        kwargs = parse_content_with_bindings(kwargs, variables_binds, functions_binds)
        return func(*args, **kwargs)

    elif get_contain_variables(content):
        parsed_data = parse_variables(content, variables_binds)
        return parsed_data

    else:
        return content
