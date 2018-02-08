import os
import shutil
import sys

from httprunner.task import TaskSuite
from pyunitreport import HTMLTestRunner
from tests.base import ApiServerUnittest


class TestCli(ApiServerUnittest):

    def test_run_times(self):
        testset_path = "tests/data/demo_testset_cli.yml"
        output_folder_name = os.path.basename(os.path.splitext(testset_path)[0])
        kwargs = {
            "output": output_folder_name
        }

        task_suite = TaskSuite(testset_path)
        result = HTMLTestRunner(**kwargs).run(task_suite)
        self.assertEqual(result.testsRun, 5)

        report_save_dir = os.path.join(os.getcwd(), 'reports', output_folder_name)
        shutil.rmtree(report_save_dir)
