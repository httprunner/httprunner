import unittest

from httprunner.ext.make import main_make, convert_testcase_path


class TestLoader(unittest.TestCase):
    def test_make_testcase(self):
        path = ["examples/postman_echo/request_methods/request_with_variables.yml"]
        testcase_python_list = main_make(path)
        self.assertEqual(
            testcase_python_list[0],
            "examples/postman_echo/request_methods/request_with_variables_test.py",
        )

    def test_make_testcase_folder(self):
        path = ["examples/postman_echo/request_methods/"]
        testcase_python_list = main_make(path)
        self.assertIn(
            "examples/postman_echo/request_methods/request_with_functions_test.py",
            testcase_python_list,
        )

    def test_convert_testcase_path(self):
        self.assertEqual(
            convert_testcase_path("mubu.login.yml")[0], "mubu_login_test.py"
        )
        self.assertEqual(
            convert_testcase_path("/path/to/mubu.login.yml")[0],
            "/path/to/mubu_login_test.py",
        )
        self.assertEqual(
            convert_testcase_path("/path/to 2/mubu.login.yml")[0],
            "/path/to 2/mubu_login_test.py",
        )
        self.assertEqual(
            convert_testcase_path("/path/to 2/mubu.login.yml")[1], "MubuLogin"
        )
        self.assertEqual(
            convert_testcase_path("mubu login.yml")[0], "mubu_login_test.py"
        )
        self.assertEqual(
            convert_testcase_path("/path/to 2/mubu login.yml")[1], "MubuLogin"
        )
        self.assertEqual(
            convert_testcase_path("/path/to 2/mubu-login.yml")[0],
            "/path/to 2/mubu_login_test.py",
        )
        self.assertEqual(
            convert_testcase_path("/path/to 2/mubu-login.yml")[1], "MubuLogin"
        )
        self.assertEqual(
            convert_testcase_path("/path/to 2/幕布login.yml")[0],
            "/path/to 2/幕布login_test.py",
        )
        self.assertEqual(convert_testcase_path("/path/to/幕布login.yml")[1], "幕布Login")

    def test_make_testsuite(self):
        path = ["examples/postman_echo/request_methods/demo_testsuite.yml"]
        testcase_python_list = main_make(path)
        self.assertEqual(len(testcase_python_list), 2)
        self.assertIn(
            "examples/postman_echo/request_methods/demo_testsuite_yml/request_with_functions_test.py",
            testcase_python_list,
        )
        self.assertIn(
            "examples/postman_echo/request_methods/demo_testsuite_yml/request_with_testcase_reference_test.py",
            testcase_python_list,
        )
