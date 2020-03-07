
import os
import unittest

from httprunner import exceptions
from httprunner.loader import locate


class TestLoaderLocate(unittest.TestCase):

    def test_locate_file(self):
        with self.assertRaises(exceptions.FileNotFound):
            locate.locate_file(os.getcwd(), "debugtalk.py")

        with self.assertRaises(exceptions.FileNotFound):
            locate.locate_file("", "debugtalk.py")

        start_path = os.path.join(os.getcwd(), "tests")
        self.assertEqual(
            locate.locate_file(start_path, "debugtalk.py"),
            os.path.join(
                os.getcwd(), "tests/debugtalk.py"
            )
        )
        self.assertEqual(
            locate.locate_file("tests/", "debugtalk.py"),
            os.path.join(os.getcwd(), "tests", "debugtalk.py")
        )
        self.assertEqual(
            locate.locate_file("tests", "debugtalk.py"),
            os.path.join(os.getcwd(), "tests", "debugtalk.py")
        )
        self.assertEqual(
            locate.locate_file("tests/base.py", "debugtalk.py"),
            os.path.join(os.getcwd(), "tests", "debugtalk.py")
        )
        self.assertEqual(
            locate.locate_file("tests/data/demo_testcase.yml", "debugtalk.py"),
            os.path.join(os.getcwd(), "tests", "debugtalk.py")
        )
