import re
from ate.exception import ParamsError
from ate.utils import string_type


def parse_content_with_variables(content, variables_binds):
    """ replace variables with bind value
    """
    # check if content includes $variable
    matched = re.match(r"^(.*)\$(\w+)(.*)$", content)
    if matched:
        # this is a variable, and will replace with its bind value
        variable_name = matched.group(2)
        value = variables_binds.get(variable_name)
        if value is None:
            raise ParamsError(
                "%s is not defined in bind variables!" % variable_name)
        if matched.group(1) or matched.group(3):
            # e.g. /api/users/$uid
            return content.replace("$%s" % variable_name, value)

        return value

    return content

def parse_template(testcase_template, variables_binds):
    """ parse testcase_template, replace all variables with bind value.
    variables marker: $variable.
    @param (dict) testcase_template
        {
            "url": "http://127.0.0.1:5000/api/users/$uid",
            "method": "POST",
            "headers": {
                "Content-Type": "application/json",
                "authorization": "$authorization",
                "random": "$random"
            },
            "body": "$data"
        }
    @param (dict) variables binds mapping
        {
            "authorization": "a83de0ff8d2e896dbd8efb81ba14e17d",
            "random": "A2dEx",
            "data": '{"name": "user", "password": "123456"}'
        }
    @return (dict) parsed testcase with bind variable values
        {
            "url": "http://127.0.0.1:5000/api/users/1000",
            "method": "POST",
            "headers": {
                "Content-Type": "application/json",
                "authorization": "a83de0ff8d2e896dbd8efb81ba14e17d",
                "random": "A2dEx"
            },
            "body": '{"name": "user", "password": "123456"}'
        }
    """

    def substitute(content):
        """ substitute content recursively, each variable will be replaced with bind value.
        """
        if isinstance(content, string_type):
            return parse_content_with_variables(content, variables_binds)

        if isinstance(content, list):
            return [substitute(item) for item in content]

        if isinstance(content, dict):
            parsed_content = {}
            for key, value in content.items():
                parsed_content[key] = substitute(value)

            return parsed_content

        return content

    return substitute(testcase_template)
