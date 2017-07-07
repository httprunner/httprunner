import os
import requests
from ate import runner, exception, utils
from test.base import ApiServerUnittest

class TestRunner(ApiServerUnittest):

    def setUp(self):
        self.test_runner = runner.Runner()
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
            "extract_binds": {
                "resp_status_code": "status_code",
                "resp_body_success": "content.success",
                "resp_headers_contenttype": "headers.content-type"
            },
            "validators": [
                {"check": "resp_status_code", "comparator": "eq", "expected": 200},
                {"check": "resp_body_success", "comparator": "eq", "expected": False},
                {"check": "resp_headers_contenttype", "comparator": "eq", "expected": "html/text"}
            ]
        }

        success, diff_content_list = self.test_runner.run_test(testcase)
        self.assertFalse(success)
        self.assertEqual(
            diff_content_list[0],
            {"check": "resp_status_code", "comparator": "eq", "expected": 200, 'value': 201}
        )
        self.assertEqual(
            diff_content_list[1],
            {"check": "resp_body_success", "comparator": "eq", "expected": False, 'value': True}
        )
        self.assertEqual(
            diff_content_list[2],
            {"check": "resp_headers_contenttype", "comparator": "eq", "expected": "html/text", 'value': "application/json"}
        )

    def test_run_testset_json_success(self):
        testcase_file_path = os.path.join(os.getcwd(), 'test/data/simple_demo_no_auth.json')
        testsets = utils.load_testcases_by_path(testcase_file_path)
        results = self.test_runner.run_testset(testsets[0])
        self.assertEqual(len(results), 2)
        self.assertEqual(results, [(True, []), (True, [])])

    def test_run_testsets_json_success(self):
        testcase_file_path = os.path.join(os.getcwd(), 'test/data/simple_demo_no_auth.json')
        testsets = utils.load_testcases_by_path(testcase_file_path)
        results = self.test_runner.run_testsets(testsets)
        self.assertEqual(len(results), 1)
        self.assertEqual(results[0], [(True, []), (True, [])])

    def test_run_testset_yaml_success(self):
        testcase_file_path = os.path.join(os.getcwd(), 'test/data/simple_demo_no_auth.yml')
        testsets = utils.load_testcases_by_path(testcase_file_path)
        results = self.test_runner.run_testset(testsets[0])
        self.assertEqual(len(results), 2)
        self.assertEqual(results, [(True, []), (True, [])])

    def test_run_testsets_yaml_success(self):
        testcase_file_path = os.path.join(os.getcwd(), 'test/data/simple_demo_no_auth.yml')
        testsets = utils.load_testcases_by_path(testcase_file_path)
        results = self.test_runner.run_testsets(testsets)
        self.assertEqual(len(results), 1)
        self.assertEqual(results[0], [(True, []), (True, [])])
