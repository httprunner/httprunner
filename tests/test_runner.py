import os
import time

from httprunner import exceptions, loader, runner
from httprunner.utils import deep_update_dict
from tests.api_server import HTTPBIN_SERVER
from tests.base import ApiServerUnittest


class TestRunner(ApiServerUnittest):

    def setUp(self):
        loader.load_project_tests(os.path.join(os.getcwd(), "tests"))
        self.debugtalk_module = loader.project_mapping["debugtalk"]
        config_dict = {
            "variables": self.debugtalk_module["variables"],
            "functions": self.debugtalk_module["functions"]
        }
        self.test_runner = runner.Runner(config_dict)
        self.reset_all()

    def reset_all(self):
        url = "%s/api/reset-all" % self.host
        headers = self.get_authenticated_headers()
        return self.api_client.get(url, headers=headers)

    def test_run_single_testcase(self):
        testcase_file_path_list = [
            os.path.join(
                os.getcwd(), 'tests/data/demo_testset_hardcode.yml'),
            os.path.join(
                os.getcwd(), 'tests/data/demo_testset_hardcode.json')
        ]

        for testcase_file_path in testcase_file_path_list:
            testcases = loader.load_file(testcase_file_path)

            config_dict = {
                "variables": self.debugtalk_module["variables"],
                "functions": self.debugtalk_module["functions"]
            }
            test_runner = runner.Runner(config_dict)

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

    def test_run_testset_with_hooks(self):
        start_time = time.time()

        config_dict = {
            "name": "basic test with httpbin",
            "variables": self.debugtalk_module["variables"],
            "functions": self.debugtalk_module["functions"],
            "request": {
                "base_url": HTTPBIN_SERVER
            },
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
        test_runner = runner.Runner(config_dict)
        end_time = time.time()
        # check if testset setup hook executed
        self.assertGreater(end_time - start_time, 0.5)

        start_time = time.time()
        test_runner.run_test(test)
        test_runner.run_test(test)
        end_time = time.time()
        # testset teardown hook has not been executed now
        self.assertLess(end_time - start_time, 1)

    def test_run_testset_with_hooks_modify_request(self):
        config_dict = {
            "name": "basic test with httpbin",
            "variables": self.debugtalk_module["variables"],
            "functions": self.debugtalk_module["functions"],
            "request": {
                "base_url": HTTPBIN_SERVER
            }
        }
        test = {
            "name": "modify request headers",
            "request": {
                "url": "/anything",
                "method": "POST",
                "headers": {
                    "content-type": "application/json",
                    "user_agent": "iOS/10.3",
                    "os_platform": "ios"
                },
                "json": {
                    "sign": "f1219719911caae89ccc301679857ebfda115ca2"
                }
            },
            "setup_hooks": [
                "${modify_headers_os_platform($request, android)}"
            ],
            "validate": [
                {"check": "status_code", "expect": 200},
                {"check": "content.headers.Os-Platform", "expect": "android"}
            ]
        }
        test_runner = runner.Runner(config_dict)
        test_runner.run_test(test)

    def test_run_testset_with_teardown_hooks_success(self):
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
        config_dict = {}
        self.test_runner.init_test(config_dict, "testcase")

        start_time = time.time()
        self.test_runner.run_test(test)
        end_time = time.time()
        # check if teardown function executed
        self.assertLess(end_time - start_time, 0.5)

    def test_run_testset_with_teardown_hooks_fail(self):
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
        config_dict = {}
        self.test_runner.init_test(config_dict, "testcase")

        start_time = time.time()
        self.test_runner.run_test(test)
        end_time = time.time()
        # check if teardown function executed
        self.assertGreater(end_time - start_time, 2)

    def test_run_testcase_with_empty_header(self):
        testcase_file_path = os.path.join(
            os.getcwd(), 'tests/data/test_bugfix.yml')
        testsets = loader.load_testcases(testcase_file_path)
        testset = testsets[0]
        config_dict_headers = testset["config"]["request"]["headers"]
        test_dict_headers = testset["teststeps"][0]["request"]["headers"]
        headers = deep_update_dict(
            config_dict_headers,
            test_dict_headers
        )
        self.assertEqual(headers["Content-Type"], "application/json")

    def test_bugfix_type_match(self):
        testcase_file_path = os.path.join(
            os.getcwd(), 'tests/data/test_bugfix.yml')
        testcases = loader.load_file(testcase_file_path)
        config_dict = {}
        self.test_runner.init_test(config_dict, "testcase")

        test = testcases[2]["test"]
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
