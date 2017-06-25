import re
from ate import exception

class TestcaseParser(object):

    def __init__(self, variables_binds={}):
        self.variables_binds = variables_binds

    def parse(self, testcase_template):
        """ parse testcase_template, replace all variables with bind value.
        variables marker: ${variable}.
        @param testcase_template
            "request": {
                "url": "http://127.0.0.1:5000/api/users/${uid}",
                "method": "POST",
                "headers": {
                    "Content-Type": "application/json",
                    "authorization": "${authorization}",
                    "random": "${random}"
                },
                "body": "${json}"
            },
            "response": {
                "status_code": "${expected_status}",
                "headers": {
                    "Content-Type": "application/json"
                },
                "body": {
                    "success": True,
                    "msg": "user created successfully."
                }
            }
        """
        return self.substitute(testcase_template)

    def substitute(self, content):
        """ substitute content recursively, each variable will be replaced with bind value.
            variables marker: ${variable}.
        """
        if isinstance(content, str):
            # check if content includes ${variable}
            matched = re.match(r"(.*)\$\{(.*)\}(.*)", content)
            if matched:
                # this is a variable, and will replace with its bind value
                variable_name = matched.group(2)
                value = self.variables_binds.get(variable_name)
                if value is None:
                    raise exception.ParamsError(
                        "%s is not defined in bind variables!" % variable_name)
                if matched.group(1) or matched.group(3):
                    # e.g. /api/users/${uid}
                    return re.sub(r"\$\{.*\}", value, content)

                return value

            return content

        if isinstance(content, list):
            return [self.substitute(item) for item in content]

        if isinstance(content, dict):
            parsed_content = {}
            for key, value in content.items():
                parsed_content[key] = self.substitute(value)

            return parsed_content

        return content
