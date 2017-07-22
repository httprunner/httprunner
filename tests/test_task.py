import os
from tests.base import ApiServerUnittest
from ate import task, utils

class TestTask(ApiServerUnittest):

    def setUp(self):
        self.reset_all()

    def reset_all(self):
        url = "%s/api/reset-all" % self.host
        headers = self.get_authenticated_headers()
        return self.api_client.get(url, headers=headers)

    def test_create_suite(self):
        testcase_file_path = os.path.join(os.getcwd(), 'tests/data/demo_testset_variables.yml')
        testsets = utils.load_testcases_by_path(testcase_file_path)
        suite = task.create_suite(testsets[0])
        self.assertEqual(suite.countTestCases(), 3)
        for testcase in suite:
            self.assertIsInstance(testcase, task.ApiTestCase)

    def test_create_task(self):
        testcase_file_path = os.path.join(os.getcwd(), 'tests/data/demo_testset_variables.yml')
        task_suite = task.create_task(testcase_file_path)
        self.assertEqual(task_suite.countTestCases(), 3)
        for suite in task_suite:
            for testcase in suite:
                self.assertIsInstance(testcase, task.ApiTestCase)
