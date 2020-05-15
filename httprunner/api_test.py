import unittest

from httprunner.api import HttpRunner


class TestHttpRunner(unittest.TestCase):
    def setUp(self):
        self.runner = HttpRunner()

    def test_run_testcase_by_path_request_only(self):
        summary = self.runner.run_path(
            "examples/postman_echo/request_methods/request_with_variables.yml"
        )
        self.assertTrue(summary.success)
        self.assertEqual(
            summary.testcases[0].name, "request methods testcase with variables"
        )
        self.assertGreater(summary.stat.total, 1)

    def test_run_testcase_by_path_ref_testcase(self):
        summary = self.runner.run_path(
            "examples/postman_echo/request_methods/request_with_testcase_reference.yml"
        )
        self.assertTrue(summary.success)
        self.assertEqual(
            summary.testcases[0].name, "request methods testcase with variables"
        )
        self.assertGreater(summary.stat.total, 1)