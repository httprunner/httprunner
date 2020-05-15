import ast
import builtins
import re
from typing import Any, Set, Text, Callable, List, Dict

from httprunner import loader, utils, exceptions
from httprunner.schema import VariablesMapping, FunctionsMapping

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
            vars_list.append(var_tuple[0] or var_tuple[1])
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


def parse_function_params(params: Text) -> Dict:
    """ parse function params to args and kwargs.

    Args:
        params (str): function param in string

    Returns:
        dict: function meta dict

            {
                "args": [],
                "kwargs": {}
            }

    Examples:
        >>> parse_function_params("")
        {'args': [], 'kwargs': {}}

        >>> parse_function_params("5")
        {'args': [5], 'kwargs': {}}

        >>> parse_function_params("1, 2")
        {'args': [1, 2], 'kwargs': {}}

        >>> parse_function_params("a=1, b=2")
        {'args': [], 'kwargs': {'a': 1, 'b': 2}}

        >>> parse_function_params("1, 2, a=3, b=4")
        {'args': [1, 2], 'kwargs': {'a':3, 'b':4}}

    """
    function_meta = {"args": [], "kwargs": {}}

    params_str = params.strip()
    if params_str == "":
        return function_meta

    args_list = params_str.split(",")
    for arg in args_list:
        arg = arg.strip()
        if "=" in arg:
            key, value = arg.split("=")
            function_meta["kwargs"][key.strip()] = parse_string_value(value.strip())
        else:
            function_meta["args"].append(parse_string_value(arg))

    return function_meta


def get_mapping_variable(
    variable_name: Text, variables_mapping: VariablesMapping
) -> Any:
    """ get variable from variables_mapping.

    Args:
        variable_name (str): variable name
        variables_mapping (dict): variables mapping

    Returns:
        mapping variable value.

    Raises:
        exceptions.VariableNotFound: variable is not found.

    """
    # TODO: get variable from debugtalk module and environ
    try:
        return variables_mapping[variable_name]
    except KeyError:
        raise exceptions.VariableNotFound(
            f"{variable_name} not found in {variables_mapping}"
        )


def get_mapping_function(
    function_name: Text, functions_mapping: FunctionsMapping
) -> Callable:
    """ get function from functions_mapping,
        if not found, then try to check if builtin function.

    Args:
        function_name (str): function name
        functions_mapping (dict): functions mapping

    Returns:
        mapping function object.

    Raises:
        exceptions.FunctionNotFound: function is neither defined in debugtalk.py nor builtin.

    """
    if function_name in functions_mapping:
        return functions_mapping[function_name]

    elif function_name in ["parameterize", "P"]:
        return loader.load_csv_file

    elif function_name in ["environ", "ENV"]:
        return utils.get_os_environ

    elif function_name in ["multipart_encoder", "multipart_content_type"]:
        # extension for upload test
        from httprunner.ext import uploader

        return getattr(uploader, function_name)

    try:
        # check if HttpRunner builtin functions
        built_in_functions = loader.load_builtin_functions()
        return built_in_functions[function_name]
    except KeyError:
        pass

    try:
        # check if Python builtin functions
        return getattr(builtins, function_name)
    except AttributeError:
        pass

    raise exceptions.FunctionNotFound(f"{function_name} is not found.")


