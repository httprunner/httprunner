import os
import shutil
import time

from httprunner import HttpRunner, LocustRunner, loader
from locust import HttpLocust
from tests.api_server import HTTPBIN_SERVER
from tests.base import ApiServerUnittest


class TestHttpRunner(ApiServerUnittest):

    def setUp(self):
        self.testset_path = "tests/data/demo_testset_cli.yml"
        self.testcase_file_path_list = [
            os.path.join(
                os.getcwd(), 'tests/data/demo_testset_hardcode.yml'),
            os.path.join(
                os.getcwd(), 'tests/data/demo_testset_hardcode.json')
        ]
        self.testcase = {
            'name': 'testset description',
            'config': {
                'name': 'testset description',
                'request': {
                    'base_url': '',
                    'headers': {'User-Agent': 'python-requests/2.18.4'}
                },
                'variables': [],
                'output': ['token']
            },
            'api': {},
            'teststeps': [
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
        }
        self.reset_all()

    def reset_all(self):
        url = "%s/api/reset-all" % self.host
        headers = self.get_authenticated_headers()
        return self.api_client.get(url, headers=headers)

    def test_text_run_times(self):
        runner = HttpRunner().run(self.testset_path)
        self.assertEqual(runner.summary["stat"]["testsRun"], 10)

    def test_text_skip(self):
        runner = HttpRunner().run(self.testset_path)
        self.assertEqual(runner.summary["stat"]["skipped"], 4)

    def test_html_report(self):
        kwargs = {}
        output_folder_name = os.path.basename(os.path.splitext(self.testset_path)[0])
        runner = HttpRunner().run(self.testset_path)
        summary = runner.summary
        self.assertEqual(summary["stat"]["testsRun"], 10)
        self.assertEqual(summary["stat"]["skipped"], 4)

        runner.gen_html_report(html_report_name=output_folder_name)
        report_save_dir = os.path.join(loader.project_working_directory, 'reports', output_folder_name)
        self.assertGreater(len(os.listdir(report_save_dir)), 0)
        shutil.rmtree(report_save_dir)

    def test_run_testcases(self):
        testcases = [self.testcase]
        runner = HttpRunner().run(testcases)
        summary = runner.summary
        self.assertTrue(summary["success"])
        self.assertEqual(summary["stat"]["testsRun"], 2)
        self.assertIn("details", summary)
        self.assertIn("records", summary["details"][0])

    def test_run_testcase(self):
        testcases = self.testcase
        runner = HttpRunner().run(testcases)
        summary = runner.summary
        self.assertTrue(summary["success"])
        self.assertEqual(summary["stat"]["testsRun"], 2)
        self.assertIn("records", summary["details"][0])

    def test_run_yaml_upload(self):
        testset_path = "tests/httpbin/upload.yml"
        runner = HttpRunner().run(testset_path)
        summary = runner.summary
        self.assertTrue(summary["success"])
        self.assertEqual(summary["stat"]["testsRun"], 1)
        self.assertIn("details", summary)
        self.assertIn("records", summary["details"][0])

    def test_run_post_data(self):
        testcases = [
            {
                "name": "post data",
                "teststeps": [
                    {
                        "name": "post data",
                        "request": {
                            "url": "{}/post".format(HTTPBIN_SERVER),
                            "method": "POST",
                            "headers": {
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
        runner = HttpRunner().run(testcases)
        summary = runner.summary
        self.assertTrue(summary["success"])
        self.assertEqual(summary["stat"]["testsRun"], 1)
        self.assertEqual(summary["details"][0]["records"][0]["meta_data"]["response"]["json"]["data"], "abc")

    def test_html_report_repsonse_image(self):
        testset_path = "tests/httpbin/load_image.yml"
        runner = HttpRunner().run(testset_path)
        summary = runner.summary
        output_folder_name = os.path.basename(os.path.splitext(testset_path)[0])
        report = runner.gen_html_report(html_report_name=output_folder_name)
        self.assertTrue(os.path.isfile(report))
        report_save_dir = os.path.join(loader.project_working_directory, 'reports', output_folder_name)
        shutil.rmtree(report_save_dir)

    def test_testcase_layer(self):
        testcase_path = "tests/testcases/smoketest.yml"
        runner = HttpRunner(failfast=True).run(testcase_path)
        summary = runner.summary
        self.assertTrue(summary["success"])
        self.assertEqual(summary["stat"]["testsRun"], 8)

    def test_run_httprunner_with_hooks(self):
        testcase_file_path = os.path.join(
            os.getcwd(), 'tests/httpbin/hooks.yml')

        start_time = time.time()
        runner = HttpRunner().run(testcase_file_path)
        end_time = time.time()
        summary = runner.summary
        self.assertTrue(summary["success"])
        self.assertLess(end_time - start_time, 20)

    def test_run_httprunner_with_teardown_hooks_alter_response(self):
        testcases = [
            {
                "config": {
                    "name": "test teardown hooks",
                    "path": "tests/httpbin/hooks.yml"
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
        runner = HttpRunner().run(testcases)
        summary = runner.summary
        self.assertTrue(summary["success"])

    def test_run_httprunner_with_teardown_hooks_not_exist_attribute(self):
        testcases = [
            {
                "name": "test teardown hooks",
                "config": {
                    "path": "tests/httpbin/hooks.yml"
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
        runner = HttpRunner().run(testcases)
        summary = runner.summary
        self.assertFalse(summary["success"])
        self.assertEqual(summary["stat"]["errors"], 1)

    def test_run_httprunner_with_teardown_hooks_error(self):
        testcases = [
            {
                "name": "test teardown hooks",
                "config": {},
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
        runner = HttpRunner().run(testcases)
        summary = runner.summary
        self.assertFalse(summary["success"])
        self.assertEqual(summary["stat"]["errors"], 1)

    def test_run_testcase_hardcode(self):
        for testcase_file_path in self.testcase_file_path_list:
            runner = HttpRunner().run(testcase_file_path)
            self.assertTrue(runner.summary["success"])

    def test_run_testcases_hardcode(self):
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
        self.assertEqual(len(summary["details"]), 1)

    def test_run_testcase_output(self):
        testcase_file_path = os.path.join(
            os.getcwd(), 'tests/data/demo_testset_layer.yml')
        runner = HttpRunner(failfast=True).run(testcase_file_path)
        summary = runner.summary
        self.assertTrue(summary["success"])
        self.assertIn("token", summary["details"][0]["in_out"]["out"])
        self.assertIn("user_agent", summary["details"][0]["in_out"]["in"])

    def test_run_testcase_with_variables_mapping(self):
        testcase_file_path = os.path.join(
            os.getcwd(), 'tests/data/demo_testset_layer.yml')
        variables_mapping = {
            "app_version": '2.9.7'
        }
        runner = HttpRunner(failfast=True).run(testcase_file_path, mapping=variables_mapping)
        summary = runner.summary
        self.assertTrue(summary["success"])
        self.assertIn("token", summary["details"][0]["in_out"]["out"])
        self.assertGreater(len(summary["details"][0]["in_out"]["in"]), 7)

    def test_run_testcase_with_parameters(self):
        testcase_file_path = os.path.join(
            os.getcwd(), 'tests/data/demo_parameters.yml')
        runner = HttpRunner().run(testcase_file_path)
        summary = runner.summary
        self.assertEqual(
            summary["details"][0]["in_out"]["in"]["user_agent"],
            "iOS/10.1"
        )
        self.assertEqual(
            summary["details"][2]["in_out"]["in"]["user_agent"],
            "iOS/10.2"
        )
        self.assertEqual(
            summary["details"][4]["in_out"]["in"]["user_agent"],
            "iOS/10.3"
        )
        self.assertTrue(summary["success"])
        self.assertEqual(len(summary["details"]), 3 * 2)
        self.assertEqual(summary["stat"]["testsRun"], 3 * 2)
        self.assertIn("in", summary["details"][0]["in_out"])
        self.assertIn("out", summary["details"][0]["in_out"])

    def test_run_testcase_with_parameters_name(self):
        testcase_file_path = os.path.join(
            os.getcwd(), 'tests/data/demo_parameters.yml')
        runner = HttpRunner()
        testcases = runner.load_tests(testcase_file_path)
        parsed_testcases = runner.parse_tests(testcases)
        unittest_runner, test_suite = runner.initialize(parsed_testcases)

        self.assertEqual(
            test_suite._tests[0].teststeps[0]['name'],
            'get token with iOS/10.1 and test1'
        )
        self.assertEqual(
            test_suite._tests[1].teststeps[0]['name'],
            'get token with iOS/10.1 and test2'
        )
        self.assertEqual(
            test_suite._tests[2].teststeps[0]['name'],
            'get token with iOS/10.2 and test1'
        )
        self.assertEqual(
            test_suite._tests[3].teststeps[0]['name'],
            'get token with iOS/10.2 and test2'
        )
        self.assertEqual(
            test_suite._tests[4].teststeps[0]['name'],
            'get token with iOS/10.3 and test1'
        )
        self.assertEqual(
            test_suite._tests[5].teststeps[0]['name'],
            'get token with iOS/10.3 and test2'
        )

    def test_load_tests(self):
        testcase_file_path = os.path.join(
            os.getcwd(), 'tests/data/demo_testcase.yml')
        runner = HttpRunner()
        testcases = runner.load_tests(testcase_file_path)
        self.assertIsInstance(testcases, list)
        self.assertEqual(
            testcases[0]["config"]["request"],
            '$demo_default_request'
        )
        self.assertEqual(testcases[0]["config"]["name"], '123$var_a')
        self.assertIn(
            "sum_two",
            runner.project_mapping["debugtalk"]["functions"]
        )

    def test_parse_tests(self):
        testcase_file_path = os.path.join(
            os.getcwd(), 'tests/data/demo_testcase.yml')
        runner = HttpRunner()
        testcases = runner.load_tests(testcase_file_path)
        parsed_testcases = runner.parse_tests(testcases)
        self.assertEqual(parsed_testcases[0]["config"]["variables"]["var_c"], 3)
        self.assertEqual(len(parsed_testcases), 2 * 2)
        self.assertEqual(
            parsed_testcases[0]["config"]["request"]["base_url"],
            '$BASE_URL'
        )
        self.assertEqual(
            parsed_testcases[0]["config"]["variables"]["BASE_URL"],
            'http://127.0.0.1:5000'
        )
        self.assertIsInstance(parsed_testcases, list)
        self.assertEqual(parsed_testcases[0]["config"]["name"], '12311')

    def test_validate_response_content(self):
        testcase_file_path = os.path.join(
            os.getcwd(), 'tests/httpbin/basic.yml')
        runner = HttpRunner().run(testcase_file_path)
        self.assertTrue(runner.summary["success"])


class TestLocustRunner(ApiServerUnittest):

    def setUp(self):
        WebPageUser = type('WebPageUser', (HttpLocust,), {})
        self.locust_client = WebPageUser.client

    def test_LocustRunner(self):
        testcase_file = os.path.join(os.getcwd(), 'tests', 'httpbin', 'basic.yml')
        locust_runner = LocustRunner(self.locust_client)
        locust_runner.run(testcase_file)
