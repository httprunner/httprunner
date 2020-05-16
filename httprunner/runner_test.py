import unittest

from httprunner.runner import HttpRunner


class TestHttpRunner(unittest.TestCase):
    def setUp(self):
        self.runner = HttpRunner()

    def test_run_testcase_by_path_request_only(self):
        self.runner.run_path(
            "examples/postman_echo/request_methods/request_with_variables.yml"
        )
        result = self.runner.get_summary()
        self.assertTrue(result.success)
        self.assertEqual(result.name, "request methods testcase with variables")
        self.assertEqual(result.step_datas[0].name, "get with params")
        self.assertEqual(len(result.step_datas), 3)

    def test_run_testcase_by_path_ref_testcase(self):
        self.runner.run_path(
            "examples/postman_echo/request_methods/request_with_testcase_reference.yml"
        )
        result = self.runner.get_summary()
        self.assertTrue(result.success)
        self.assertEqual(result.name, "request methods testcase: reference testcase")
        self.assertEqual(result.step_datas[0].name, "request with variables")
        self.assertEqual(len(result.step_datas), 1)
