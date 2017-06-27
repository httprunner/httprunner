import os
import requests
from ate import runner, exception, utils
from test.base import ApiServerUnittest

class TestRunner(ApiServerUnittest):

    def setUp(self):
        self.test_runner = runner.TestRunner()
        self.clear_users()

    def clear_users(self):
        url = "http://127.0.0.1:5000/api/users"
        return requests.delete(url)

    def test_run_single_testcase_yaml_success(self):
        testcase_file_path = os.path.join(os.getcwd(), 'test/data/simple_demo_no_auth.yml')
        testcases = utils.load_testcases(testcase_file_path)
        testcase = testcases[0]["test"]
        success, _ = self.test_runner.run_test(testcase)
        self.assertTrue(success)

    def test_run_single_testcase_json_success(self):
        testcase_file_path = os.path.join(os.getcwd(), 'test/data/simple_demo_no_auth.json')
        testcases = utils.load_testcases(testcase_file_path)
        testcase = testcases[0]["test"]
        success, _ = self.test_runner.run_test(testcase)
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
        success, diff_content = self.test_runner.run_test(testcase)
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

    def test_run_testset_json_success(self):
        testcase_file_path = os.path.join(os.getcwd(), 'test/data/simple_demo_no_auth.json')
        testsets = utils.load_testcases_by_path(testcase_file_path)
        results = self.test_runner.run_testset(testsets[0])
        self.assertEqual(len(results), 2)
        self.assertEqual(results, [(True, {}), (True, {})])

    def test_run_testsets_json_success(self):
        testcase_file_path = os.path.join(os.getcwd(), 'test/data/simple_demo_no_auth.json')
        testsets = utils.load_testcases_by_path(testcase_file_path)
        results = self.test_runner.run_testsets(testsets)
        self.assertEqual(len(results), 1)
        self.assertEqual(results[0], [(True, {}), (True, {})])

    def test_run_testset_yaml_success(self):
        testcase_file_path = os.path.join(os.getcwd(), 'test/data/simple_demo_no_auth.yml')
        testsets = utils.load_testcases_by_path(testcase_file_path)
        results = self.test_runner.run_testset(testsets[0])
        self.assertEqual(len(results), 2)
        self.assertEqual(results, [(True, {}), (True, {})])

    def test_run_testsets_yaml_success(self):
        testcase_file_path = os.path.join(os.getcwd(), 'test/data/simple_demo_no_auth.yml')
        testsets = utils.load_testcases_by_path(testcase_file_path)
        results = self.test_runner.run_testsets(testsets)
        self.assertEqual(len(results), 1)
        self.assertEqual(results[0], [(True, {}), (True, {})])
