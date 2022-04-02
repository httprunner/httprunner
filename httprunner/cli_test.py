import io
import os
import sys
import unittest

import pytest

from httprunner import loader
from httprunner.cli import main, main_run


class TestCli(unittest.TestCase):
    def setUp(self):
        self.captured_output = io.StringIO()
        sys.stdout = self.captured_output

    def tearDown(self):
        sys.stdout = sys.__stdout__  # Reset redirect.

    def test_show_version(self):
        sys.argv = ["hrun", "-V"]

        with self.assertRaises(SystemExit) as cm:
            main()

        self.assertEqual(cm.exception.code, 0)

        from httprunner import __version__

        self.assertIn(__version__, self.captured_output.getvalue().strip())

    def test_show_help(self):
        sys.argv = ["hrun", "-h"]

        with self.assertRaises(SystemExit) as cm:
            main()

        self.assertEqual(cm.exception.code, 0)

        from httprunner import __description__

        self.assertIn(__description__, self.captured_output.getvalue().strip())

    def test_debug_pytest(self):
        cwd = os.getcwd()
        try:
            os.chdir(os.path.join(cwd, "examples", "postman_echo"))
            exit_code = pytest.main(
                ["-s", "request_methods/request_with_testcase_reference_test.py"]
            )
            self.assertEqual(exit_code, 0)
        finally:
            os.chdir(cwd)

    def test_run_testcase_with_abnormal_path(self):
        loader.project_meta = None
        exit_code = main_run(["examples/data/a-b.c/2 3.yml"])
        self.assertEqual(exit_code, 0)
        self.assertTrue(os.path.exists("examples/data/a_b_c/__init__.py"))
        self.assertTrue(os.path.exists("examples/data/debugtalk.py"))
        self.assertTrue(os.path.exists("examples/data/a_b_c/T1_test.py"))
        self.assertTrue(os.path.exists("examples/data/a_b_c/T2_3_test.py"))
