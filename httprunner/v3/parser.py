import ast
import re
from typing import Any, Set, Text, Callable, Tuple, List, Dict, Union

from httprunner.v3 import exceptions
from httprunner.v3.exceptions import VariableNotFound, FunctionNotFound

absolute_http_url_regexp = re.compile(r"^https?://", re.I)

# use $$ to escape $ notation
dolloar_regex_compile = re.compile(r"\$\$")
# variable notation, e.g. ${var} or $var
variable_regex_compile = re.compile(r"\$\{(\w+)\}|\$(\w+)")
# function notation, e.g. ${func1($var_1, $var_3)}
function_regex_compile = re.compile(r"\$\{(\w+)\(([\$\w\.\-/\s=,]*)\)\}")


def parse_string_value(str_value: Text) -> Any:
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


def build_url(base_url, path):
    """ prepend url with base_url unless it's already an absolute URL """
    if absolute_http_url_regexp.match(path):
        return path
    elif base_url:
        return "{}/{}".format(base_url.rstrip("/"), path.lstrip("/"))
    else:
        raise exceptions.ParamsError("base url missed!")


def regex_findall_variables(content: Text) -> List[Text]:
    """ extract all variable names from content, which is in format $variable

    Args:
        content (str): string content

    Returns:
        list: variables list extracted from string content

    Examples:
        >>> regex_findall_variables("$variable")
        ["variable"]

        >>> regex_findall_variables("/blog/$postid")
        ["postid"]

        >>> regex_findall_variables("/$var1/$var2")
        ["var1", "var2"]

        >>> regex_findall_variables("abc")
        []

    """
    try:
        vars_list = []
        for var_tuple in variable_regex_compile.findall(content):
            vars_list.append(
                var_tuple[0] or var_tuple[1]
            )
        return vars_list
    except TypeError:
        return []


def regex_findall_functions(content: Text) -> List[Text]:
    """ extract all functions from string content, which are in format ${fun()}

    Args:
        content (str): string content

    Returns:
        list: functions list extracted from string content

    Examples:
        >>> regex_findall_functions("${func(5)}")
        ["func(5)"]

        >>> regex_findall_functions("${func(a=1, b=2)}")
        ["func(a=1, b=2)"]

        >>> regex_findall_functions("/api/1000?_t=${get_timestamp()}")
        ["get_timestamp()"]

        >>> regex_findall_functions("/api/${add(1, 2)}")
        ["add(1, 2)"]

        >>> regex_findall_functions("/api/${add(1, 2)}?_t=${get_timestamp()}")
        ["add(1, 2)", "get_timestamp()"]

    """
    try:
        return function_regex_compile.findall(content)
    except TypeError:
        return []


def parse_args_str(arg_str: Text) -> Tuple[List, Dict]:
    """ parse function args and kwargs from function.

    Args:
        arg_str (str): function str contains args and kwargs

    Returns:
        dict: function meta dict

            {
                "func_name": "xxx",
                "args": [],
                "kwargs": {}
            }

    Examples:
        >>> parse_args_str("")
        {'args': [], 'kwargs': {}}

        >>> parse_args_str("5")
        {'args': [5], 'kwargs': {}}

        >>> parse_args_str("1, 2")
        {'args': [1, 2], 'kwargs': {}}

        >>> parse_args_str("a=1, b=2")
        {'args': [], 'kwargs': {'a': 1, 'b': 2}}

        >>> parse_args_str("1, 2, a=3, b=4")
        {'args': [1, 2], 'kwargs': {'a':3, 'b':4}}

    """
    args = []
    kwargs = {}
    arg_str = arg_str.strip()
    if arg_str == "":
        return args, kwargs

    arg_list = arg_str.split(',')
    for arg in arg_list:
        arg = arg.strip()
        if '=' in arg:
            key, value = arg.split('=')
            kwargs[key.strip()] = parse_string_value(value.strip())
        else:
            args.append(parse_string_value(arg))

    return args, kwargs


def extract_variables(content: Any) -> Set:
    """ extract all variables in content recursively.
    """
    if isinstance(content, (list, set, tuple)):
        variables = set()
        for item in content:
            variables = variables | extract_variables(item)
        return variables

    elif isinstance(content, dict):
        variables = set()
        for key, value in content.items():
            variables = variables | extract_variables(value)
        return variables

    elif isinstance(content, str):
        return set(regex_findall_variables(content))

    return set()


