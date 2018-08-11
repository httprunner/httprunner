import os

from httprunner import loader, task
from tests.base import ApiServerUnittest


class TestTask(ApiServerUnittest):

    def setUp(self):
        self.reset_all()

    def reset_all(self):
        url = "%s/api/reset-all" % self.host
        headers = self.get_authenticated_headers()
        return self.api_client.get(url, headers=headers)

    def test_create_suite(self):
        testcase_file_path = os.path.join(os.getcwd(), 'tests/data/demo_testset_variables.yml')
        testset = loader._load_test_file(testcase_file_path)
        suite = task.TestSuite(testset)
        self.assertEqual(suite.countTestCases(), 3)
        for testcase in suite:
            self.assertIsInstance(testcase, task.TestCase)

    def test_create_task(self):
        testsets = [
            {
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
        ]
        test_suite_list = task.init_test_suites(testsets)
        self.assertEqual(len(test_suite_list), 1)
        task_suite = test_suite_list[0]
        self.assertEqual(task_suite.countTestCases(), 2)
        for testcase in task_suite:
            self.assertIsInstance(testcase, task.TestCase)
