import unittest

from examples.postman_echo.request_methods.request_with_functions_test import (
    TestCaseRequestWithFunctions,
)


class TestRunRequest(unittest.TestCase):
    def test_run_request(self):
        runner = TestCaseRequestWithFunctions().test_start()
        summary = runner.get_summary()
        self.assertTrue(summary.success)
        self.assertEqual(summary.name, "request methods testcase with functions")
        self.assertEqual(len(summary.step_results), 3)
        self.assertEqual(summary.step_results[0].name, "get with params")
        self.assertEqual(summary.step_results[1].name, "post raw text")
        self.assertEqual(summary.step_results[2].name, "post form data")
