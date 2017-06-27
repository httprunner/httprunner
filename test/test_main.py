import os
import random
import requests
from test.base import ApiServerUnittest
from ate import main, utils

class TestMain(ApiServerUnittest):

    def setUp(self):
        self.clear_users()

    def clear_users(self):
        url = "http://127.0.0.1:5000/api/users"
        return requests.delete(url)

    def test_create_suite(self):
        testcase_file_path = os.path.join(os.getcwd(), 'test/data/simple_demo_no_auth.yml')
        testsets = utils.load_testcases_by_path(testcase_file_path)
        suite = main.create_suite(testsets[0])
        self.assertEqual(suite.countTestCases(), 2)
        for testcase in suite:
            self.assertIsInstance(testcase, main.ApiTestCase)

    def test_create_task(self):
        testcase_file_path = os.path.join(os.getcwd(), 'test/data/simple_demo_no_auth.yml')
        task_suite = main.create_task(testcase_file_path)
        self.assertEqual(task_suite.countTestCases(), 2)
        for suite in task_suite:
            for testcase in suite:
                self.assertIsInstance(testcase, main.ApiTestCase)
