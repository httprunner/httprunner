import os
import unittest

from httprunner.ext.locusts.utils import prepare_locust_tests


class TestLocust(unittest.TestCase):

    def test_prepare_locust_tests(self):
        path = os.path.join(
            os.getcwd(), 'tests/locust_tests/demo_locusts.yml')
        locust_tests = prepare_locust_tests(path)
        self.assertEqual(len(locust_tests), 2 + 3)
        name_list = [
            "create user 1000 and check result.",
            "create user 1001 and check result."
        ]
        self.assertIn(locust_tests[0]["config"]["name"], name_list)
        self.assertIn(locust_tests[4]["config"]["name"], name_list)
