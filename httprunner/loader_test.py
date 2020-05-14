import unittest
from httprunner.new_loader import load_testcase_file


class TestLoader(unittest.TestCase):

    def test_load_testcase_file(self):
        path = "examples/postman_echo/request_methods/request_with_variables.yml"
        testcase = load_testcase_file(path)
        self.assertEqual(testcase.config.name, "request methods testcase with variables")
        self.assertEqual(len(testcase.teststeps), 3)