def parse_string_functions(
        content: Text,
        variables_mapping: Dict[Text, Any],
        functions_mapping: Dict[Text, Callable]) -> Text:
    """ parse string content with functions mapping.

    Args:
        content (str): string content to be parsed.
        variables_mapping (dict): variables mapping.
        functions_mapping (dict): functions mapping.

    Returns:
        str: parsed string content.

    Examples:
        >>> content = "abc${add_one(3)}def"
        >>> functions_mapping = {"add_one": lambda x: x + 1}
        >>> parse_string_functions(content, {}, functions_mapping)
            "abc4def"

    """
    functions_list = regex_findall_functions(content)
    for func_meta_tuple in functions_list:
        func_name, args_str = func_meta_tuple
        args, kwargs = parse_args_str(args_str)

        args = parse_content(args, variables_mapping, functions_mapping)
        kwargs = parse_content(kwargs, variables_mapping, functions_mapping)

        try:
            func = functions_mapping[func_name]
        except KeyError:
            raise FunctionNotFound(f"{func_name} not found in {functions_mapping}")

        eval_value = func(*args, **kwargs)

        func_content = "${" + func_name + f"({args_str})" + "}"
        if func_content == content:
            # content is a function, e.g. "${add_one(3)}"
            content = eval_value
        else:
            # content contains one or many functions, e.g. "abc${add_one(3)}def"
            content = content.replace(
                func_content,
                str(eval_value), 1
            )

    return content


def parse_string_variables(
        content: Text,
        variables_mapping: Dict[Text, Any]) -> Text:
    """ parse string content with variables mapping.

    Args:
        content (str): string content to be parsed.
        variables_mapping (dict): variables mapping.

    Returns:
        str: parsed string content.

    Examples:
        >>> content = "/api/users/$uid"
        >>> variables_mapping = {"uid": 1000}
        >>> parse_string_variables(content, variables_mapping)
            "/api/users/1000"

    """
    variables_list = extract_variables(content)
    for variable_name in variables_list:
        try:
            variable_value = variables_mapping[variable_name]
        except KeyError:
            raise VariableNotFound(f"{variable_name} not found in {variables_mapping}")

        # TODO: replace variable label from $var to {{var}}
        if f"${variable_name}" == content:
            # content is a variable
            content = variable_value
        else:
            # content contains one or several variables
            if not isinstance(variable_value, str):
                variable_value = str(variable_value)

            content = content.replace(
                f"${variable_name}",
                variable_value, 1
            )

    return content


def parse_content(
        content: Any,
        variables_mapping: Dict[Text, Any] = None,
        functions_mapping: Dict[Text, Callable] = None) -> Any:
    """ parse content with evaluated variables mapping.
        Notice: variables_mapping should not contain any variable or function.
    """
    # TODO: refactor type check
    if content is None or isinstance(content, (int, float, bool)):
        return content

    elif isinstance(content, str):
        # content is in string format here
        variables_mapping = variables_mapping or {}
        functions_mapping = functions_mapping or {}
        content = content.strip()

        # replace functions with evaluated value
        # Notice: parse_string_functions must be called before parse_string_variables
        content = parse_string_functions(content, variables_mapping, functions_mapping)

        # replace variables with binding value
        content = parse_string_variables(content, variables_mapping)

        return content

    elif isinstance(content, (list, set, tuple)):
        return [
            parse_content(item, variables_mapping, functions_mapping)
            for item in content
        ]

    elif isinstance(content, dict):
        parsed_content = {}
        for key, value in content.items():
            parsed_key = parse_content(key, variables_mapping, functions_mapping)
            parsed_value = parse_content(value, variables_mapping, functions_mapping)
            parsed_content[parsed_key] = parsed_value

        return parsed_content

    return content


def parse_variables_mapping(
        variables_mapping: Dict[Text, Any],
        functions_mapping: Dict[Text, Callable] = None) -> Dict[Text, Any]:

    parsed_variables: Dict[Text, Any] = {}

    while len(parsed_variables) != len(variables_mapping):
        for var_name in variables_mapping:

            if var_name in parsed_variables:
                continue

            var_value = variables_mapping[var_name]
            variables = extract_variables(var_value)

            # check if reference variable itself
            if var_name in variables:
                # e.g.
                # variables_mapping = {"token": "abc$token"}
                # variables_mapping = {"key": ["$key", 2]}
                raise exceptions.VariableNotFound(var_name)

            # check if reference variable not in variables_mapping
            not_defined_variables = [
                v_name
                for v_name in variables
                if v_name not in variables_mapping
            ]
            if not_defined_variables:
                # e.g. {"varA": "123$varB", "varB": "456$varC"}
                # e.g. {"varC": "${sum_two($a, $b)}"}
                raise VariableNotFound(not_defined_variables)

            try:
                parsed_value = parse_content(
                    var_value, parsed_variables, functions_mapping)
            except VariableNotFound:
                continue

            parsed_variables[var_name] = parsed_value

    return parsed_variables
