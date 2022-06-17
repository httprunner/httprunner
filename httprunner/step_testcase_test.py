import unittest

from httprunner.runner import HttpRunner
from httprunner.step_testcase import RunTestCase
from examples.postman_echo.request_methods.request_with_functions_test import (
    TestCaseRequestWithFunctions,
)


class TestRunTestCase(unittest.TestCase):
    def setUp(self):
        self.runner = TestCaseRequestWithFunctions()
        self.runner.test_start()

    def test_run_testcase_by_path(self):

        step_result = (
            RunTestCase("run referenced testcase")
            .call(TestCaseRequestWithFunctions)
            .run(self.runner)
        )
        self.assertTrue(step_result.success)
        self.assertEqual(step_result.name, "run referenced testcase")
        self.assertEqual(len(step_result.data), 3)
        self.assertEqual(step_result.data[0].name, "get with params")
        self.assertEqual(step_result.data[1].name, "post raw text")
        self.assertEqual(step_result.data[2].name, "post form data")
