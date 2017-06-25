import os
import random
import requests
from ate import runner, exception, utils
from .base import ApiServerUnittest

class TestRunnerV2(ApiServerUnittest):

    authentication = True

    def setUp(self):
        self.test_runner = runner.TestRunner()
        self.clear_users()

    def clear_users(self):
        url = "http://127.0.0.1:5000/api/users"
        return requests.delete(url, headers=self.prepare_headers())

    def test_run_single_testcase_yaml(self):
        testcase_file_path = os.path.join(
            os.getcwd(), 'test/data/simple_demo_auth_hardcode.yml')
        testcases = utils.load_testcases(testcase_file_path)
        success, _ = self.test_runner.run_single_testcase(testcases[0])
        self.assertTrue(success)

    def test_run_single_testcase_json(self):
        testcase_file_path = os.path.join(
            os.getcwd(), 'test/data/simple_demo_auth_hardcode.json')
        testcases = utils.load_testcases(testcase_file_path)
        success, _ = self.test_runner.run_single_testcase(testcases[0])
        self.assertTrue(success)

    def test_run_testcase_auth_suite_yaml(self):
        testcase_file_path = os.path.join(
            os.getcwd(), 'test/data/simple_demo_auth_hardcode.yml')
        testcases = utils.load_testcases(testcase_file_path)
        result = self.test_runner.run_testcase_suite(testcases)
        self.assertEqual(len(result), 2)
        self.assertEqual(result, [(True, {}), (True, {})])

    def test_run_testcase_auth_suite_json(self):
        testcase_file_path = os.path.join(
            os.getcwd(), 'test/data/simple_demo_auth_hardcode.json')
        testcases = utils.load_testcases(testcase_file_path)
        result = self.test_runner.run_testcase_suite(testcases)
        self.assertEqual(len(result), 2)
        self.assertEqual(result, [(True, {}), (True, {})])
