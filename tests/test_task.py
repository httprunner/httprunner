import os

from httprunner import task
from httprunner.testcase import load_test_file
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
        testset = load_test_file(testcase_file_path)
        suite = task.ApiTestSuite(testset)
        self.assertEqual(suite.countTestCases(), 3)
        for testcase in suite:
            self.assertIsInstance(testcase, task.ApiTestCase)

    def test_create_task(self):
        testcase_file_path = os.path.join(os.getcwd(), 'tests/data/demo_testset_variables.yml')
        task_suite = task.TaskSuite(testcase_file_path)
        self.assertEqual(task_suite.countTestCases(), 3)
        for suite in task_suite:
            for testcase in suite:
                self.assertIsInstance(testcase, task.ApiTestCase)
