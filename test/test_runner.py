import os
import random
import requests
from ate import runner, exception, utils
from .base import ApiServerUnittest

class TestUtils(ApiServerUnittest):

    def setUp(self):
        self.test_runner = runner.TestRunner()
        self.clear_users()

    def clear_users(self):
        url = "http://127.0.0.1:5000/api/users"
        return requests.delete(url)

    def test_run_single_testcase_success(self):
        testcase_file_path = os.path.join(os.getcwd(), 'test/data/demo.json')
        testcases = utils.load_testcases(testcase_file_path)
        success, _ = self.test_runner.run_single_testcase(testcases[0])
        self.assertTrue(success)

    def test_run_single_testcase_fail(self):
        testcase = {
            "name": "create user which does not exist",
            "request": {
                "url": "http://127.0.0.1:5000/api/users/1000",
                "method": "POST",
                "headers": {
                    "content-type": "application/json"
                },
                "json": {
                    "name": "user1",
                    "password": "123456"
                }
            },
            "response": {
                "status_code": 200,
                "headers": {
                    "Content-Type": "html/text"
                },
                "body": {
                    'success': False,
                    'msg': "user already existed."
                }
            }
        }
        success, diff_content = self.test_runner.run_single_testcase(testcase)
        self.assertFalse(success)
        self.assertEqual(
            diff_content['status_code'],
            {'expected': 200, 'value': 201}
        )
        self.assertEqual(
            diff_content['headers'],
            {'Content-Type': {'expected': 'html/text', 'value': 'application/json'}}
        )
        self.assertEqual(
            diff_content['body'],
            {
                'msg': {
                    'expected': 'user already existed.',
                    'value': 'user created successfully.'
                },
                'success': {
                    'expected': False,
                    'value': True
                }
            }
        )

    def test_run_testcase_suite_json_success(self):
        testcase_file_path = os.path.join(os.getcwd(), 'test/data/demo.json')
        testcases = utils.load_testcases(testcase_file_path)
        result = self.test_runner.run_testcase_suite(testcases)
        self.assertEqual(len(result), 2)
        self.assertEqual(result, [(True, {}), (True, {})])

    def test_run_testcase_suite_yaml_success(self):
        testcase_file_path = os.path.join(os.getcwd(), 'test/data/demo.yml')
        testcases = utils.load_testcases(testcase_file_path)
        result = self.test_runner.run_testcase_suite(testcases)
        self.assertEqual(len(result), 2)
        self.assertEqual(result, [(True, {}), (True, {})])
