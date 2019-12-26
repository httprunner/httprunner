import json
import os
import re
import shutil
import time

from httprunner import exceptions, loader, parser, report
from httprunner.api import HttpRunner
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
                        'headers': {'Content-Type': 'application/json', 'app_version': '2.8.6',
                                    'device_sn': 'FwgRiO7CNA50DSU', 'os_platform': 'ios', 'user_agent': 'iOS/10.3'},
                        'json': {'sign': '9c0c7e51c91ae963c833a4ccbab8d683c4a90c98'}
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
                        'headers': {'Content-Type': 'application/json',
                                    'device_sn': 'FwgRiO7CNA50DSU','token': '$token'},
                        'json': {'name': 'user1', 'password': '123456'}
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
        summary = self.runner.run(self.testcase_cli_path)
        self.assertEqual(summary["stat"]["testcases"]["total"], 1)
        self.assertEqual(summary["stat"]["teststeps"]["total"], 10)

    def test_text_run_times_invalid(self):
        testcases = [
            {
                "config": {
                    'name': "post data",
                    'variables': []
                },
                "teststeps": [
                    {
                        "name": "post data",
                        "times": "1.5",
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
        with self.assertRaises(exceptions.ParamsError):
            self.runner.run_tests(tests_mapping)

    def test_text_skip(self):
        summary = self.runner.run(self.testcase_cli_path)
        self.assertEqual(summary["stat"]["teststeps"]["skipped"], 4)

    def test_save_variables_output(self):
        testcases = [
            {
                "config": {
                    'name': "post data",
                    'variables': {
                        "var1": "abc",
                        "var2": "def"
                    },
                    "export": ["status_code", "req_data"]
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
        token1 = vars_out[0]["out"]["session_token"]
        self.assertEqual(len(token1), 16)
        self.assertEqual(vars_out[5]["in"]["uid"], 103)
        self.assertEqual(vars_out[5]["in"]["device_sn"], "TESTSUITE_X2")
        token2 = vars_out[0]["out"]["session_token"]
        self.assertEqual(len(token2), 16)
        self.assertEqual(token1, token2)

    def test_html_report(self):
        runner = HttpRunner(failfast=True)
        summary = runner.run(self.testcase_cli_path)
        self.assertEqual(summary["stat"]["testcases"]["total"], 1)
        self.assertEqual(summary["stat"]["teststeps"]["total"], 10)
        self.assertEqual(summary["stat"]["teststeps"]["skipped"], 4)

        report_save_dir = os.path.join(os.getcwd(), 'reports', "demo")
        report.gen_html_report(summary, report_dir=report_save_dir)
        self.assertGreater(len(os.listdir(report_save_dir)), 0)
        shutil.rmtree(report_save_dir)

    def test_html_report_with_fixed_report_file(self):
        runner = HttpRunner(failfast=True)
        summary = runner.run(self.testcase_cli_path)
        self.assertEqual(summary["stat"]["testcases"]["total"], 1)
        self.assertEqual(summary["stat"]["teststeps"]["total"], 10)
        self.assertEqual(summary["stat"]["teststeps"]["skipped"], 4)

        report_file = os.path.join(os.getcwd(), 'reports', "demo", "test.html")
        report.gen_html_report(summary, report_file=report_file)
        report_save_dir = os.path.dirname(report_file)
        self.assertEqual(len(os.listdir(report_save_dir)), 1)
        self.assertTrue(os.path.isfile(report_file))
        shutil.rmtree(report_save_dir)

    def test_log_file(self):
        log_file_path = os.path.join(os.getcwd(), 'reports', "test_log_file.log")
        runner = HttpRunner(failfast=True, log_file=log_file_path)
        runner.run(self.testcase_cli_path)
        self.assertTrue(os.path.isfile(log_file_path))
        os.remove(log_file_path)

    def test_run_testcases(self):
        summary = self.runner.run_tests(self.tests_mapping)
        self.assertTrue(summary["success"])
        self.assertEqual(summary["stat"]["testcases"]["total"], 1)
        self.assertEqual(summary["stat"]["teststeps"]["total"], 2)
        self.assertIn("details", summary)
        self.assertIn("records", summary["details"][0])

    def test_run_yaml_upload(self):
        upload_cases_list = [
            "tests/httpbin/upload.yml",
            "tests/httpbin/upload.v2.yml"
        ]
        for upload_case in upload_cases_list:
            summary = self.runner.run(upload_case)
            self.assertTrue(summary["success"])
            self.assertEqual(summary["stat"]["testcases"]["total"], 1)
            self.assertEqual(summary["stat"]["teststeps"]["total"], 2)
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
        summary = self.runner.run_tests(tests_mapping)
        self.assertTrue(summary["success"])
        self.assertEqual(summary["stat"]["testcases"]["total"], 1)
        self.assertEqual(summary["stat"]["teststeps"]["total"], 1)
        resp_json = json.loads(summary["details"][0]["records"][0]["meta_datas"]["data"][0]["response"]["body"])
        self.assertEqual(
            resp_json["data"],
            "abc"
        )

    def test_html_report_repsonse_image(self):
        runner = HttpRunner(failfast=True)
        summary = runner.run("tests/httpbin/load_image.yml")

        report_save_dir = os.path.join(os.getcwd(), 'reports', "demo")
        report_path = report.gen_html_report(summary, report_dir=report_save_dir)
        self.assertTrue(os.path.isfile(report_path))
        shutil.rmtree(report_save_dir)

    def test_testcase_layer_with_api(self):
        summary = self.runner.run("tests/testcases/setup.yml")
        self.assertTrue(summary["success"])
        self.assertEqual(summary["details"][0]["records"][0]["name"], "get token (setup)")
        self.assertEqual(summary["stat"]["testcases"]["total"], 1)
        self.assertEqual(summary["stat"]["teststeps"]["total"], 2)

    def test_testcase_layer_with_testcase(self):
        summary = self.runner.run("tests/testsuites/create_users.yml")
        self.assertTrue(summary["success"])
        self.assertEqual(summary["stat"]["testcases"]["total"], 2)
        self.assertEqual(summary["stat"]["teststeps"]["total"], 4)

    def test_validate_script(self):
        summary = self.runner.run("tests/httpbin/validate.yml")
        self.assertFalse(summary["success"])

    def test_run_httprunner_with_hooks(self):
        testcase_file_path = os.path.join(
            os.getcwd(), 'tests/httpbin/hooks.yml')
        start_time = time.time()
        summary = self.runner.run(testcase_file_path)
        end_time = time.time()
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

        tests_mapping = {
            "project_mapping": loader.load_project_data("tests"),
            "testcases": testcases
        }
        summary = self.runner.run_tests(tests_mapping)
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
        tests_mapping = {
            "project_mapping": loader.load_project_data("tests"),
            "testcases": testcases
        }
        summary = self.runner.run_tests(tests_mapping)
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
        tests_mapping = {
            "project_mapping": loader.load_project_data("tests"),
            "testcases": testcases
        }
        summary = self.runner.run_tests(tests_mapping)
        self.assertFalse(summary["success"])
        self.assertEqual(summary["stat"]["teststeps"]["errors"], 1)

    def test_run_api(self):
        path = "tests/httpbin/api/get_headers.yml"
        summary = self.runner.run(path)
        self.assertTrue(summary["success"])
        self.assertEqual(summary["stat"]["testcases"]["total"], 1)
        self.assertEqual(summary["stat"]["teststeps"]["total"], 1)
        self.assertEqual(summary["stat"]["teststeps"]["successes"], 1)

    def test_request_302_logs(self):
        path = "tests/httpbin/api/302_redirect.yml"
        summary = self.runner.run(path)
        self.assertTrue(summary["success"])
        self.assertEqual(summary["stat"]["testcases"]["total"], 1)
        self.assertEqual(summary["stat"]["teststeps"]["total"], 1)
        self.assertEqual(summary["stat"]["teststeps"]["successes"], 1)

        req_resp_data = summary["details"][0]["records"][0]["meta_datas"]["data"]
        self.assertEqual(len(req_resp_data), 2)
        self.assertEqual(req_resp_data[0]["response"]["status_code"], 302)
        self.assertEqual(req_resp_data[1]["response"]["status_code"], 200)

    def test_request_302_logs_teardown_hook(self):
        path = "tests/httpbin/api/302_redirect_teardown_hook.yml"
        summary = self.runner.run(path)
        self.assertTrue(summary["success"])
        self.assertEqual(summary["stat"]["testcases"]["total"], 1)
        self.assertEqual(summary["stat"]["teststeps"]["total"], 1)
        self.assertEqual(summary["stat"]["teststeps"]["successes"], 1)

        req_resp_data = summary["details"][0]["records"][0]["meta_datas"]["data"]
        self.assertEqual(len(req_resp_data), 2)
        self.assertEqual(req_resp_data[0]["response"]["status_code"], 302)
        self.assertEqual(req_resp_data[1]["response"]["status_code"], 500)

    def test_request_with_params(self):
        path = "tests/httpbin/api/302_redirect.yml"
        summary = self.runner.run(path)
        self.assertTrue(summary["success"])
        self.assertEqual(summary["stat"]["testcases"]["total"], 1)
        self.assertEqual(summary["stat"]["teststeps"]["total"], 1)
        self.assertEqual(summary["stat"]["teststeps"]["successes"], 1)

        req_resp_data = summary["details"][0]["records"][0]["meta_datas"]["data"]
        self.assertEqual(len(req_resp_data), 2)
        self.assertIn(
            "url=https%3A%2F%2Fgithub.com",
            req_resp_data[0]["request"]["url"]
        )

    def test_run_api_folder(self):
        api_folder = "tests/httpbin/api/"
        summary = self.runner.run(api_folder)
        print(summary["stat"]["testcases"]["total"])
        print(len(summary["details"]))
        self.assertTrue(summary["success"])
        self.assertEqual(summary["stat"]["testcases"]["total"], 3)
        self.assertEqual(summary["stat"]["teststeps"]["total"], 3)
        self.assertEqual(summary["stat"]["teststeps"]["successes"], 3)
        self.assertEqual(len(summary["details"]), 3)
        self.assertEqual(summary["details"][0]["stat"]["total"], 1)
        self.assertEqual(summary["details"][1]["stat"]["total"], 1)
        self.assertEqual(summary["details"][2]["stat"]["total"], 1)


    def test_run_testcase_hardcode(self):
        for testcase_file_path in self.testcase_file_path_list:
            summary = self.runner.run(testcase_file_path)
            self.assertTrue(summary["success"])
            self.assertEqual(summary["stat"]["testcases"]["total"], 1)
            self.assertEqual(summary["stat"]["teststeps"]["total"], 3)
            self.assertEqual(summary["stat"]["teststeps"]["successes"], 3)


    def test_run_testcase_template_variables(self):
        testcase_file_path = os.path.join(
            os.getcwd(), 'tests/data/demo_testcase_variables.yml')
        summary = self.runner.run(testcase_file_path)
        self.assertTrue(summary["success"])

    def test_run_testcase_template_import_functions(self):
        testcase_file_path = os.path.join(
            os.getcwd(), 'tests/data/demo_testcase_functions.yml')
        summary = self.runner.run(testcase_file_path)
        self.assertTrue(summary["success"])

    def test_run_testcase_layered(self):
        testcase_file_path = os.path.join(
            os.getcwd(), 'tests/data/demo_testcase_layer.yml')
        summary = self.runner.run(testcase_file_path)
        self.assertTrue(summary["success"])
        self.assertEqual(len(summary["details"]), 1)

    def test_run_testcase_output(self):
        testcase_file_path = os.path.join(
            os.getcwd(), 'tests/data/demo_testcase_layer.yml')
        summary = self.runner.run(testcase_file_path)
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
        summary = self.runner.run(testcase_file_path, mapping=variables_mapping)
        self.assertTrue(summary["success"])
        self.assertIn("token", summary["details"][0]["in_out"]["out"])
        # TODO: add
        # self.assertGreater(len(summary["details"][0]["in_out"]["in"]), 3)

    def test_run_testcase_with_parameters(self):
        testcase_file_path = os.path.join(
            os.getcwd(), 'tests/testsuites/create_users_with_parameters.yml')
        summary = self.runner.run(testcase_file_path)
        self.assertTrue(summary["success"])
        self.assertEqual(len(summary["details"]), 3 * 2)

        self.assertEqual(summary["stat"]["testcases"]["total"], 6)
        self.assertEqual(summary["stat"]["teststeps"]["total"], 3 * 2 * 2)
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
            2
        )
        records_name_list = [
            summary["details"][i]["records"][1]["meta_datas"][1]["name"]
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

    def test_validate_response_content(self):
        testcase_file_path = os.path.join(
            os.getcwd(), 'tests/httpbin/basic.yml')
        summary = self.runner.run(testcase_file_path)
        self.assertTrue(summary["success"])

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
        summary = self.runner.run(tests_mapping)
        report_path = report.gen_html_report(summary)
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
        tests_mapping = loader.load_cases(testcase_path)

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
        tests_mapping = loader.load_cases(testcase_path)

        parsed_testcases = parser.parse_tests(tests_mapping)

        self.assertEqual(len(parsed_testcases), 1)

        self.assertIn("variables", parsed_testcases[0]["config"])
        self.assertEqual(len(parsed_testcases[0]["teststeps"]), 2)

        test_dict1 = parsed_testcases[0]["teststeps"][0]
        self.assertEqual(test_dict1["name"], "get token (setup)")
        self.assertNotIn("api_def", test_dict1)
        self.assertEqual(test_dict1["variables"]["device_sn"], "TESTCASE_SETUP_XXX")
        self.assertEqual(test_dict1["request"]["url"], "/api/get-token")
        self.assertEqual(test_dict1["request"]["verify"], False)

        test_dict2 = parsed_testcases[0]["teststeps"][1]
        self.assertEqual(test_dict2["request"]["verify"], False)

    def test_testcase_add_tests(self):
        testcase_path = "tests/testcases/setup.yml"
        tests_mapping = loader.load_cases(testcase_path)

        testcases = parser.parse_tests(tests_mapping)
        runner = HttpRunner()
        test_suite = runner._add_tests(testcases)

        self.assertEqual(len(test_suite._tests), 1)
        teststeps = test_suite._tests[0].teststeps
        self.assertEqual(teststeps[0]["name"], "get token (setup)")
        self.assertEqual(teststeps[0]["variables"]["device_sn"], "TESTCASE_SETUP_XXX")
        self.assertIn("api", teststeps[0])

    def test_testcase_complex_verify(self):
        testcase_path = "tests/testcases/create_user.yml"
        tests_mapping = loader.load_cases(testcase_path)
        testcases = parser.parse_tests(tests_mapping)
        teststeps = testcases[0]["teststeps"]

        # testcases/setup.yml
        teststep0 = teststeps[0]
        self.assertEqual(teststep0["teststeps"][0]["request"]["verify"], False)
        self.assertEqual(teststep0["teststeps"][1]["request"]["verify"], False)

        # testcases/create_user.yml
        teststep1 = teststeps[1]
        self.assertEqual(teststep1["teststeps"][0]["request"]["verify"], True)
        self.assertEqual(teststep1["teststeps"][1]["request"]["verify"], True)
        self.assertEqual(teststep1["teststeps"][2]["request"]["verify"], True)

    def test_testcase_simple_run_suite(self):
        testcase_path = "tests/testcases/setup.yml"
        tests_mapping = loader.load_cases(testcase_path)
        testcases = parser.parse_tests(tests_mapping)
        runner = HttpRunner()
        test_suite = runner._add_tests(testcases)
        tests_results = runner._run_suite(test_suite)
        self.assertEqual(len(tests_results[0][1].records), 2)

    def test_testcase_complex_run_suite(self):
        for testcase_path in [
            "tests/testcases/create_user.yml",
            "tests/testcases/create_user.v2.yml",
            "tests/testcases/create_user.json",
            "tests/testcases/create_user.v2.json"
        ]:
            tests_mapping = loader.load_cases(testcase_path)
            testcases = parser.parse_tests(tests_mapping)
            runner = HttpRunner()
            test_suite = runner._add_tests(testcases)
            tests_results = runner._run_suite(test_suite)
            self.assertEqual(len(tests_results[0][1].records), 2)

            results = tests_results[0][1]
            self.assertEqual(
                results.records[0]["name"],
                "setup and reset all (override) for TESTCASE_CREATE_XXX."
            )
            self.assertEqual(
                results.records[1]["name"],
                "create user and check result."
            )

    def test_testsuite_loader(self):
        testcase_path = "tests/testsuites/create_users.yml"
        tests_mapping = loader.load_cases(testcase_path)

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
        self.assertEqual(len(testcase_tests["testcase_def"]["teststeps"]), 2)
        self.assertEqual(
            testcase_tests["testcase_def"]["teststeps"][0]["name"],
            "setup and reset all (override) for $device_sn."
        )

    def test_testsuite_parser(self):
        testcase_path = "tests/testsuites/create_users.yml"
        tests_mapping = loader.load_cases(testcase_path)

        parsed_testcases = parser.parse_tests(tests_mapping)
        self.assertEqual(len(parsed_testcases), 2)
        self.assertEqual(len(parsed_testcases[0]["teststeps"]), 2)

        testcase1 = parsed_testcases[0]["teststeps"][0]
        self.assertIn("setup and reset all (override)", testcase1["config"]["name"].raw_string)
        teststeps = testcase1["teststeps"]
        self.assertNotIn("testcase_def", testcase1)
        self.assertEqual(len(teststeps), 2)
        self.assertEqual(
            teststeps[0]["request"]["url"],
            "/api/get-token"
        )

    def test_testsuite_add_tests(self):
        testcase_path = "tests/testsuites/create_users.yml"
        tests_mapping = loader.load_cases(testcase_path)

        testcases = parser.parse_tests(tests_mapping)
        runner = HttpRunner()
        test_suite = runner._add_tests(testcases)

        self.assertEqual(len(test_suite._tests), 2)
        tests = test_suite._tests[0].teststeps
        self.assertIn("setup and reset all (override)", tests[0]["config"]["name"].raw_string)

    def test_testsuite_run_suite(self):
        testcase_path = "tests/testsuites/create_users.yml"
        tests_mapping = loader.load_cases(testcase_path)

        testcases = parser.parse_tests(tests_mapping)

        runner = HttpRunner()
        test_suite = runner._add_tests(testcases)
        tests_results = runner._run_suite(test_suite)

        self.assertEqual(len(tests_results[0][1].records), 2)

        results = tests_results[0][1]
        self.assertIn(
            "setup and reset all (override)",
            results.records[0]["name"]
        )
        self.assertEqual(
            results.records[1]["name"],
            "create user and check result."
        )