def parse_string(
    raw_string: Text,
    variables_mapping: VariablesMapping,
    functions_mapping: FunctionsMapping,
) -> Any:
    """ parse string content with variables and functions mapping.

    Args:
        raw_string: raw string content to be parsed.
        variables_mapping: variables mapping.
        functions_mapping: functions mapping.

    Returns:
        str: parsed string content.

    Examples:
        >>> raw_string = "abc${add_one($num)}def"
        >>> variables_mapping = {"num": 3}
        >>> functions_mapping = {"add_one": lambda x: x + 1}
        >>> parse_string(raw_string, variables_mapping, functions_mapping)
            "abc4def"

    """
    try:
        match_start_position = raw_string.index("$", 0)
        parsed_string = raw_string[0:match_start_position]
    except ValueError:
        parsed_string = raw_string
        return parsed_string

    while match_start_position < len(raw_string):

        # Notice: notation priority
        # $$ > ${func($a, $b)} > $var

        # search $$
        dollar_match = dolloar_regex_compile.match(raw_string, match_start_position)
        if dollar_match:
            match_start_position = dollar_match.end()
            parsed_string += "$"
            continue

        # search function like ${func($a, $b)}
        func_match = function_regex_compile.match(raw_string, match_start_position)
        if func_match:
            func_name = func_match.group(1)
            func = get_mapping_function(func_name, functions_mapping)

            func_params_str = func_match.group(2)
            function_meta = parse_function_params(func_params_str)
            args = function_meta["args"]
            kwargs = function_meta["kwargs"]

            parsed_args = parse_data(args, variables_mapping, functions_mapping)
            parsed_kwargs = parse_data(kwargs, variables_mapping, functions_mapping)
            func_eval_value = func(*parsed_args, **parsed_kwargs)

            func_raw_str = "${" + func_name + f"({func_params_str})" + "}"
            if func_raw_str == raw_string:
                # raw_string is a function, e.g. "${add_one(3)}", return its eval value directly
                return func_eval_value

            # raw_string contains one or many functions, e.g. "abc${add_one(3)}def"
            parsed_string += str(func_eval_value)
            match_start_position = func_match.end()
            continue

        # search variable like ${var} or $var
        var_match = variable_regex_compile.match(raw_string, match_start_position)
        if var_match:
            var_name = var_match.group(1) or var_match.group(2)
            var_value = get_mapping_variable(var_name, variables_mapping)

            if f"${var_name}" == raw_string or "${" + var_name + "}" == raw_string:
                # raw_string is a variable, $var or ${var}, return its value directly
                return var_value

            # raw_string contains one or many variables, e.g. "abc${var}def"
            parsed_string += str(var_value)
            match_start_position = var_match.end()
            continue

        curr_position = match_start_position
        try:
            # find next $ location
            match_start_position = raw_string.index("$", curr_position + 1)
            remain_string = raw_string[curr_position:match_start_position]
        except ValueError:
            remain_string = raw_string[curr_position:]
            # break while loop
            match_start_position = len(raw_string)

        parsed_string += remain_string

    return parsed_string


def parse_data(
    raw_data: Any,
    variables_mapping: VariablesMapping = None,
    functions_mapping: FunctionsMapping = None,
) -> Any:
    """ parse raw data with evaluated variables mapping.
        Notice: variables_mapping should not contain any variable or function.
    """
    if isinstance(raw_data, str):
        # content in string format may contains variables and functions
        variables_mapping = variables_mapping or {}
        functions_mapping = functions_mapping or {}
        raw_data = raw_data.strip()
        return parse_string(raw_data, variables_mapping, functions_mapping)

    elif isinstance(raw_data, (list, set, tuple)):
        return [
            parse_data(item, variables_mapping, functions_mapping) for item in raw_data
        ]

    elif isinstance(raw_data, dict):
        parsed_data = {}
        for key, value in raw_data.items():
            parsed_key = parse_data(key, variables_mapping, functions_mapping)
            parsed_value = parse_data(value, variables_mapping, functions_mapping)
            parsed_data[parsed_key] = parsed_value

        return parsed_data

    else:
        # other types, e.g. None, int, float, bool
        return raw_data


def parse_variables_mapping(
    variables_mapping: VariablesMapping, functions_mapping: FunctionsMapping = None
) -> VariablesMapping:

    parsed_variables: VariablesMapping = {}

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
                v_name for v_name in variables if v_name not in variables_mapping
            ]
            if not_defined_variables:
                # e.g. {"varA": "123$varB", "varB": "456$varC"}
                # e.g. {"varC": "${sum_two($a, $b)}"}
                raise exceptions.VariableNotFound(not_defined_variables)

            try:
                parsed_value = parse_data(
                    var_value, parsed_variables, functions_mapping
                )
            except exceptions.VariableNotFound:
                continue

            parsed_variables[var_name] = parsed_value

    return parsed_variables
