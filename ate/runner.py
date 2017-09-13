from ate import exception, response, utils
from ate.client import HttpSession
from ate.context import Context


class Runner(object):

    def __init__(self, http_client_session=None):
        self.http_client_session = http_client_session
        self.context = Context()

    def init_config(self, config_dict, level):
        """ create/update context variables binds
        @param (dict) config_dict
        @param (str) level, "testset" or "testcase"
        testset:
            {
                "name": "smoke testset",
                "path": "tests/data/demo_testset_variables.yml",
                "requires": [],         # optional
                "function_binds": {},   # optional
                "import_module_items": [],  # optional
                "variable_binds": [],   # optional
                "request": {
                    "base_url": "http://127.0.0.1:5000",
                    "headers": {
                        "User-Agent": "iOS/2.8.3"
                    }
                }
            }
        testcase:
            {
                "name": "testcase description",
                "requires": [],         # optional
                "function_binds": {},   # optional
                "import_module_items": [],  # optional
                "variable_binds": [],   # optional
                "request": {
                    "url": "/api/get-token",
                    "method": "POST",
                    "headers": {
                        "Content-Type": "application/json"
                    }
                },
                "json": {
                    "sign": "f1219719911caae89ccc301679857ebfda115ca2"
                }
            }
        @param (str) context level, testcase or testset
        """
        # convert keys in request headers to lowercase
        config_dict = utils.lower_dict_key(config_dict)

        self.context.init_context(level)
        self.context.config_context(config_dict, level)

        request_config = config_dict.get('request', {})
        parsed_request = self.context.get_parsed_request(request_config, level)

        base_url = parsed_request.pop("base_url", None)
        self.http_client_session = self.http_client_session or HttpSession(base_url)

        return parsed_request

    def run_test(self, testcase):
        """ run single testcase.
        @param (dict) testcase
            {
                "name": "testcase description",
                "times": 3,
                "requires": [],  # optional, override
                "function_binds": {}, # optional, override
                "variable_binds": {}, # optional, override
                "request": {
                    "url": "http://127.0.0.1:5000/api/users/1000",
                    "method": "POST",
                    "headers": {
                        "Content-Type": "application/json",
                        "authorization": "$authorization",
                        "random": "$random"
                    },
                    "body": '{"name": "user", "password": "123456"}'
                },
                "extract_binds": [], # optional
                "validators": [],    # optional
                "setup": [],         # optional
                "teardown": []       # optional
            }
        @return True or raise exception during test
        """
        parsed_request = self.init_config(testcase, level="testcase")

        try:
            url = parsed_request.pop('url')
            method = parsed_request.pop('method')
        except KeyError:
            raise exception.ParamsError("URL or METHOD missed!")

        run_times = int(testcase.get("times", 1))
        extract_binds = testcase.get("extract_binds", [])
        validators = testcase.get("validators", [])
        setup_actions = testcase.get("setup", [])
        teardown_actions = testcase.get("teardown", [])

        def setup_teardown(actions):
            for action in actions:
                self.context.exec_content_functions(action)

        for _ in range(run_times):
            setup_teardown(setup_actions)

            resp = self.http_client_session.request(url=url, method=method, **parsed_request)
            resp_obj = response.ResponseObject(resp)

            extracted_variables_mapping_list = resp_obj.extract_response(extract_binds)
            self.context.bind_variables(extracted_variables_mapping_list, level="testset")

            resp_obj.validate(validators, self.context.get_testcase_variables_mapping())

            setup_teardown(teardown_actions)

        return True

    def run_testset(self, testset):
        """ run single testset, including one or several testcases.
        @param (dict) testset
            {
                "name": "testset description",
                "config": {
                    "name": "testset description",
                    "requires": [],
                    "function_binds": {},
                    "variable_binds": [],
                    "request": {}
                },
                "testcases": [
                    {
                        "name": "testcase description",
                        "variable_binds": {}, # optional, override
                        "request": {},
                        "extract_binds": {},  # optional
                        "validators": {}      # optional
                    },
                    testcase12
                ]
            }
        @return (list) test results of testcases
            [
                True,    # testcase11
                True     # testcase12
            ]
        """
        results = []

        config_dict = testset.get("config", {})
        self.init_config(config_dict, level="testset")
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
                    True,    # testcase11
                    True     # testcase12
                ],
                [   # testset2
                    True,    # testcase21
                    True     # testcase22
                ]
            ]
        """
        return [self.run_testset(testset) for testset in testsets]
