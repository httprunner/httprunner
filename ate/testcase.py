from ate import utils


def parse_template(testcase_template, variables_binds):
    """ parse testcase_template, replace all variables with bind value.
    variables marker: ${variable}.
    @param (dict) testcase_template
        {
            "url": "http://127.0.0.1:5000/api/users/${uid}",
            "method": "POST",
            "headers": {
                "Content-Type": "application/json",
                "authorization": "${authorization}",
                "random": "${random}"
            },
            "body": "${data}"
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
            variables marker: ${variable}.
        """
        if isinstance(content, str):
            return utils.parse_content_with_variables(content, variables_binds)

        if isinstance(content, list):
            return [substitute(item) for item in content]

        if isinstance(content, dict):
            parsed_content = {}
            for key, value in content.items():
                parsed_content[key] = substitute(value)

            return parsed_content

        return content

    return substitute(testcase_template)
