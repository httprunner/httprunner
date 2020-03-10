import unittest

from httprunner.loader import check


class TestLoaderCheck(unittest.TestCase):

    def test_is_testcases(self):
        data_structure = "path/to/file"
        self.assertFalse(check.is_test_content(data_structure))
        data_structure = ["path/to/file1", "path/to/file2"]
        self.assertFalse(check.is_test_content(data_structure))

        data_structure = {
            "project_mapping": {
                "PWD": "XXXXX",
                "functions": {},
                "env": {}
            },
            "testcases": [
                {  # testcase data structure
                    "config": {
                        "name": "desc1",
                        "path": "testcase1_path",
                        "variables": [],  # optional
                    },
                    "teststeps": [
                        # test data structure
                        {
                            'name': 'test step desc1',
                            'variables': [],  # optional
                            'extract': {},  # optional
                            'validate': [],
                            'request': {
                                "method": "GET",
                                "url": "https://docs.httprunner.org"
                            }
                        },
                        # test_dict2   # another test dict
                    ]
                },
                # testcase_dict_2     # another testcase dict
            ]
        }
        self.assertTrue(check.is_test_content(data_structure))
