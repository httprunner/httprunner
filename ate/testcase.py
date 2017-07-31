from ate.exception import ParamsError
from ate import utils


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

    if utils.is_functon(content):
        # function marker: ${func(1, 2, a=3, b=4)}
        fuction_meta = utils.parse_function(content)
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

    elif utils.get_contain_variables(content):
        parsed_data = utils.parse_variables(content, variables_binds)
        return parsed_data

    else:
        return content
