from ate import utils

class TestcaseParser(object):

    def __init__(self, variables_binds={}):
        self.variables_binds = variables_binds

    def update_variables_binds(self, variables_mapping):
        """ update variables binds with new mapping.
        """
        if variables_mapping:
            self.variables_binds.update(variables_mapping)

    def parse(self, testcase_template):
        """ parse testcase_template, replace all variables with bind value.
        variables marker: ${variable}.
        @param (dict) testcase_template
            {
                "request": {
                    "url": "http://127.0.0.1:5000/api/users/${uid}",
                    "method": "POST",
                    "headers": {
                        "Content-Type": "application/json",
                        "authorization": "${authorization}",
                        "random": "${random}"
                    },
                    "body": "${data}"
                },
                "response": {
                    "status_code": "${expected_status}"
                }
            }
        @return (dict) parsed testcase with bind values
            {
                "request": {
                    "url": "http://127.0.0.1:5000/api/users/1000",
                    "method": "POST",
                    "headers": {
                        "Content-Type": "application/json",
                        "authorization": "a83de0ff8d2e896dbd8efb81ba14e17d",
                        "random": "A2dEx"
                    },
                    "body": '{"name": "user", "password": "123456"}'
                },
                "response": {
                    "status_code": 201
                }
            }
        """
        return self.substitute(testcase_template)

    def substitute(self, content):
        """ substitute content recursively, each variable will be replaced with bind value.
            variables marker: ${variable}.
        """
        if isinstance(content, str):
            return utils.parse_content_with_variables(content, self.variables_binds)

        if isinstance(content, list):
            return [self.substitute(item) for item in content]

        if isinstance(content, dict):
            parsed_content = {}
            for key, value in content.items():
                parsed_content[key] = self.substitute(value)

            return parsed_content

        return content
