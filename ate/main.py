import os
import argparse
import logging
import unittest

import HtmlTestRunner

from ate import runner, utils


class ApiTestCase(unittest.TestCase):
    """ create a testcase.
    """
    def __init__(self, test_runner, testcase):
        super(ApiTestCase, self).__init__()
        self.test_runner = test_runner
        self.testcase = testcase

    def runTest(self):
        """ run testcase and check result.
        """
        result = self.test_runner.run_test(self.testcase)
        self.assertEqual(result, (True, []))

def create_suite(testset):
    """ create test suite with a testset, it may include one or several testcases.
        each suite should initialize a seperate Runner() with testset config.
    """
    suite = unittest.TestSuite()

    test_runner = runner.Runner()
    config_dict = testset.get("config", {})
    test_runner.init_config(config_dict, level="testset")
    testcases = testset.get("testcases", [])

    for testcase in testcases:
        if utils.PYTHON_VERSION == 3:
            ApiTestCase.runTest.__doc__ = testcase['name']
        else:
            ApiTestCase.runTest.__func__.__doc__ = testcase['name']

        test = ApiTestCase(test_runner, testcase)
        suite.addTest(test)

    return suite

def create_task(testcase_path):
    """ create test task suite with specified testcase path.
        each task suite may include one or several test suite.
    """
    task_suite = unittest.TestSuite()
    testsets = utils.load_testcases_by_path(testcase_path)

    for testset in testsets:
        suite = create_suite(testset)
        task_suite.addTest(suite)

    return task_suite

def main():
    """ parse command line options and run commands.
    """
    parser = argparse.ArgumentParser(
        description='Api Test Engine.')
    parser.add_argument(
        '--testcase-path', default='testcases',
        help="testcase file path")
    parser.add_argument(
        '--log-level', default='INFO',
        help="Specify logging level, default is INFO.")

    args = parser.parse_args()

    log_level = getattr(logging, args.log_level.upper())
    logging.basicConfig(level=log_level)

    testcase_path = args.testcase_path.rstrip('/')
    task_suite = create_task(testcase_path)

    output_folder_name = os.path.basename(os.path.splitext(testcase_path)[0])
    HtmlTestRunner.HTMLTestRunner(output=output_folder_name).run(task_suite)
