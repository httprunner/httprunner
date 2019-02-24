import os
import time

from httprunner import exceptions, loader, runner
from httprunner.utils import deep_update_dict
from tests.api_server import HTTPBIN_SERVER
from tests.base import ApiServerUnittest


class TestRunner(ApiServerUnittest):

    def setUp(self):
        loader.load_project_tests(os.path.join(os.getcwd(), "tests"))
        project_mapping = loader.project_mapping
        self.debugtalk_functions = project_mapping["functions"]

        config = {
            "name": "XXX",
            "base_url": "http://127.0.0.1",
            "verify": False
        }
        self.test_runner = runner.Runner(config, self.debugtalk_functions)
        self.reset_all()

    def reset_all(self):
        url = "%s/api/reset-all" % self.host
        headers = self.get_authenticated_headers()
        return self.api_client.get(url, headers=headers)

    def test_run_single_testcase(self):
        testcase_file_path_list = [
            os.path.join(
                os.getcwd(), 'tests/data/demo_testcase_hardcode.yml'),
            os.path.join(
                os.getcwd(), 'tests/data/demo_testcase_hardcode.json')
        ]

        for testcase_file_path in testcase_file_path_list:
            testcases = loader.load_file(testcase_file_path)

            config_dict = {}
            test_runner = runner.Runner(config_dict, self.debugtalk_functions)

            test = testcases[0]["test"]
            test_runner.run_test(test)

            test = testcases[1]["test"]
            test_runner.run_test(test)

            test = testcases[2]["test"]
            test_runner.run_test(test)

    def test_run_single_testcase_fail(self):
        test = {
            "name": "get token",
            "request": {
                "url": "http://127.0.0.1:5000/api/get-token",
                "method": "POST",
                "headers": {
                    "content-type": "application/json",
                    "user_agent": "iOS/10.3",
                    "device_sn": "HZfFBh6tU59EdXJ",
                    "os_platform": "ios",
                    "app_version": "2.8.6"
                },
                "json": {
                    "sign": "f1219719911caae89ccc301679857ebfda115ca2"
                }
            },
            "validate": [
                {"check": "status_code", "expect": 205},
                {"check": "content.token", "comparator": "len_eq", "expect": 19}
            ]
        }

        with self.assertRaises(exceptions.ValidationFailure):
            self.test_runner.run_test(test)

    def test_run_testcase_with_hooks(self):
        start_time = time.time()

        config_dict = {
            "name": "basic test with httpbin",
            "base_url": HTTPBIN_SERVER,
            "setup_hooks": [
                "${sleep_N_secs(0.5)}"
                "${hook_print(setup)}"
            ],
            "teardown_hooks": [
                "${sleep_N_secs(1)}",
                "${hook_print(teardown)}"
            ]
        }
        test = {
            "name": "get token",
            "request": {
                "url": "http://127.0.0.1:5000/api/get-token",
                "method": "POST",
                "headers": {
                    "content-type": "application/json",
                    "user_agent": "iOS/10.3",
                    "device_sn": "HZfFBh6tU59EdXJ",
                    "os_platform": "ios",
                    "app_version": "2.8.6"
                },
                "json": {
                    "sign": "f1219719911caae89ccc301679857ebfda115ca2"
                }
            },
            "validate": [
                {"check": "status_code", "expect": 200}
            ]
        }
        test_runner = runner.Runner(config_dict, self.debugtalk_functions)
        end_time = time.time()
        # check if testcase setup hook executed
        self.assertGreater(end_time - start_time, 0.5)

        start_time = time.time()
        test_runner.run_test(test)
        test_runner.run_test(test)
        end_time = time.time()
        # testcase teardown hook has not been executed now
        self.assertLess(end_time - start_time, 1)

    def test_run_testcase_with_hooks_assignment(self):
        config_dict = {
            "name": "basic test with httpbin",
            "base_url": HTTPBIN_SERVER
        }
        test = {
            "name": "modify request headers",
            "request": {
                "url": "/anything",
                "method": "POST",
                "headers": {
                    "user_agent": "iOS/10.3",
                    "os_platform": "ios"
                },
                "data": "a=1&b=2"
            },
            "setup_hooks": [
                {"total": "${sum_two(1, 5)}"}
            ],
            "validate": [
                {"check": "status_code", "expect": 200}
            ]
        }
        test_runner = runner.Runner(config_dict, self.debugtalk_functions)
        test_runner.run_test(test)
        test_variables_mapping = test_runner.session_context.test_variables_mapping
        self.assertEqual(test_variables_mapping["total"], 6)
        self.assertEqual(test_variables_mapping["request"]["data"], "a=1&b=2")

    def test_run_testcase_with_hooks_modify_request(self):
        config_dict = {
            "name": "basic test with httpbin",
            "base_url": HTTPBIN_SERVER
        }
        test = {
            "name": "modify request headers",
            "request": {
                "url": "/anything",
                "method": "POST",
                "headers": {
                    "content-type": "application/json",
                    "user_agent": "iOS/10.3"
                },
                "json": {
                    "os_platform": "ios",
                    "sign": "f1219719911caae89ccc301679857ebfda115ca2"
                }
            },
            "setup_hooks": [
                "${modify_request_json($request, android)}"
            ],
            "validate": [
                {"check": "status_code", "expect": 200},
                {"check": "content.json.os_platform", "expect": "android"}
            ]
        }
        test_runner = runner.Runner(config_dict, self.debugtalk_functions)
        test_runner.run_test(test)

    def test_run_testcase_with_teardown_hooks_success(self):
        test = {
            "name": "get token",
            "request": {
                "url": "http://127.0.0.1:5000/api/get-token",
                "method": "POST",
                "headers": {
                    "content-type": "application/json",
                    "user_agent": "iOS/10.3",
                    "device_sn": "HZfFBh6tU59EdXJ",
                    "os_platform": "ios",
                    "app_version": "2.8.6"
                },
                "json": {
                    "sign": "f1219719911caae89ccc301679857ebfda115ca2"
                }
            },
            "validate": [
                {"check": "status_code", "expect": 200}
            ],
            "teardown_hooks": ["${teardown_hook_sleep_N_secs($response, 2)}"]
        }
        start_time = time.time()
        self.test_runner.run_test(test)
        end_time = time.time()
        # check if teardown function executed
        self.assertLess(end_time - start_time, 0.5)

    def test_run_testcase_with_teardown_hooks_fail(self):
        test = {
            "name": "get token",
            "request": {
                "url": "http://127.0.0.1:5000/api/get-token2",
                "method": "POST",
                "headers": {
                    "content-type": "application/json",
                    "user_agent": "iOS/10.3",
                    "device_sn": "HZfFBh6tU59EdXJ",
                    "os_platform": "ios",
                    "app_version": "2.8.6"
                },
                "json": {
                    "sign": "f1219719911caae89ccc301679857ebfda115ca2"
                }
            },
            "validate": [
                {"check": "status_code", "expect": 404}
            ],
            "teardown_hooks": ["${teardown_hook_sleep_N_secs($response, 2)}"]
        }
        start_time = time.time()
        self.test_runner.run_test(test)
        end_time = time.time()
        # check if teardown function executed
        self.assertGreater(end_time - start_time, 2)

    def test_bugfix_type_match(self):
        testcase_file_path = os.path.join(
            os.getcwd(), 'tests/data/bugfix_type_match.yml')
        testcases = loader.load_file(testcase_file_path)

        test = testcases[1]["test"]
        self.test_runner.run_test(test)

    def test_run_validate_elapsed(self):
        test = {
            "name": "get token",
            "request": {
                "url": "http://127.0.0.1:5000/api/get-token",
                "method": "POST",
                "headers": {
                    "content-type": "application/json",
                    "user_agent": "iOS/10.3",
                    "device_sn": "HZfFBh6tU59EdXJ",
                    "os_platform": "ios",
                    "app_version": "2.8.6"
                },
                "json": {
                    "sign": "f1219719911caae89ccc301679857ebfda115ca2"
                }
            },
            "validate": [
                {"check": "status_code", "expect": 200},
                {"check": "elapsed.seconds", "comparator": "lt", "expect": 1},
                {"check": "elapsed.days", "comparator": "eq", "expect": 0},
                {"check": "elapsed.microseconds", "comparator": "gt", "expect": 1000},
                {"check": "elapsed.total_seconds", "comparator": "lt", "expect": 1}
            ]
        }
        self.test_runner.run_test(test)
