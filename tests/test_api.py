import os
import re
import shutil
import time
import unittest

from httprunner import loader, parser
from httprunner.api import HttpRunner, prepare_locust_tests
from tests.api_server import HTTPBIN_SERVER
from tests.base import ApiServerUnittest


class TestHttpRunner(ApiServerUnittest):

    def setUp(self):
        self.testcase_cli_path = "tests/data/demo_testcase_cli.yml"
        self.testcase_file_path_list = [
            os.path.join(
                os.getcwd(), 'tests/data/demo_testcase_hardcode.yml'),
            os.path.join(
                os.getcwd(), 'tests/data/demo_testcase_hardcode.json')
        ]
        testcases = [{
            'config': {
                'name': 'testcase description',
                'request': {
                    'base_url': '',
                    'headers': {'User-Agent': 'python-requests/2.18.4'}
                },
                'variables': []
            },
            "teststeps": [
                {
                    'name': '/api/get-token',
                    'request': {
                        'url': 'http://127.0.0.1:5000/api/get-token',
                        'method': 'POST',
                        'headers': {'Content-Type': 'application/json', 'app_version': '2.8.6', 'device_sn': 'FwgRiO7CNA50DSU', 'os_platform': 'ios', 'user_agent': 'iOS/10.3'},
                        'json': {'sign': '958a05393efef0ac7c0fb80a7eac45e24fd40c27'}
                    },
                    'extract': [
                        {'token': 'content.token'}
                    ],
                    'validate': [
                        {'eq': ['status_code', 200]},
                        {'eq': ['headers.Content-Type', 'application/json']},
                        {'eq': ['content.success', True]}
                    ]
                },
                {
                    'name': '/api/users/1000',
                    'request': {
                        'url': 'http://127.0.0.1:5000/api/users/1000',
                        'method': 'POST',
                        'headers': {'Content-Type': 'application/json', 'device_sn': 'FwgRiO7CNA50DSU','token': '$token'}, 'json': {'name': 'user1', 'password': '123456'}
                    },
                    'validate': [
                        {'eq': ['status_code', 201]},
                        {'eq': ['headers.Content-Type', 'application/json']},
                        {'eq': ['content.success', True]},
                        {'eq': ['content.msg', 'user created successfully.']}
                    ]
                }
            ]
        }]
        self.tests_mapping = {
            "testcases": testcases
        }
        self.runner = HttpRunner(failfast=True)
        self.reset_all()

    def reset_all(self):
        url = "%s/api/reset-all" % self.host
        headers = self.get_authenticated_headers()
        return self.api_client.get(url, headers=headers)

    def test_text_run_times(self):
        self.runner.run(self.testcase_cli_path)
        self.assertEqual(self.runner.summary["stat"]["testcases"]["total"], 1)
        self.assertEqual(self.runner.summary["stat"]["teststeps"]["total"], 10)

    def test_text_skip(self):
        self.runner.run(self.testcase_cli_path)
        self.assertEqual(self.runner.summary["stat"]["teststeps"]["skipped"], 4)

    def test_save_variables_output(self):
        testcases = [
            {
                "config": {
                    'name': "post data",
                    'variables': {
                        "var1": "abc",
                        "var2": "def"
                    },
                    "output": ["status_code", "req_data"]
                },
                "teststeps": [
                    {
                        "name": "post data",
                        "request": {
                            "url": "{}/post".format(HTTPBIN_SERVER),
                            "method": "POST",
                            "headers": {
                                "User-Agent": "python-requests/2.18.4",
                                "Content-Type": "application/json"
                            },
                            "data": "$var1"
                        },
                        "extract": {
                            "status_code": "status_code",
                            "req_data": "content.data"
                        },
                        "validate": [
                            {"eq": ["status_code", 200]}
                        ]
                    }
                ]
            }
        ]
        tests_mapping = {
            "testcases": testcases
        }
        self.runner.run_tests(tests_mapping)
        vars_out = self.runner.get_vars_out()
        self.assertIsInstance(vars_out, list)
        self.assertEqual(vars_out[0]["in"]["var1"], "abc")
        self.assertEqual(vars_out[0]["in"]["var2"], "def")
        self.assertEqual(vars_out[0]["out"]["status_code"], 200)
        self.assertEqual(vars_out[0]["out"]["req_data"], "abc")

    def test_save_variables_output_with_parameters(self):
        testcase_file_path = os.path.join(
            os.getcwd(), 'tests/testsuites/create_users_with_parameters.yml')
        self.runner.run(testcase_file_path)
        vars_out = self.runner.get_vars_out()
        self.assertIsInstance(vars_out, list)
        self.assertEqual(len(vars_out), 6)
        self.assertEqual(vars_out[0]["in"]["uid"], 101)
        self.assertEqual(vars_out[0]["in"]["device_sn"], "TESTSUITE_X1")
        token1 = vars_out[0]["out"]["token"]
        self.assertEqual(len(token1), 16)
        self.assertEqual(vars_out[5]["in"]["uid"], 103)
        self.assertEqual(vars_out[5]["in"]["device_sn"], "TESTSUITE_X2")
        token2 = vars_out[0]["out"]["token"]
        self.assertEqual(len(token2), 16)
        self.assertEqual(token1, token2)

    def test_html_report(self):
        report_save_dir = os.path.join(os.getcwd(), 'reports', "demo")
        runner = HttpRunner(failfast=True, report_dir=report_save_dir)
        runner.run(self.testcase_cli_path)
        summary = runner.summary
        self.assertEqual(summary["stat"]["testcases"]["total"], 1)
        self.assertEqual(summary["stat"]["teststeps"]["total"], 10)
        self.assertEqual(summary["stat"]["teststeps"]["skipped"], 4)
        self.assertGreater(len(os.listdir(report_save_dir)), 0)
        shutil.rmtree(report_save_dir)

    def test_log_file(self):
        log_file_path = os.path.join(os.getcwd(), 'reports', "test_log_file.log")
        runner = HttpRunner(failfast=True, log_file=log_file_path)
        runner.run(self.testcase_cli_path)
        self.assertTrue(os.path.isfile(log_file_path))
        os.remove(log_file_path)

    def test_run_testcases(self):
        self.runner.run_tests(self.tests_mapping)
        summary = self.runner.summary
        self.assertTrue(summary["success"])
        self.assertEqual(summary["stat"]["testcases"]["total"], 1)
        self.assertEqual(summary["stat"]["teststeps"]["total"], 2)
        self.assertIn("details", summary)
        self.assertIn("records", summary["details"][0])

    def test_run_yaml_upload(self):
        self.runner.run("tests/httpbin/upload.yml")
        summary = self.runner.summary
        self.assertTrue(summary["success"])
        self.assertEqual(summary["stat"]["testcases"]["total"], 1)
        self.assertEqual(summary["stat"]["teststeps"]["total"], 1)
        self.assertIn("details", summary)
        self.assertIn("records", summary["details"][0])

    def test_run_post_data(self):
        testcases = [
            {
                "config": {
                    'name': "post data",
                    'variables': []
                },
                "teststeps": [
                    {
                        "name": "post data",
                        "request": {
                            "url": "{}/post".format(HTTPBIN_SERVER),
                            "method": "POST",
                            "headers": {
                                "User-Agent": "python-requests/2.18.4",
                                "Content-Type": "application/json"
                            },
                            "data": "abc"
                        },
                        "validate": [
                            {"eq": ["status_code", 200]}
                        ]
                    }
                ]
            }
        ]
        tests_mapping = {
            "testcases": testcases
        }
        self.runner.run_tests(tests_mapping)
        summary = self.runner.summary
        self.assertTrue(summary["success"])
        self.assertEqual(summary["stat"]["testcases"]["total"], 1)
        self.assertEqual(summary["stat"]["teststeps"]["total"], 1)
        self.assertEqual(
            summary["details"][0]["records"][0]["meta_datas"]["data"][0]["response"]["json"]["data"],
            "abc"
        )

    def test_html_report_repsonse_image(self):
        report_save_dir = os.path.join(os.getcwd(), 'reports', "demo")
        runner = HttpRunner(failfast=True, report_dir=report_save_dir)
        report = runner.run("tests/httpbin/load_image.yml")
        self.assertTrue(os.path.isfile(report))
        shutil.rmtree(report_save_dir)

    def test_testcase_layer_with_api(self):
        self.runner.run("tests/testcases/setup.yml")
        summary = self.runner.summary
        self.assertTrue(summary["success"])
        self.assertEqual(summary["details"][0]["records"][0]["name"], "get token (setup)")
        self.assertEqual(summary["stat"]["testcases"]["total"], 1)
        self.assertEqual(summary["stat"]["teststeps"]["total"], 2)

    def test_testcase_layer_with_testcase(self):
        self.runner.run("tests/testsuites/create_users.yml")
        summary = self.runner.summary
        self.assertTrue(summary["success"])
        self.assertEqual(summary["stat"]["testcases"]["total"], 2)
        self.assertEqual(summary["stat"]["teststeps"]["total"], 8)

    def test_run_httprunner_with_hooks(self):
        testcase_file_path = os.path.join(
            os.getcwd(), 'tests/httpbin/hooks.yml')
        start_time = time.time()
        self.runner.run(testcase_file_path)
        end_time = time.time()
        summary = self.runner.summary
        self.assertTrue(summary["success"])
        self.assertLess(end_time - start_time, 60)

    def test_run_httprunner_with_teardown_hooks_alter_response(self):
        testcases = [
            {
                "config": {"name": "test teardown hooks"},
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
        loader.load_project_tests("tests")
        tests_mapping = {
            "project_mapping": loader.project_mapping,
            "testcases": testcases
        }
        self.runner.run_tests(tests_mapping)
        summary = self.runner.summary
        self.assertTrue(summary["success"])

    def test_run_httprunner_with_teardown_hooks_not_exist_attribute(self):
        testcases = [
            {
                "config": {
                    "name": "test teardown hooks"
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
        loader.load_project_tests("tests")
        tests_mapping = {
            "project_mapping": loader.project_mapping,
            "testcases": testcases
        }
        self.runner.run_tests(tests_mapping)
        summary = self.runner.summary
        self.assertFalse(summary["success"])
        self.assertEqual(summary["stat"]["teststeps"]["errors"], 1)

    def test_run_httprunner_with_teardown_hooks_error(self):
        testcases = [
            {
                "config": {
                    "name": "test teardown hooks"
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
        loader.load_project_tests("tests")
        tests_mapping = {
            "project_mapping": loader.project_mapping,
            "testcases": testcases
        }
        self.runner.run_tests(tests_mapping)
        summary = self.runner.summary
        self.assertFalse(summary["success"])
        self.assertEqual(summary["stat"]["teststeps"]["errors"], 1)

    def test_run_api(self):
        path = "tests/httpbin/api/get_headers.yml"
        self.runner.run(path)
        summary = self.runner.summary
        self.assertTrue(summary["success"])
        self.assertEqual(summary["stat"]["testcases"]["total"], 1)
        self.assertEqual(summary["stat"]["teststeps"]["total"], 1)
        self.assertEqual(summary["stat"]["teststeps"]["successes"], 1)

    def test_request_302_logs(self):
        path = "tests/httpbin/api/302_redirect.yml"
        self.runner.run(path)
        summary = self.runner.summary
        self.assertTrue(summary["success"])
        self.assertEqual(summary["stat"]["testcases"]["total"], 1)
        self.assertEqual(summary["stat"]["teststeps"]["total"], 1)
        self.assertEqual(summary["stat"]["teststeps"]["successes"], 1)

        req_resp_data = summary["details"][0]["records"][0]["meta_datas"]["data"]
        self.assertEqual(len(req_resp_data), 2)
        self.assertEqual(req_resp_data[0]["response"]["status_code"], 302)
        self.assertEqual(req_resp_data[1]["response"]["status_code"], 200)

    def test_request_with_params(self):
        path = "tests/httpbin/api/302_redirect.yml"
        self.runner.run(path)
        summary = self.runner.summary
        self.assertTrue(summary["success"])
        self.assertEqual(summary["stat"]["testcases"]["total"], 1)
        self.assertEqual(summary["stat"]["teststeps"]["total"], 1)
        self.assertEqual(summary["stat"]["teststeps"]["successes"], 1)

        req_resp_data = summary["details"][0]["records"][0]["meta_datas"]["data"]
        self.assertEqual(len(req_resp_data), 2)
        self.assertIn(
            "url=https%3A%2F%2Fdebugtalk.com",
            req_resp_data[0]["request"]["url"]
        )

    def test_run_api_folder(self):
        api_folder = "tests/httpbin/api/"
        self.runner.run(api_folder)
        summary = self.runner.summary
        self.assertTrue(summary["success"])
        self.assertEqual(summary["stat"]["testcases"]["total"], 2)
        self.assertEqual(summary["stat"]["teststeps"]["total"], 2)
        self.assertEqual(summary["stat"]["teststeps"]["successes"], 2)
        self.assertEqual(len(summary["details"]), 2)
        self.assertEqual(summary["details"][0]["stat"]["total"], 1)
        self.assertEqual(summary["details"][1]["stat"]["total"], 1)

    def test_run_testcase_hardcode(self):
        for testcase_file_path in self.testcase_file_path_list:
            self.runner.run(testcase_file_path)
            summary = self.runner.summary
            self.assertTrue(summary["success"])
            self.assertEqual(summary["stat"]["testcases"]["total"], 1)
            self.assertEqual(summary["stat"]["teststeps"]["total"], 3)
            self.assertEqual(summary["stat"]["teststeps"]["successes"], 3)


    def test_run_testcase_template_variables(self):
        testcase_file_path = os.path.join(
            os.getcwd(), 'tests/data/demo_testcase_variables.yml')
        self.runner.run(testcase_file_path)
        summary = self.runner.summary
        self.assertTrue(summary["success"])

    def test_run_testcase_template_import_functions(self):
        testcase_file_path = os.path.join(
            os.getcwd(), 'tests/data/demo_testcase_functions.yml')
        self.runner.run(testcase_file_path)
        summary = self.runner.summary
        self.assertTrue(summary["success"])

    def test_run_testcase_layered(self):
        testcase_file_path = os.path.join(
            os.getcwd(), 'tests/data/demo_testcase_layer.yml')
        self.runner.run(testcase_file_path)
        summary = self.runner.summary
        self.assertTrue(summary["success"])
        self.assertEqual(len(summary["details"]), 1)

    def test_run_testcase_output(self):
        testcase_file_path = os.path.join(
            os.getcwd(), 'tests/data/demo_testcase_layer.yml')
        self.runner.run(testcase_file_path)
        summary = self.runner.summary
        self.assertTrue(summary["success"])
        self.assertIn("token", summary["details"][0]["in_out"]["out"])
        # TODO: add
        # self.assertIn("user_agent", summary["details"][0]["in_out"]["in"])

    def test_run_testcase_with_variables_mapping(self):
        testcase_file_path = os.path.join(
            os.getcwd(), 'tests/data/demo_testcase_layer.yml')
        variables_mapping = {
            "app_version": '2.9.7'
        }
        self.runner.run(testcase_file_path, mapping=variables_mapping)
        summary = self.runner.summary
        self.assertTrue(summary["success"])
        self.assertIn("token", summary["details"][0]["in_out"]["out"])
        # TODO: add
        # self.assertGreater(len(summary["details"][0]["in_out"]["in"]), 3)

    def test_run_testcase_with_parameters(self):
        testcase_file_path = os.path.join(
            os.getcwd(), 'tests/testsuites/create_users_with_parameters.yml')
        self.runner.run(testcase_file_path)
        summary = self.runner.summary
        self.assertTrue(summary["success"])
        self.assertEqual(len(summary["details"]), 3 * 2)

        self.assertEqual(summary["stat"]["testcases"]["total"], 6)
        self.assertEqual(summary["stat"]["teststeps"]["total"], 3 * 2 * 4)
        self.assertEqual(
            summary["details"][0]["name"],
            "create user 101 and check result for TESTSUITE_X1."
        )
        self.assertEqual(
            summary["details"][5]["name"],
            "create user 103 and check result for TESTSUITE_X2."
        )
        self.assertEqual(
            summary["details"][0]["stat"]["total"],
            4
        )
        records_name_list = [
            summary["details"][i]["records"][2]["name"]
            for i in range(6)
        ]
        self.assertEqual(
            set(records_name_list),
            {
                "create user 101 for TESTSUITE_X1",
                "create user 101 for TESTSUITE_X2",
                "create user 102 for TESTSUITE_X1",
                "create user 102 for TESTSUITE_X2",
                "create user 103 for TESTSUITE_X1",
                "create user 103 for TESTSUITE_X2"
            }
        )

    # def test_validate_response_content(self):
    #     # TODO: fix compatibility with Python 2.7
    #     testcase_file_path = os.path.join(
    #         os.getcwd(), 'tests/httpbin/basic.yml')
    #     self.runner.run(testcase_file_path)
    #     self.assertTrue(self.runner.summary["success"])

    def test_html_report_xss(self):
        testcases = [
            {
                "config": {
                    'name': "post data"
                },
                "teststeps": [
                    {
                        "name": "post data",
                        "request": {
                            "url": "{}/anything".format(HTTPBIN_SERVER),
                            "method": "POST",
                            "headers": {
                                "Content-Type": "application/json"
                            },
                            "json": {
                                'success': False,
                                "person": "<img src=x onerror=alert(1)>"
                            }
                        },
                        "validate": [
                            {"eq": ["status_code", 200]}
                        ]
                    }
                ]
            }
        ]
        tests_mapping = {
            "testcases": testcases
        }
        report_path = self.runner.run(tests_mapping)
        with open(report_path) as f:
            content = f.read()
            m = re.findall(
                re.escape("&#34;person&#34;: &#34;&lt;img src=x onerror=alert(1)&gt;&#34;"),
                content
            )
            self.assertEqual(len(m), 2)


class TestApi(ApiServerUnittest):

    def test_testcase_loader(self):
        testcase_path = "tests/testcases/setup.yml"
        tests_mapping = loader.load_tests(testcase_path)

        project_mapping = tests_mapping["project_mapping"]
        self.assertIsInstance(project_mapping, dict)
        self.assertIn("PWD", project_mapping)
        self.assertIn("functions", project_mapping)
        self.assertIn("env", project_mapping)

        testcases = tests_mapping["testcases"]
        self.assertIsInstance(testcases, list)
        self.assertEqual(len(testcases), 1)
        testcase_config = testcases[0]["config"]
        self.assertEqual(testcase_config["name"], "setup and reset all.")
        self.assertIn("path", testcases[0])

        testcase_tests = testcases[0]["teststeps"]
        self.assertEqual(len(testcase_tests), 2)
        self.assertIn("api", testcase_tests[0])
        self.assertEqual(testcase_tests[0]["name"], "get token (setup)")
        self.assertIsInstance(testcase_tests[0]["variables"], dict)
        self.assertIn("api_def", testcase_tests[0])
        self.assertEqual(testcase_tests[0]["api_def"]["request"]["url"], "/api/get-token")

    def test_testcase_parser(self):
        testcase_path = "tests/testcases/setup.yml"
        tests_mapping = loader.load_tests(testcase_path)

        parsed_tests_mapping = parser.parse_tests(tests_mapping)
        parsed_testcases = parsed_tests_mapping["testcases"]

        self.assertEqual(len(parsed_testcases), 1)

        self.assertIn("variables", parsed_testcases[0]["config"])
        self.assertEqual(len(parsed_testcases[0]["teststeps"]), 2)

        test_dict1 = parsed_testcases[0]["teststeps"][0]
        self.assertEqual(test_dict1["name"], "get token (setup)")
        self.assertNotIn("api_def", test_dict1)
        self.assertEqual(test_dict1["variables"]["device_sn"], "TESTCASE_SETUP_XXX")
        self.assertEqual(test_dict1["request"]["url"], "http://127.0.0.1:5000/api/get-token")
        self.assertEqual(test_dict1["request"]["verify"], False)

        test_dict2 = parsed_testcases[0]["teststeps"][1]
        self.assertEqual(test_dict2["request"]["verify"], False)

    def test_testcase_add_tests(self):
        testcase_path = "tests/testcases/setup.yml"
        tests_mapping = loader.load_tests(testcase_path)

        parsed_tests_mapping = parser.parse_tests(tests_mapping)
        runner = HttpRunner()
        test_suite = runner._add_tests(parsed_tests_mapping)

        self.assertEqual(len(test_suite._tests), 1)
        teststeps = test_suite._tests[0].teststeps
        self.assertEqual(teststeps[0]["name"], "get token (setup)")
        self.assertEqual(teststeps[0]["variables"]["device_sn"], "TESTCASE_SETUP_XXX")
        self.assertIn("api", teststeps[0])

    def test_testcase_complex_verify(self):
        testcase_path = "tests/testcases/create_and_check.yml"
        tests_mapping = loader.load_tests(testcase_path)
        parsed_tests_mapping = parser.parse_tests(tests_mapping)
        teststeps = parsed_tests_mapping["testcases"][0]["teststeps"]

        # testcases/setup.yml
        teststep1 = teststeps[0]
        self.assertEqual(teststep1["teststeps"][0]["request"]["verify"], False)
        self.assertEqual(teststep1["teststeps"][1]["request"]["verify"], False)

        # testcases/create_and_check.yml teststep 2/3/4
        self.assertEqual(teststeps[1]["request"]["verify"], True)
        self.assertEqual(teststeps[2]["request"]["verify"], True)
        self.assertEqual(teststeps[3]["request"]["verify"], True)

    def test_testcase_simple_run_suite(self):
        testcase_path = "tests/testcases/setup.yml"
        tests_mapping = loader.load_tests(testcase_path)
        parsed_tests_mapping = parser.parse_tests(tests_mapping)
        runner = HttpRunner()
        test_suite = runner._add_tests(parsed_tests_mapping)
        tests_results = runner._run_suite(test_suite)
        self.assertEqual(len(tests_results[0][1].records), 2)

    def test_testcase_complex_run_suite(self):
        testcase_path = "tests/testcases/create_and_check.yml"
        tests_mapping = loader.load_tests(testcase_path)
        parsed_tests_mapping = parser.parse_tests(tests_mapping)
        runner = HttpRunner()
        test_suite = runner._add_tests(parsed_tests_mapping)
        tests_results = runner._run_suite(test_suite)
        self.assertEqual(len(tests_results[0][1].records), 4)

        results = tests_results[0][1]
        self.assertEqual(
            results.records[0]["name"],
            "setup and reset all (override) for TESTCASE_CREATE_XXX."
        )
        self.assertEqual(
            results.records[1]["name"],
            "make sure user 9001 does not exist"
        )

    def test_testsuite_loader(self):
        testcase_path = "tests/testsuites/create_users.yml"
        tests_mapping = loader.load_tests(testcase_path)

        project_mapping = tests_mapping["project_mapping"]
        self.assertIsInstance(project_mapping, dict)
        self.assertIn("PWD", project_mapping)
        self.assertIn("functions", project_mapping)
        self.assertIn("env", project_mapping)

        testsuites = tests_mapping["testsuites"]
        self.assertIsInstance(testsuites, list)
        self.assertEqual(len(testsuites), 1)

        self.assertIn("path", testsuites[0])
        testsuite_config = testsuites[0]["config"]
        self.assertEqual(testsuite_config["name"], "create users with uid")

        testcases = testsuites[0]["testcases"]
        self.assertEqual(len(testcases), 2)
        self.assertIn("create user 1000 and check result.", testcases)
        testcase_tests = testcases["create user 1000 and check result."]
        self.assertIn("testcase_def", testcase_tests)
        self.assertEqual(testcase_tests["name"], "create user 1000 and check result.")
        self.assertIsInstance(testcase_tests["testcase_def"], dict)
        self.assertEqual(testcase_tests["testcase_def"]["config"]["name"], "create user and check result.")
        self.assertEqual(len(testcase_tests["testcase_def"]["teststeps"]), 4)
        self.assertEqual(
            testcase_tests["testcase_def"]["teststeps"][0]["name"],
            "setup and reset all (override) for $device_sn."
        )

    def test_testsuite_parser(self):
        testcase_path = "tests/testsuites/create_users.yml"
        tests_mapping = loader.load_tests(testcase_path)

        parsed_tests_mapping = parser.parse_tests(tests_mapping)

        parsed_testcases = parsed_tests_mapping["testcases"]
        self.assertEqual(len(parsed_testcases), 2)
        self.assertEqual(len(parsed_testcases[0]["teststeps"]), 4)

        testcase1 = parsed_testcases[0]["teststeps"][0]
        self.assertIn("setup and reset all (override)", testcase1["config"]["name"])
        self.assertEqual(testcase1["teststeps"][0]["variables"]["var_c"], testcase1["teststeps"][0]["variables"]["var_d"])
        self.assertEqual(testcase1["teststeps"][0]["variables"]["var_a"], testcase1["teststeps"][0]["variables"]["var_b"])
        self.assertNotEqual(testcase1["teststeps"][0]["variables"]["var_a"], testcase1["teststeps"][0]["variables"]["var_c"])
        self.assertNotIn("testcase_def", testcase1)
        self.assertEqual(len(testcase1["teststeps"]), 2)
        self.assertEqual(
            testcase1["teststeps"][0]["request"]["url"],
            "http://127.0.0.1:5000/api/get-token"
        )
        self.assertEqual(len(testcase1["teststeps"][0]["variables"]["device_sn"]), 15)

    def test_testsuite_add_tests(self):
        testcase_path = "tests/testsuites/create_users.yml"
        tests_mapping = loader.load_tests(testcase_path)

        parsed_tests_mapping = parser.parse_tests(tests_mapping)
        runner = HttpRunner()
        test_suite = runner._add_tests(parsed_tests_mapping)

        self.assertEqual(len(test_suite._tests), 2)
        tests = test_suite._tests[0].teststeps
        self.assertIn("setup and reset all (override)", tests[0]["config"]["name"])

    def test_testsuite_run_suite(self):
        testcase_path = "tests/testsuites/create_users.yml"
        tests_mapping = loader.load_tests(testcase_path)

        parsed_tests_mapping = parser.parse_tests(tests_mapping)

        runner = HttpRunner()
        test_suite = runner._add_tests(parsed_tests_mapping)
        tests_results = runner._run_suite(test_suite)

        self.assertEqual(len(tests_results[0][1].records), 4)

        results = tests_results[0][1]
        self.assertIn(
            "setup and reset all (override)",
            results.records[0]["name"]
        )
        self.assertIn(
            results.records[1]["name"],
            ["make sure user 1000 does not exist", "make sure user 1001 does not exist"]
        )


class TestLocust(unittest.TestCase):

    def test_prepare_locust_tests(self):
        path = os.path.join(
            os.getcwd(), 'tests/locust_tests/demo_locusts.yml')
        locust_tests = prepare_locust_tests(path)
        self.assertIn("gen_md5", locust_tests["functions"])
        self.assertEqual(len(locust_tests["tests"]), 2 + 3)
        name_list = [
            "create user 1000 and check result.",
            "create user 1001 and check result."
        ]
        self.assertIn(locust_tests["tests"][0]["config"]["name"], name_list)
        self.assertIn(locust_tests["tests"][4]["config"]["name"], name_list)
