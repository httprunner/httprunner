import os
import shutil

from httprunner import HttpRunner
from tests.api_server import HTTPBIN_SERVER
from tests.base import ApiServerUnittest


class TestHttpRunner(ApiServerUnittest):

    def setUp(self):
        self.testset_path = "tests/data/demo_testset_cli.yml"
        self.testset = {
            'name': 'testset description',
            'config': {
                'path': 'docs/data/demo-quickstart-2.yml',
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
        report_save_dir = os.path.join(os.getcwd(), 'reports', output_folder_name)
        shutil.rmtree(report_save_dir)

    def test_run_testsets(self):
        testsets = [self.testset]
        runner = HttpRunner().run(testsets)
        summary = runner.summary
        self.assertTrue(summary["success"])
        self.assertEqual(summary["stat"]["testsRun"], 2)
        self.assertIn("details", summary)
        self.assertIn("records", summary["details"][0])

    def test_run_testset(self):
        testsets = self.testset
        runner = HttpRunner().run(testsets)
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
        testsets = [
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
        runner = HttpRunner().run(testsets)
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
        report_save_dir = os.path.join(os.getcwd(), 'reports', output_folder_name)
        shutil.rmtree(report_save_dir)

    def test_testcase_layer(self):
        testcase_path = "tests/testcases/smoketest.yml"
        runner = HttpRunner(failfast=True).run(testcase_path)
        summary = runner.summary
        self.assertTrue(summary["success"])
        self.assertEqual(summary["stat"]["testsRun"], 8)
