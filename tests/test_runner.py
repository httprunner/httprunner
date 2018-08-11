import os
import time

from httprunner import HttpRunner, exceptions, loader, runner
from httprunner.utils import deep_update_dict
from tests.api_server import HTTPBIN_SERVER
from tests.base import ApiServerUnittest


class TestRunner(ApiServerUnittest):

    def setUp(self):
        self.test_runner = runner.Runner()
        self.reset_all()

        self.testcase_file_path_list = [
            os.path.join(
                os.getcwd(), 'tests/data/demo_testset_hardcode.yml'),
            os.path.join(
                os.getcwd(), 'tests/data/demo_testset_hardcode.json')
        ]

    def reset_all(self):
        url = "%s/api/reset-all" % self.host
        headers = self.get_authenticated_headers()
        return self.api_client.get(url, headers=headers)

    def test_run_single_testcase(self):
        for testcase_file_path in self.testcase_file_path_list:
            testcases = loader.load_file(testcase_file_path)

            config_dict = {
                "path": testcase_file_path
            }
            self.test_runner.init_config(config_dict, "testset")

            test = testcases[0]["test"]
            self.test_runner.run_test(test)

            test = testcases[1]["test"]
            self.test_runner.run_test(test)

            test = testcases[2]["test"]
            self.test_runner.run_test(test)

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
            "path": os.path.join(os.getcwd(), __file__),
            "name": "basic test with httpbin",
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
            "path": os.path.join(os.getcwd(), __file__),
            "name": "basic test with httpbin",
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

    def test_run_httprunner_with_hooks(self):
        testcase_file_path = os.path.join(
            os.getcwd(), 'tests/httpbin/hooks.yml')

        start_time = time.time()
        runner = HttpRunner().run(testcase_file_path)
        end_time = time.time()
        summary = runner.summary
        self.assertTrue(summary["success"])
        self.assertLess(end_time - start_time, 10)

    def test_run_httprunner_with_teardown_hooks_alter_response(self):
        testsets = [
            {
                "name": "test teardown hooks",
                "config": {
                    'path': 'tests/httpbin/hooks.yml',
                },
                "teststeps": [
                    {
                        "name": "test teardown hooks",
                        "request": {
                            "url": "{}/headers".format(HTTPBIN_SERVER),
                            "method": "GET",
                            "data": "abc"
                        },
                        "teardown_hooks": [
                            "${alter_response($response)}"
                        ],
                        "validate": [
                            {"eq": ["status_code", 500]},
                            {"eq": ["headers.content-type", "html/text"]},
                            {"eq": ["json.headers.Host", "127.0.0.1:8888"]},
                            {"eq": ["content.headers.Host", "127.0.0.1:8888"]},
                            {"eq": ["text.headers.Host", "127.0.0.1:8888"]},
                            {"eq": ["new_attribute", "new_attribute_value"]},
                            {"eq": ["new_attribute_dict", {"key": 123}]},
                            {"eq": ["new_attribute_dict.key", 123]}
                        ]
                    }
                ]
            }
        ]
        runner = HttpRunner().run(testsets)
        summary = runner.summary
        self.assertTrue(summary["success"])

    def test_run_httprunner_with_teardown_hooks_not_exist_attribute(self):
        testsets = [
            {
                "name": "test teardown hooks",
                "config": {
                    'path': 'tests/httpbin/hooks.yml',
                },
                "teststeps": [
                    {
                        "name": "test teardown hooks",
                        "request": {
                            "url": "{}/headers".format(HTTPBIN_SERVER),
                            "method": "GET",
                            "data": "abc"
                        },
                        "teardown_hooks": [
                            "${alter_response($response)}"
                        ],
                        "validate": [
                            {"eq": ["attribute_not_exist", "new_attribute"]}
                        ]
                    }
                ]
            }
        ]
        runner = HttpRunner().run(testsets)
        summary = runner.summary
        self.assertFalse(summary["success"])
        self.assertEqual(summary["stat"]["errors"], 1)

    def test_run_httprunner_with_teardown_hooks_error(self):
        testsets = [
            {
                "name": "test teardown hooks",
                "config": {
                    'path': 'tests/httpbin/hooks.yml',
                },
                "teststeps": [
                    {
                        "name": "test teardown hooks",
                        "request": {
                            "url": "{}/headers".format(HTTPBIN_SERVER),
                            "method": "GET",
                            "data": "abc"
                        },
                        "teardown_hooks": [
                            "${alter_response_error($response)}"
                        ]
                    }
                ]
            }
        ]
        runner = HttpRunner().run(testsets)
        summary = runner.summary
        self.assertFalse(summary["success"])
        self.assertEqual(summary["stat"]["errors"], 1)

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
        config_dict = {
            "path": os.path.join(os.getcwd(), __file__)
        }
        self.test_runner.init_config(config_dict, "testset")

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
        config_dict = {
            "path": os.path.join(os.getcwd(), __file__)
        }
        self.test_runner.init_config(config_dict, "testset")

        start_time = time.time()
        self.test_runner.run_test(test)
        end_time = time.time()
        # check if teardown function executed
        self.assertGreater(end_time - start_time, 2)

    def test_run_testset_hardcode(self):
        for testcase_file_path in self.testcase_file_path_list:
            runner = HttpRunner().run(testcase_file_path)
            self.assertTrue(runner.summary["success"])

    def test_run_testsets_hardcode(self):
        runner = HttpRunner().run(self.testcase_file_path_list)
        summary = runner.summary
        self.assertTrue(summary["success"])
        self.assertEqual(summary["stat"]["testsRun"], 6)
        self.assertEqual(summary["stat"]["successes"], 6)

    def test_run_testset_template_variables(self):
        testcase_file_path = os.path.join(
            os.getcwd(), 'tests/data/demo_testset_variables.yml')
        runner = HttpRunner().run(testcase_file_path)
        summary = runner.summary
        self.assertTrue(summary["success"])

    def test_run_testset_template_import_functions(self):
        testcase_file_path = os.path.join(
            os.getcwd(), 'tests/data/demo_testset_functions.yml')
        runner = HttpRunner().run(testcase_file_path)
        summary = runner.summary
        self.assertTrue(summary["success"])

    def test_run_testset_layered(self):
        testcase_file_path = os.path.join(
            os.getcwd(), 'tests/data/demo_testset_layer.yml')
        runner = HttpRunner().run(testcase_file_path)
        summary = runner.summary
        self.assertTrue(summary["success"])

    def test_run_testset_output(self):
        testcase_file_path = os.path.join(
            os.getcwd(), 'tests/data/demo_testset_layer.yml')
        runner = HttpRunner().run(testcase_file_path)
        summary = runner.summary
        self.assertTrue(summary["success"])
        self.assertIn("token", summary["details"][0]["output"][0]["out"])
        #TODO: fix
        self.assertEqual(len(summary["details"][0]["output"]), 3)

    def test_run_testset_with_variables_mapping(self):
        testcase_file_path = os.path.join(
            os.getcwd(), 'tests/data/demo_testset_layer.yml')
        variables_mapping = {
            "app_version": '2.9.7'
        }
        runner = HttpRunner().run(testcase_file_path, mapping=variables_mapping)
        summary = runner.summary
        self.assertTrue(summary["success"])
        self.assertIn("token", summary["details"][0]["output"][0]["out"])
        #TODO: fix
        self.assertEqual(len(summary["details"][0]["output"]), 3)

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
        config_dict = {
            "path": testcase_file_path
        }
        self.test_runner.init_config(config_dict, "testset")

        test = testcases[2]["test"]
        self.test_runner.run_test(test)

    def test_run_testset_with_parameters(self):
        testcase_file_path = os.path.join(
            os.getcwd(), 'tests/data/demo_parameters.yml')
        runner = HttpRunner().run(testcase_file_path)
        summary = runner.summary
        self.assertTrue(summary["success"])
        self.assertEqual(len(summary["details"][0]["output"]), 3 * 2 * 2)
        self.assertEqual(summary["stat"]["testsRun"], 3 * 2 * 2)

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
