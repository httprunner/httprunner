import unittest

from httprunner import validator


class TestValidator(unittest.TestCase):

    def test_is_testcases(self):
        data_structure = "path/to/file"
        self.assertFalse(validator.is_testcases(data_structure))
        data_structure = ["path/to/file1", "path/to/file2"]
        self.assertFalse(validator.is_testcases(data_structure))

        data_structure = {
            "name": "desc1",
            "config": {},
            "api": {},
            "testcases": ["testcase11", "testcase12"]
        }
        self.assertTrue(data_structure)
        data_structure = [
            {
                "name": "desc1",
                "config": {},
                "api": {},
                "testcases": ["testcase11", "testcase12"]
            },
            {
                "name": "desc2",
                "config": {},
                "api": {},
                "testcases": ["testcase21", "testcase22"]
            }
        ]
        self.assertTrue(data_structure)
