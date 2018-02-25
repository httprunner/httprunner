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
        kwargs = {
            "gen_html_report": False
        }
        result = HttpRunner(self.testset_path, **kwargs).run()
        self.assertEqual(result["stat"]["testsRun"], 10)

    def test_text_skip(self):
        kwargs = {
            "gen_html_report": False
        }
        result = HttpRunner(self.testset_path, **kwargs).run()
        self.assertEqual(result["stat"]["skipped"], 4)

    def test_html_report(self):
        kwargs = {
            "gen_html_report": True
        }
        output_folder_name = os.path.basename(os.path.splitext(self.testset_path)[0])
        run_kwargs = {
            "html_report_name": output_folder_name
        }
        result = HttpRunner(self.testset_path).run(**run_kwargs)
        self.assertEqual(result["stat"]["testsRun"], 10)
        self.assertEqual(result["stat"]["skipped"], 4)

        report_save_dir = os.path.join(os.getcwd(), 'reports', output_folder_name)
        shutil.rmtree(report_save_dir)
