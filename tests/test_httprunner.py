import os
import shutil

from httprunner import HttpRunner
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
            'testcases': [
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
        self.assertIn("records", summary)

    def test_run_testset(self):
        testsets = self.testset
        runner = HttpRunner().run(testsets)
        summary = runner.summary
        self.assertTrue(summary["success"])
        self.assertEqual(summary["stat"]["testsRun"], 2)
        self.assertIn("records", summary)
