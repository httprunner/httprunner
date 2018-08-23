import os
import time

import requests
from httprunner import context, exceptions, loader, response
from tests.base import ApiServerUnittest


class TestContext(ApiServerUnittest):

    def setUp(self):
        loader.load_project_tests(os.path.join(os.getcwd(), "tests"))
        self.debugtalk_module = loader.project_mapping["debugtalk"]

        self.context = context.Context(
            self.debugtalk_module["variables"],
            self.debugtalk_module["functions"]
        )
        testcase_file_path = os.path.join(os.getcwd(), 'tests/data/demo_binds.yml')
        self.testcases = loader.load_file(testcase_file_path)

    def test_init_context_functions(self):
        context_functions = self.context.TESTCASE_SHARED_FUNCTIONS_MAPPING
        self.assertIn("gen_md5", context_functions)
        self.assertIn("equals", context_functions)

    def test_init_context_variables(self):
        self.assertEqual(
            self.context.teststep_variables_mapping["SECRET_KEY"],
            "DebugTalk"
        )
        self.assertEqual(
            self.context.testcase_runtime_variables_mapping["SECRET_KEY"],
            "DebugTalk"
        )

    def test_update_context_testcase_level(self):
        variables = [
            {"TOKEN": "debugtalk"},
            {"data": '{"name": "user", "password": "123456"}'}
        ]
        self.context.update_context_variables(variables, "testcase")
        self.assertEqual(
            self.context.teststep_variables_mapping["TOKEN"],
            "debugtalk"
        )
        self.assertEqual(
            self.context.testcase_runtime_variables_mapping["TOKEN"],
            "debugtalk"
        )

    def test_update_context_teststep_level(self):
        variables = [
            {"TOKEN": "debugtalk"},
            {"data": '{"name": "user", "password": "123456"}'}
        ]
        self.context.update_context_variables(variables, "teststep")
        self.assertEqual(
            self.context.teststep_variables_mapping["TOKEN"],
            "debugtalk"
        )
        self.assertNotIn(
            "TOKEN",
            self.context.testcase_runtime_variables_mapping
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

    def test_update_testcase_runtime_variables_mapping(self):
        variables = {"abc": 123}
        self.context.update_testcase_runtime_variables_mapping(variables)
        self.assertEqual(
            self.context.testcase_runtime_variables_mapping["abc"],
            123
        )
        self.assertEqual(
            self.context.teststep_variables_mapping["abc"],
            123
        )

    def test_update_teststep_variables_mapping(self):
        self.context.update_teststep_variables_mapping("abc", 123)
        self.assertEqual(
            self.context.teststep_variables_mapping["abc"],
            123
        )
        self.assertNotIn(
            "abc",
            self.context.testcase_runtime_variables_mapping
        )

    def test_get_parsed_request(self):
        variables = [
            {"TOKEN": "debugtalk"},
            {"random": "${gen_random_string(5)}"},
            {"data": '{"name": "user", "password": "123456"}'},
            {"authorization": "${gen_md5($TOKEN, $data, $random)}"}
        ]
        self.context.update_context_variables(variables, "teststep")

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
        parsed_request = self.context.get_parsed_request(request, level="teststep")
        self.assertIn("authorization", parsed_request["headers"])
        self.assertEqual(len(parsed_request["headers"]["authorization"]), 32)
        self.assertIn("random", parsed_request["headers"])
        self.assertEqual(len(parsed_request["headers"]["random"]), 5)
        self.assertIn("data", parsed_request)
        self.assertEqual(parsed_request["data"], variables[2]["data"])
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
        variables = [
            {"resp_status_code": 200},
            {"resp_body_success": True}
        ]
        self.context.update_context_variables(variables, "teststep")

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
        self.context.update_context_variables(variables, "teststep")
        self.context.validate(validators, resp_obj)

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
        self.context.update_context_variables(variables, "teststep")

        with self.assertRaises(exceptions.VariableNotFound):
            self.context.validate(validators, resp_obj)

        # expected value missed in variables mapping
        variables = [
            {"resp_status_code": 200}
        ]
        self.context.update_context_variables(variables, "teststep")

        with self.assertRaises(exceptions.ValidationFailure):
            self.context.validate(validators, resp_obj)
