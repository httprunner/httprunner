import os
import shutil
import subprocess
import unittest

from httprunner.ext.scaffold import create_scaffold


class TestUtils(unittest.TestCase):
    def test_create_scaffold(self):
        project_name = "projectABC"
        create_scaffold(project_name)
        self.assertTrue(os.path.isdir(os.path.join(project_name, "har")))
        self.assertTrue(os.path.isdir(os.path.join(project_name, "testcases")))
        self.assertTrue(os.path.isdir(os.path.join(project_name, "reports")))
        self.assertTrue(os.path.isfile(os.path.join(project_name, "debugtalk.py")))
        self.assertTrue(os.path.isfile(os.path.join(project_name, ".env")))

        # run demo testcases
        try:
            subprocess.check_call(["hrun", project_name])
        except subprocess.SubprocessError:
            raise
        finally:
            shutil.rmtree(project_name)
