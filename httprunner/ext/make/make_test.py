import unittest
from httprunner.ext.make import make_testcase, main_make


class TestLoader(unittest.TestCase):
    def test_make_testcase(self):
        path = "examples/postman_echo/request_methods/request_with_variables.yml"
        testcase_python_path = make_testcase(path)
        self.assertEqual(
            testcase_python_path,
            "examples/postman_echo/request_methods/request_with_variables_test.py",
        )

    def test_make_testcase_folder(self):
        path = ["examples/postman_echo/request_methods/"]
        testcase_python_list = main_make(path)
        self.assertIn(
            "examples/postman_echo/request_methods/request_with_functions_test.py",
            testcase_python_list,
        )
