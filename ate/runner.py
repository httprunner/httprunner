import requests

from ate import exception, utils
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
        @return variables binds mapping
            {
                "TOKEN": "debugtalk",
                "random": "A2dEx"
            }
        """
        requires = config_dict.get('requires', [])
        self.context.import_requires(requires)

        function_binds = config_dict.get('function_binds', {})
        self.context.bind_functions(function_binds)

        variable_binds = config_dict.get('variable_binds', [])
        self.context.bind_variables(variable_binds)

        self.testcase_parser.update_variables_binds(self.context.variables)

    def parse_testcase(self, testcase):
        """ parse testcase with variables binds if it is a template.
        """
        self.pre_config(testcase)

        parsed_testcase = self.testcase_parser.parse(testcase)
        return parsed_testcase

    def run_test(self, testcase):
        """ run single testcase.
        """
        testcase = self.parse_testcase(testcase)

        req_kwargs = testcase['request']

        try:
            url = req_kwargs.pop('url')
            method = req_kwargs.pop('method')
        except KeyError:
            raise exception.ParamsError("URL or METHOD missed!")

        resp_obj = self.client.request(url=url, method=method, **req_kwargs)
        diff_content = utils.diff_response(resp_obj, testcase['response'])
        success = False if diff_content else True
        return success, diff_content

    def run_testsets(self, testsets):
        """ run testcase suite.
        @testsets
            [
                {
                    "config": {
                        "requires": [],
                        "function_binds": {},
                        "variable_binds": []
                    }
                },
                {
                    "test": {
                        "variable_binds": {}, # override
                        "request": {},
                        "response": {}
                    }
                }
            ]
        """
        results = []
        for item in testsets:
            for key in item:
                if key == "config":
                    config_dict = item[key]
                    self.pre_config(config_dict)
                elif key == "test":
                    testcase = item[key]
                    result = self.run_test(testcase)
                    results.append(result)

        return results
