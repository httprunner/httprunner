import unittest
from httprunner.new_loader import load_testcase_file


class TestLoader(unittest.TestCase):

    def test_load_testcase_file(self):
        path = "examples/postman_echo/request_methods/request_with_variables.yml"
        testcase_json, testcase_obj = load_testcase_file(path)
        self.assertEqual(testcase_json["config"]["name"], "request methods testcase with variables")
        self.assertEqual(testcase_obj.config.name, "request methods testcase with variables")
        self.assertEqual(len(testcase_json["teststeps"]), 3)
        self.assertEqual(len(testcase_obj.teststeps), 3)
