import os
import shutil

from httprunner import HttpRunner
from tests.base import ApiServerUnittest


class TestHttpRunner(ApiServerUnittest):

    def setUp(self):
        self.testset_path = "tests/data/demo_testset_cli.yml"
        self.reset_all()

    def reset_all(self):
        url = "%s/api/reset-all" % self.host
        headers = self.get_authenticated_headers()
        return self.api_client.get(url, headers=headers)

    def test_text_run_times(self):
        runner = HttpRunner().run(self.testset_path)
        self.assertEqual(runner.summary["stat"]["testsRun"], 10)

    def test_text_skip(self):
        runner = HttpRunner().run(self.testset_path)
        self.assertEqual(runner.summary["stat"]["skipped"], 4)

    def test_html_report(self):
        kwargs = {}
        output_folder_name = os.path.basename(os.path.splitext(self.testset_path)[0])
        runner = HttpRunner().run(self.testset_path)
        summary = runner.summary
        self.assertEqual(summary["stat"]["testsRun"], 10)
        self.assertEqual(summary["stat"]["skipped"], 4)

        runner.gen_html_report(html_report_name=output_folder_name)
        report_save_dir = os.path.join(os.getcwd(), 'reports', output_folder_name)
        shutil.rmtree(report_save_dir)
