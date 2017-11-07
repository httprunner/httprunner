import logging
import unittest

from httprunner import exception, runner, testcase, utils


class ApiTestCase(unittest.TestCase):
    """ create a testcase.
    """
    def __init__(self, test_runner, testcase_dict):
        super(ApiTestCase, self).__init__()
        self.test_runner = test_runner
        self.testcase_dict = testcase_dict

    def runTest(self):
        """ run testcase and check result.
        """
        self.assertTrue(self.test_runner._run_test(self.testcase_dict))

class ApiTestSuite(unittest.TestSuite):
    """ create test suite with a testset, it may include one or several testcases.
        each suite should initialize a separate Runner() with testset config.
    """
    def __init__(self, testset):
        super(ApiTestSuite, self).__init__()
        self.test_runner = runner.Runner()
        self.config_dict = testset.get("config", {})
        self.test_runner.init_config(self.config_dict, level="testset")
        testcases = testset.get("testcases", [])
        self._add_tests_to_suite(testcases)

    def _add_tests_to_suite(self, testcases):
        for testcase_dict in testcases:
            if utils.PYTHON_VERSION == 3:
                ApiTestCase.runTest.__doc__ = testcase_dict['name']
            else:
                ApiTestCase.runTest.__func__.__doc__ = testcase_dict['name']

            test = ApiTestCase(self.test_runner, testcase_dict)
            self.addTest(test)

    def print_output(self):
        output_variables_list = self.config_dict.get("output", [])
        self.test_runner.generate_output(output_variables_list)

class TaskSuite(unittest.TestSuite):
    """ create test task suite with specified testcase path.
        each task suite may include one or several test suite.
    """
    def __init__(self, testcase_path):
        super(TaskSuite, self).__init__()
        self.suite_list = []
        testsets = testcase.load_testcases_by_path(testcase_path)
        if not testsets:
            raise exception.TestcaseNotFound

        for testset in testsets:
            suite = ApiTestSuite(testset)
            self.addTest(suite)
            self.suite_list.append(suite)

    @property
    def tasks(self):
        return self.suite_list
