import os
import time

from httprunner import loader, parser, runner
from tests.api_server import HTTPBIN_SERVER
from tests.base import ApiServerUnittest


class TestRunner(ApiServerUnittest):

    def setUp(self):
        project_mapping = loader.load_project_data(os.path.join(os.getcwd(), "tests"))
        self.debugtalk_functions = project_mapping["functions"]

        config = {
            "name": "XXX",
            "base_url": "http://127.0.0.1",
            "verify": False
        }
        self.test_runner = runner.Runner(config)
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
            tests_mapping = loader.load_cases(testcase_file_path)
            parsed_testcases = parser.parse_tests(tests_mapping)
            parsed_testcase = parsed_testcases[0]
            test_runner = runner.Runner(parsed_testcase["config"])
            test_runner.run_test(parsed_testcase["teststeps"][0])
            test_runner.run_test(parsed_testcase["teststeps"][1])
            test_runner.run_test(parsed_testcase["teststeps"][2])

    def test_run_testcase_with_hooks(self):
        start_time = time.time()

        testcases = [
            {
                "config": {
                    "name": "basic test with httpbin",
                    "base_url": HTTPBIN_SERVER,
                    "setup_hooks": [
                        "${sleep(0.5)}",
                        "${hook_print(setup)}"
                    ],
                    "teardown_hooks": [
                        "${sleep(1)}",
                        "${hook_print(teardown)}"
                    ]
                },
                "teststeps": [
                    {
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
                                "sign": "5188962c489d1a35effa99e9346dd5efd4fdabad"
                            }
                        },
                        "validate": [
                            {"check": "status_code", "expect": 200}
                        ]
                    }
                ]
            }
        ]
        tests_mapping = {
            "project_mapping": {
                "functions": self.debugtalk_functions
            },
            "testcases": testcases
        }
        parsed_testcases = parser.parse_tests(tests_mapping)
        parsed_testcase = parsed_testcases[0]
        test_runner = runner.Runner(parsed_testcase["config"])
        end_time = time.time()
        # check if testcase setup hook executed
        self.assertGreater(end_time - start_time, 0.5)

        start_time = time.time()
        test_runner.run_test(parsed_testcase["teststeps"][0])
        end_time = time.time()
        # testcase teardown hook has not been executed now
        self.assertLess(end_time - start_time, 1)

    def test_run_testcase_with_hooks_assignment(self):
        testcases = [
            {
                "config": {
                    "name": "basic test with httpbin",
                    "base_url": HTTPBIN_SERVER
                },
                "teststeps": [
                    {
                        "name": "modify request headers",
                        "base_url": HTTPBIN_SERVER,
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
                ]
            }
        ]
        tests_mapping = {
            "project_mapping": {
                "functions": self.debugtalk_functions
            },
            "testcases": testcases
        }
        parsed_testcases = parser.parse_tests(tests_mapping)
        parsed_testcase = parsed_testcases[0]
        test_runner = runner.Runner(parsed_testcase["config"])
        test_runner.run_test(parsed_testcase["teststeps"][0])
        test_variables_mapping = test_runner.session_context.test_variables_mapping
        self.assertEqual(test_variables_mapping["total"], 6)
        self.assertEqual(test_variables_mapping["request"]["data"], "a=1&b=2")

    def test_run_testcase_with_hooks_modify_request(self):
        testcases = [
            {
                "config": {
                    "name": "basic test with httpbin",
                    "base_url": HTTPBIN_SERVER
                },
                "teststeps": [
                    {
                        "name": "modify request headers",
                        "base_url": HTTPBIN_SERVER,
                        "request": {
                            "url": "/anything",
                            "method": "POST",
                            "headers": {
                                "content-type": "application/json",
                                "user_agent": "iOS/10.3"
                            },
                            "json": {
                                "os_platform": "ios",
                                "sign": "5188962c489d1a35effa99e9346dd5efd4fdabad"
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
                ]
            }
        ]
        tests_mapping = {
            "project_mapping": {
                "functions": self.debugtalk_functions
            },
            "testcases": testcases
        }
        parsed_testcases = parser.parse_tests(tests_mapping)
        parsed_testcase = parsed_testcases[0]
        test_runner = runner.Runner(parsed_testcase["config"])
        test_runner.run_test(parsed_testcase["teststeps"][0])

    def test_run_testcase_with_teardown_hooks_success(self):
        testcases = [
            {
                "config": {
                    "name": "basic test with httpbin"
                },
                "teststeps": [
                    {
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
                                "sign": "5188962c489d1a35effa99e9346dd5efd4fdabad"
                            }
                        },
                        "validate": [
                            {"check": "status_code", "expect": 200}
                        ],
                        "teardown_hooks": ["${teardown_hook_sleep_N_secs($response, 2)}"]
                    }
                ]
            }
        ]
        tests_mapping = {
            "project_mapping": {
                "functions": self.debugtalk_functions
            },
            "testcases": testcases
        }
        parsed_testcases = parser.parse_tests(tests_mapping)
        parsed_testcase = parsed_testcases[0]
        test_runner = runner.Runner(parsed_testcase["config"])

        start_time = time.time()
        test_runner.run_test(parsed_testcase["teststeps"][0])
        end_time = time.time()
        # check if teardown function executed
        self.assertLess(end_time - start_time, 0.5)

    def test_run_testcase_with_teardown_hooks_fail(self):
        testcases = [
            {
                "config": {
                    "name": "basic test with httpbin"
                },
                "teststeps": [
                    {
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
                                "sign": "5188962c489d1a35effa99e9346dd5efd4fdabad"
                            }
                        },
                        "validate": [
                            {"check": "status_code", "expect": 404}
                        ],
                        "teardown_hooks": ["${teardown_hook_sleep_N_secs($response, 2)}"]
                    }
                ]
            }
        ]
        tests_mapping = {
            "project_mapping": {
                "functions": self.debugtalk_functions
            },
            "testcases": testcases
        }
        parsed_testcases = parser.parse_tests(tests_mapping)
        parsed_testcase = parsed_testcases[0]
        test_runner = runner.Runner(parsed_testcase["config"])

        start_time = time.time()
        test_runner.run_test(parsed_testcase["teststeps"][0])
        end_time = time.time()
        # check if teardown function executed
        self.assertGreater(end_time - start_time, 2)

    def test_bugfix_type_match(self):
        testcase_file_path = os.path.join(
            os.getcwd(), 'tests/data/bugfix_type_match.yml')
        tests_mapping = loader.load_cases(testcase_file_path)
        parsed_testcases = parser.parse_tests(tests_mapping)
        parsed_testcase = parsed_testcases[0]
        test_runner = runner.Runner(parsed_testcase["config"])
        test_runner.run_test(parsed_testcase["teststeps"][0])

    def test_run_validate_elapsed(self):
        testcases = [
            {
                "config": {},
                "teststeps": [
                    {
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
                                "sign": "5188962c489d1a35effa99e9346dd5efd4fdabad"
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
                ]
            }
        ]
        tests_mapping = {
            "project_mapping": {
                "functions": self.debugtalk_functions
            },
            "testcases": testcases
        }
        parsed_testcases = parser.parse_tests(tests_mapping)
        parsed_testcase = parsed_testcases[0]
        test_runner = runner.Runner(parsed_testcase["config"])
        test_runner.run_test(parsed_testcase["teststeps"][0])

    def test_run_testcase_config_variables_parsed_from_function(self):
        testcases = [
            {
                "config": {
                    "name": "basic test with httpbin",
                    "base_url": HTTPBIN_SERVER,
                    "variables": "${gen_variables()}"
                },
                "teststeps": [
                    {
                        "name": "modify request headers",
                        "base_url": HTTPBIN_SERVER,
                        "request": {
                            "url": "/anything",
                            "method": "POST",
                            "headers": {
                                "user_agent": "iOS/10.3",
                                "os_platform": "ios"
                            },
                            "data": "a=1&b=2"
                        },
                        "validate": [
                            {"check": "status_code", "expect": 200}
                        ]
                    }
                ]
            }
        ]
        tests_mapping = {
            "project_mapping": {
                "functions": self.debugtalk_functions
            },
            "testcases": testcases
        }
        parsed_testcases = parser.parse_tests(tests_mapping)
        parsed_testcase = parsed_testcases[0]
        test_runner = runner.Runner(parsed_testcase["config"])
        test_runner.run_test(parsed_testcase["teststeps"][0])
        test_variables_mapping = test_runner.session_context.test_variables_mapping
        self.assertEqual(test_variables_mapping["var_a"], 1)
        self.assertEqual(test_variables_mapping["var_b"], 2)
        self.assertEqual(test_variables_mapping["request"]["data"], "a=1&b=2")
