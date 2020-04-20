import re
from typing import Any, Set, Text
from typing import Dict

from httprunner.v3 import exceptions

absolute_http_url_regexp = re.compile(r"^https?://", re.I)

# use $$ to escape $ notation
dolloar_regex_compile = re.compile(r"\$\$")
# variable notation, e.g. ${var} or $var
variable_regex_compile = re.compile(r"\$\{(\w+)\}|\$(\w+)")
# function notation, e.g. ${func1($var_1, $var_3)}
function_regex_compile = re.compile(r"\$\{(\w+)\(([\$\w\.\-/\s=,]*)\)\}")


def build_url(base_url, path):
    """ prepend url with base_url unless it's already an absolute URL """
    if absolute_http_url_regexp.match(path):
        return path
    elif base_url:
        return "{}/{}".format(base_url.rstrip("/"), path.lstrip("/"))
    else:
        raise exceptions.ParamsError("base url missed!")


def regex_findall_variables(content):
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


def parse_string_variables(content, variables_mapping):
    """ parse string content with variables mapping.

    Args:
        content (str): string content to be parsed.
        variables_mapping (dict): variables mapping.

    Returns:
        str: parsed string content.

    Examples:
        >>> content = "/api/users/$uid"
        >>> variables_mapping = {"$uid": 1000}
        >>> parse_string_variables(content, variables_mapping)
            "/api/users/1000"

    """
    variables_list = extract_variables(content)
    for variable_name in variables_list:
        variable_value = variables_mapping[variable_name]

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


def parse_content(content: Any, variables_mapping: Dict[str, Any] = None, functions_mapping=None):
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
        # Notice: _eval_content_functions must be called before _eval_content_variables
        # content = parse_string_functions(content, variables_mapping, functions_mapping)

        # replace variables with binding value
        content = parse_string_variables(content, variables_mapping)

        return content

    elif isinstance(content, (list, set, tuple)):
        return [
            parse_content(item, variables_mapping)
            for item in content
        ]

    elif isinstance(content, dict):
        parsed_content = {}
        for key, value in content.items():
            parsed_key = parse_content(key, variables_mapping)
            parsed_value = parse_content(value, variables_mapping)
            parsed_content[parsed_key] = parsed_value

        return parsed_content

    return content


def parse_variables_mapping(variables_mapping: Dict[str, Any]):

    parsed_variables: Dict[str, Any] = {}

    while len(parsed_variables) != len(variables_mapping):
        for var_name in variables_mapping:

            var_value = variables_mapping[var_name]
            # variables = extract_variables(var_value)

            if var_name in parsed_variables:
                continue

            parsed_value = parse_content(var_value, parsed_variables)
            parsed_variables[var_name] = parsed_value

    return parsed_variables
