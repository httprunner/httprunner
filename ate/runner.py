import requests

from ate import exception, response
from ate.context import Context
from ate.testcase import TestcaseParser


class TestRunner(object):

    def __init__(self):
        self.client = requests.Session()
        self.context = Context()
        self.testcase_parser = TestcaseParser()

    def pre_config(self, config_dict):
        """ create/update variables binds
        @param config_dict
            {
                "name": "description content",
                "requires": ["random", "hashlib"],
                "function_binds": {
                    "gen_random_string": \
                        "lambda str_len: ''.join(random.choice(string.ascii_letters + \
                        string.digits) for _ in range(str_len))",
                    "gen_md5": \
                        "lambda *str_args: hashlib.md5(''.join(str_args).\
                        encode('utf-8')).hexdigest()"
                },
                "variable_binds": [
                    {"TOKEN": "debugtalk"},
                    {"random": {"func": "gen_random_string", "args": [5]}},
                ]
            }
        """
        requires = config_dict.get('requires', [])
        self.context.import_requires(requires)

        function_binds = config_dict.get('function_binds', {})
        self.context.bind_functions(function_binds)

        variable_binds = config_dict.get('variable_binds', [])
        self.context.bind_variables(variable_binds)

        extract_binds = config_dict.get('extract_binds', {})
        self.context.bind_extractors(extract_binds)

        self.testcase_parser.update_variables_binds(self.context.variables)

    def parse_testcase(self, testcase):
        """ parse testcase with variables binds if it is a template.
        @param (dict) testcase
            {
                "name": "testcase description",
                "requires": [],  # optional, override
                "function_binds": {}, # optional, override
                "variable_binds": {}, # optional, override
                "request": {},
                "response": {}
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
        self.pre_config(testcase)

        parsed_testcase = self.testcase_parser.parse(testcase)
        return parsed_testcase

    def run_test(self, testcase):
        """ run single testcase.
        @param (dict) testcase
            {
                "name": "testcase description",
                "requires": [],  # optional, override
                "function_binds": {}, # optional, override
                "variable_binds": {}, # optional, override
                "request": {},
                "response": {}
            }
        @return (tuple) test result of single testcase
            (success, diff_content)
        """
        testcase = self.parse_testcase(testcase)

        req_kwargs = testcase['request']

        try:
            url = req_kwargs.pop('url')
            method = req_kwargs.pop('method')
        except KeyError:
            raise exception.ParamsError("URL or METHOD missed!")

        resp_obj = self.client.request(url=url, method=method, **req_kwargs)
        response.extract_response(resp_obj, self.context)
        diff_content = response.diff_response(resp_obj, testcase['response'])
        success = False if diff_content else True
        return success, diff_content

    def run_testset(self, testset):
        """ run single testset, including one or several testcases.
        @param (dict) testset
            {
                "name": "testset description",
                "config": {
                    "name": "testset description",
                    "requires": [],
                    "function_binds": {},
                    "variable_binds": []
                },
                "testcases": [
                    {
                        "name": "testcase description",
                        "variable_binds": {}, # override
                        "request": {},
                        "response": {}
                    },
                    testcase12
                ]
            }
        @return (list) test results of testcases
            [
                (success, diff_content),    # testcase1
                (success, diff_content)     # testcase2
            ]
        """
        results = []

        config_dict = testset.get("config", {})
        self.pre_config(config_dict)
        testcases = testset.get("testcases", [])
        for testcase in testcases:
            result = self.run_test(testcase)
            results.append(result)

        return results

    def run_testsets(self, testsets):
        """ run testsets, including one or several testsets.
        @param testsets
            [
                testset1,
                testset2,
            ]
        @return (list) test results of testsets
            [
                [   # testset1
                    (success, diff_content),    # testcase11
                    (success, diff_content)     # testcase12
                ],
                [   # testset2
                    (success, diff_content),    # testcase21
                    (success, diff_content)     # testcase22
                ]
            ]
        """
        return [self.run_testset(testset) for testset in testsets]
