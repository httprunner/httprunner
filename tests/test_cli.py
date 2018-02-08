import os
import shutil
import sys

from httprunner.task import TaskSuite
from pyunitreport import HTMLTestRunner
from tests.base import ApiServerUnittest


class TestCli(ApiServerUnittest):

    def setUp(self):
        testset_path = "tests/data/demo_testset_cli.yml"
        output_folder_name = os.path.basename(os.path.splitext(testset_path)[0])
        self.kwargs = {
            "output": output_folder_name
        }
        self.task_suite = TaskSuite(testset_path)
        self.report_save_dir = os.path.join(os.getcwd(), 'reports', output_folder_name)
        self.reset_all()

    def reset_all(self):
        url = "%s/api/reset-all" % self.host
        headers = self.get_authenticated_headers()
        return self.api_client.get(url, headers=headers)

    def test_run_times(self):
        result = HTMLTestRunner(**self.kwargs).run(self.task_suite)
        self.assertEqual(result.testsRun, 8)
        shutil.rmtree(self.report_save_dir)

    def test_skip(self):
        result = HTMLTestRunner(**self.kwargs).run(self.task_suite)
        self.assertEqual(len(result.skipped), 4)
        shutil.rmtree(self.report_save_dir)
