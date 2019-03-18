import os
import time

import requests
from httprunner import context, exceptions, loader, response, utils
from tests.base import ApiServerUnittest


class TestContext(ApiServerUnittest):

    def setUp(self):
        loader.load_project_tests(os.path.join(os.getcwd(), "tests"))
        project_mapping = loader.project_mapping
        self.context = context.SessionContext(
            functions=project_mapping["functions"],
            variables={"SECRET_KEY": "DebugTalk"}
        )

    def test_init_context_functions(self):
        context_functions = self.context.FUNCTIONS_MAPPING
        self.assertIn("gen_md5", context_functions)

    def test_init_test_variables_initialize(self):
        self.assertEqual(
            self.context.test_variables_mapping,
            {'SECRET_KEY': 'DebugTalk'}
        )

    def test_init_test_variables(self):
        variables = {
            "random": "${gen_random_string($num)}",
            "authorization": "${gen_md5($TOKEN, $data, $random)}",
            "data": '{"name": "$username", "password": "123456"}',
            "TOKEN": "debugtalk",
            "username": "user1",
            "num": 6
        }
        self.context.init_test_variables(variables)
        variables_mapping = self.context.test_variables_mapping
        self.assertEqual(len(variables_mapping["random"]), 6)
        self.assertEqual(len(variables_mapping["authorization"]), 32)
        self.assertEqual(variables_mapping["data"], '{"name": "user1", "password": "123456"}')

    def test_update_seesion_variables(self):
        self.context.update_session_variables({"TOKEN": "debugtalk"})
        self.assertEqual(
            self.context.session_variables_mapping["TOKEN"],
            "debugtalk"
        )

    def test_eval_content_functions(self):
        content = "${sleep_N_secs(1)}"
        start_time = time.time()
        self.context.eval_content(content)
        elapsed_time = time.time() - start_time
        self.assertGreater(elapsed_time, 1)

    def test_eval_content_variables(self):
        content = "abc$SECRET_KEY"
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
        parsed_request = self.context.eval_content(request)
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

    def test_do_validation(self):
        self.context._do_validation(
            {"check": "check", "check_value": 1, "expect": 1, "comparator": "eq"}
        )
        self.context._do_validation(
            {"check": "check", "check_value": "abc", "expect": "abc", "comparator": "=="}
        )
        self.context._do_validation(
            {"check": "status_code", "check_value": "201", "expect": 3, "comparator": "sum_status_code"}
        )

    def test_validate(self):
        url = "http://127.0.0.1:5000/"
        resp = requests.get(url)
        resp_obj = response.ResponseObject(resp)

        validators = [
            {"eq": ["$resp_status_code", 201]},
            {"check": "$resp_status_code", "comparator": "eq", "expect": 201},
            {"check": "$resp_body_success", "comparator": "eq", "expect": True}
        ]
        variables = {
            "resp_status_code": 200,
            "resp_body_success": True
        }

        self.context.init_test_variables(variables)

        with self.assertRaises(exceptions.ValidationFailure):
            self.context.validate(validators, resp_obj)

        validators = [
            {"eq": ["$resp_status_code", 201]},
            {"check": "$resp_status_code", "comparator": "eq", "expect": 201},
            {"check": "$resp_body_success", "comparator": "eq", "expect": True},
            {"check": "${is_status_code_200($resp_status_code)}", "comparator": "eq", "expect": False}
        ]
        variables = [
            {"resp_status_code": 201},
            {"resp_body_success": True}
        ]
        self.context.init_test_variables(variables)
        self.context.validate(validators, resp_obj)

        self.context.validate([], resp_obj)
        self.assertEqual(self.context.validation_results, [])

    def test_validate_exception(self):
        url = "http://127.0.0.1:5000/"
        resp = requests.get(url)
        resp_obj = response.ResponseObject(resp)

        # expected value missed in validators
        validators = [
            {"eq": ["$resp_status_code", 201]},
            {"check": "$resp_status_code", "comparator": "eq", "expect": 201}
        ]
        variables = []
        self.context.init_test_variables(variables)

        with self.assertRaises(exceptions.VariableNotFound):
            self.context.validate(validators, resp_obj)

        # expected value missed in variables mapping
        variables = [
            {"resp_status_code": 200}
        ]
        self.context.init_test_variables(variables)

        with self.assertRaises(exceptions.ValidationFailure):
            self.context.validate(validators, resp_obj)
