import unittest

from httprunner.api import HttpRunner


class TestHttpRunner(unittest.TestCase):

    def setUp(self):
        self.runner = HttpRunner()

    def test_run_testcase_by_path(self):
        summary = self.runner.run_path("examples/postman_echo/request_methods/")
        self.assertTrue(summary.success)
        self.assertEqual(summary.testcases[0].name, "request methods testcase with variables")
        self.assertGreater(summary.stat.total, 1)
