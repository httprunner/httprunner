import unittest

from ate import runner, testcase, utils


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
        self.assertTrue(self.test_runner.run_test(self.testcase))

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
        for testcase in testcases:
            if utils.PYTHON_VERSION == 3:
                ApiTestCase.runTest.__doc__ = testcase['name']
            else:
                ApiTestCase.runTest.__func__.__doc__ = testcase['name']

            test = ApiTestCase(self.test_runner, testcase)
            self.addTest(test)

    def print_output(self):
        output_variables_list = self.config_dict.get("output", [])
        self.test_runner.generate_output(output_variables_list)


def create_task(testcase_path):
    """ create test task suite with specified testcase path.
        each task suite may include one or several test suite.
    """
    task_suite = unittest.TestSuite()
    testsets = testcase.load_testcases_by_path(testcase_path)

    for testset in testsets:
        suite = ApiTestSuite(testset)
        task_suite.addTest(suite)

    return task_suite
