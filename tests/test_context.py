import os

from httprunner import context, exceptions, loader, parser, runner
from tests.api_server import gen_md5
from tests.base import ApiServerUnittest, gen_random_string


class TestContext(ApiServerUnittest):

    def setUp(self):
        loader.load_project_data(os.path.join(os.getcwd(), "tests"))
        self.context = context.SessionContext(
            variables={"SECRET_KEY": "DebugTalk"}
        )

    def test_init_test_variables_initialize(self):
        self.assertEqual(
            self.context.test_variables_mapping,
            {'SECRET_KEY': 'DebugTalk'}
        )

    def test_init_test_variables(self):
        variables = {
            "random": "${gen_random_string($num)}",
            "authorization": "${gen_md5($TOKEN, $data, $random)}",
            "data": "$username",
            # TODO: escape '{' and '}'
            # "data": '{"name": "$username", "password": "123456"}',
            "TOKEN": "debugtalk",
            "username": "user1",
            "num": 6
        }
        functions = {
            "gen_random_string": gen_random_string,
            "gen_md5": gen_md5
        }
        variables = parser.prepare_lazy_data(variables, functions, variables.keys())
        variables = parser.parse_variables_mapping(variables)
        self.context.init_test_variables(variables)
        variables_mapping = self.context.test_variables_mapping
        self.assertEqual(len(variables_mapping["random"]), 6)
        self.assertEqual(len(variables_mapping["authorization"]), 32)
        self.assertEqual(variables_mapping["data"], 'user1')

    def test_update_seesion_variables(self):
        self.context.update_session_variables({"TOKEN": "debugtalk"})
        self.assertEqual(
            self.context.session_variables_mapping["TOKEN"],
            "debugtalk"
        )

    def test_eval_content_variables(self):
        variables = {
            "SECRET_KEY": "DebugTalk"
        }
        content = parser.prepare_lazy_data("abc$SECRET_KEY", {}, variables.keys())
        self.assertEqual(
            self.context.eval_content(content),
            "abcDebugTalk"
        )

        # TODO: fix variable extraction
        # content = "abc$SECRET_KEYdef"
        # self.assertEqual(
        #     self.context.eval_content(content),
        #     "abcDebugTalkdef"
        # )

    def test_get_parsed_request(self):
        variables = {
            "random": "${gen_random_string(5)}",
            "data": '{"name": "user", "password": "123456"}',
            "authorization": "${gen_md5($TOKEN, $data, $random)}",
            "TOKEN": "debugtalk"
        }
        functions = {
            "gen_random_string": gen_random_string,
            "gen_md5": gen_md5
        }
        variables = parser.prepare_lazy_data(variables, functions, variables.keys())
        variables = parser.parse_variables_mapping(variables)
        self.context.init_test_variables(variables)

        request = {
            "url": "http://127.0.0.1:5000/api/users/1000",
            "method": "POST",
            "headers": {
                "Content-Type": "application/json",
                "authorization": "$authorization",
                "random": "$random",
                "secret_key": "$SECRET_KEY"
            },
            "data": "$data"
        }
        prepared_request = parser.prepare_lazy_data(
            request,
            functions,
            {"authorization", "random", "SECRET_KEY", "data"}
        )
        parsed_request = self.context.eval_content(prepared_request)
        self.assertIn("authorization", parsed_request["headers"])
        self.assertEqual(len(parsed_request["headers"]["authorization"]), 32)
        self.assertIn("random", parsed_request["headers"])
        self.assertEqual(len(parsed_request["headers"]["random"]), 5)
        self.assertIn("data", parsed_request)
        self.assertEqual(
            parsed_request["data"],
            '{"name": "user", "password": "123456"}'
        )
        self.assertEqual(parsed_request["headers"]["secret_key"], "DebugTalk")

    def test_validate(self):
        testcases = [
            {
                "config": {
                    'name': "test validation"
                },
                "teststeps": [
                    {
                        "name": "test validation",
                        "request": {
                            "url": "http://127.0.0.1:5000/",
                            "method": "GET",
                        },
                        "variables": {
                            "resp_status_code": 200,
                            "resp_body_success": True
                        },
                        "validate": [
                            {"eq": ["$resp_status_code", 200]},
                            {"check": "$resp_status_code", "comparator": "eq", "expect": 200},
                            {"check": "$resp_body_success", "expect": True},
                            {"check": "${is_status_code_200($resp_status_code)}", "expect": True}
                        ]
                    }
                ]
            }
        ]
        from tests.debugtalk import is_status_code_200
        tests_mapping = {
            "project_mapping": {
                "functions": {
                    "is_status_code_200": is_status_code_200
                }
            },
            "testcases": testcases
        }
        testcases = parser.parse_tests(tests_mapping)
        parsed_testcase = testcases[0]
        test_runner = runner.Runner(parsed_testcase["config"])
        teststep = parsed_testcase["teststeps"][0]
        test_runner.run_test(teststep)

    def test_validate_exception(self):
        testcases = [
            {
                "config": {
                    'name': "test validation"
                },
                "teststeps": [
                    {
                        "name": "test validation",
                        "request": {
                            "url": "http://127.0.0.1:5000/",
                            "method": "GET",
                        },
                        "variables": {
                            "resp_status_code": 200,
                            "resp_body_success": True
                        },
                        "validate": [
                            {"eq": ["$resp_status_code", 201]},
                            {"check": "$resp_status_code", "expect": 201},
                            {"check": "$resp_body_success", "comparator": "eq", "expect": True}
                        ]
                    }
                ]
            }
        ]
        tests_mapping = {
            "testcases": testcases
        }
        testcases = parser.parse_tests(tests_mapping)
        parsed_testcase = testcases[0]
        test_runner = runner.Runner(parsed_testcase["config"])
        teststep = parsed_testcase["teststeps"][0]
        with self.assertRaises(exceptions.ValidationFailure):
            test_runner.run_test(teststep)
