import unittest

from httprunner.v3.api import HttpRunner


class TestHttpRunner(unittest.TestCase):

    def setUp(self):
        self.runner = HttpRunner(failfast=True)

    def test_run_testcase_by_path(self):
        summary = self.runner.run_path("examples/postman_echo/request_methods/request_with_variables.yml")
        self.assertTrue(summary["success"])
        self.assertEqual(summary["details"][0]["name"], "request methods testcase with variables")
        self.assertEqual(summary["details"][0]["records"][0]["name"], "request methods testcase with variables")
        self.assertEqual(summary["stat"]["testcases"]["total"], 1)
        # self.assertEqual(summary["stat"]["teststeps"]["total"], 2)
