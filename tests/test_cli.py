import sys
import unittest

from httprunner.cli import main
from httprunner.compat import io


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
