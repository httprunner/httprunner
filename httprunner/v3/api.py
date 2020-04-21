import os
import sys
import unittest
from typing import List, Tuple

from loguru import logger

from httprunner import report, loader, utils, exceptions, __version__
from httprunner.v3.runner import TestCaseRunner
from httprunner.v3.schema import TestsMapping


class HttpRunner(object):
    """ Developer Interface: Main Interface
        Usage:

            from httprunner.api import HttpRunner
            runner = HttpRunner(
                failfast=True,
                save_tests=True,
                log_level="INFO",
                log_file="test.log"
            )
            summary = runner.run(path_or_tests)

    """

    def __init__(self, failfast=False, save_tests=False, log_level="WARNING", log_file=None):
        """ initialize HttpRunner.

        Args:
            failfast (bool): stop the test run on the first error or failure.
            save_tests (bool): save loaded/parsed tests to JSON file.
            log_level (str): logging level.
            log_file (str): log file path.

        """
        self.exception_stage = "initialize HttpRunner()"
        kwargs = {
            "failfast": failfast,
            "resultclass": report.HtmlTestResult
        }

        logger.remove()
        log_level = log_level.upper()
        logger.add(sys.stdout, level=log_level)
        if log_file:
            logger.add(log_file, level=log_level)

        self.unittest_runner = unittest.TextTestRunner(**kwargs)
        self.test_loader = unittest.TestLoader()
        self.save_tests = save_tests
        self._summary = None
        self.test_path = None

    def _prepare_tests(self, tests: TestsMapping) -> List[unittest.TestSuite]:
        def _add_test(test_runner: TestCaseRunner):
            """ add test to testcase.
            """
            def test(self):
                try:
                    test_runner.run()
                except exceptions.MyBaseFailure as ex:
                    self.fail(str(ex))
                finally:
                    self.meta_datas = test_runner.meta_datas

            test.__doc__ = test_runner.config.name
            return test

        project_meta = tests.project_mapping
        testcases = tests.testcases

        prepared_testcases: List[unittest.TestSuite] = []

        for testcase in testcases:
            testcase.config.variables.update(project_meta.variables)
            testcase.config.functions.update(project_meta.functions)

            test_runner = TestCaseRunner().init(testcase)

            TestSequense = type('TestSequense', (unittest.TestCase,), {})
            test_method = _add_test(test_runner)
            setattr(TestSequense, "test_method_name", test_method)

            loaded_testcase = self.test_loader.loadTestsFromTestCase(TestSequense)
            setattr(loaded_testcase, "config", testcase.config)
            # setattr(loaded_testcase, "teststeps", testcase.teststeps)
            # setattr(loaded_testcase, "runner", test_runner)
            prepared_testcases.append(loaded_testcase)

        return prepared_testcases

    def _run_suite(self, prepared_testcases: List[unittest.TestSuite]) -> List[Tuple]:
        """ run prepared testcases
        """
        tests_results: List[Tuple] = []

        for index, testcase in enumerate(prepared_testcases):
            log_handler = None
            if self.save_tests:
                logs_file_abs_path = utils.prepare_log_file_abs_path(
                    self.test_path, f"testcase_{index+1}.log"
                )
                log_handler = logger.add(logs_file_abs_path, level="DEBUG")

            logger.info(f"Start to run testcase: {testcase.config.name}")

            result = self.unittest_runner.run(testcase)
            if result.wasSuccessful():
                tests_results.append((testcase, result))
            else:
                tests_results.insert(0, (testcase, result))

            if self.save_tests and log_handler:
                logger.remove(log_handler)

        return tests_results

    def _aggregate(self, tests_results: List[Tuple]):
        """ aggregate results

        Args:
            tests_results (list): list of (testcase, result)

        """
        summary = {
            "success": True,
            "stat": {
                "testcases": {
                    "total": len(tests_results),
                    "success": 0,
                    "fail": 0
                },
                "teststeps": {}
            },
            "time": {},
            "platform": report.get_platform(),
            "details": []
        }

        for index, tests_result in enumerate(tests_results):
            testcase, result = tests_result
            testcase_summary = report.get_summary(result)

            if testcase_summary["success"]:
                summary["stat"]["testcases"]["success"] += 1
            else:
                summary["stat"]["testcases"]["fail"] += 1

            summary["success"] &= testcase_summary["success"]
            testcase_summary["name"] = testcase.config.name
            testcase_summary["in_out"] = {
                "in": testcase.config.variables,
                "out": testcase.config.export
            }

            report.aggregate_stat(summary["stat"]["teststeps"], testcase_summary["stat"])
            report.aggregate_stat(summary["time"], testcase_summary["time"])

            if self.save_tests:
                logs_file_abs_path = utils.prepare_log_file_abs_path(
                    self.test_path, f"testcase_{index+1}.log"
                )
                testcase_summary["log"] = logs_file_abs_path

            summary["details"].append(testcase_summary)

        return summary

    def run_tests(self, tests_mapping):
        """ run testcase/testsuite data
        """
        tests = TestsMapping.parse_obj(tests_mapping)
        self.test_path = tests.project_mapping.test_path

        if self.save_tests:
            utils.dump_json_file(
                tests_mapping,
                utils.prepare_log_file_abs_path(self.test_path, "loaded.json")
            )

        # prepare testcases
        self.exception_stage = "prepare testcases"
        prepared_testcases = self._prepare_tests(tests)

        # run prepared testcases
        self.exception_stage = "run prepared testcases"
        results = self._run_suite(prepared_testcases)

        # aggregate results
        self.exception_stage = "aggregate results"
        self._summary = self._aggregate(results)

        # generate html report
        self.exception_stage = "generate html report"
        report.stringify_summary(self._summary)

        if self.save_tests:
            utils.dump_json_file(
                self._summary,
                utils.prepare_log_file_abs_path(self.test_path, "summary.json")
            )
            # save variables and export data
            vars_out = self.get_vars_out()  # TODO
            utils.dump_json_file(
                vars_out,
                utils.prepare_log_file_abs_path(self.test_path, "io.json")
            )

        return self._summary

    def run_path(self, path, dot_env_path=None, mapping=None):
        """ run testcase/testsuite file or folder.

        Args:
            path (str): testcase/testsuite file/foler path.
            dot_env_path (str): specified .env file path.
            mapping (dict): if mapping is specified, it will override variables in config block.

        Returns:
            dict: result summary

        """
        # load tests
        self.exception_stage = "load tests"
        tests_mapping = loader.load_cases(path, dot_env_path)

        if mapping:
            tests_mapping["project_mapping"]["variables"] = mapping

        return self.run_tests(tests_mapping)

    def run(self, path_or_tests, dot_env_path=None, mapping=None):
        """ main interface.

        Args:
            path_or_tests:
                str: testcase/testsuite file/foler path
                dict: valid testcase/testsuite data
            dot_env_path (str): specified .env file path.
            mapping (dict): if mapping is specified, it will override variables in config block.

        Returns:
            dict: result summary

        """
        logger.info(f"HttpRunner version: {__version__}")
        if loader.is_test_path(path_or_tests):
            return self.run_path(path_or_tests, dot_env_path, mapping)
        elif loader.is_test_content(path_or_tests):
            project_working_directory = path_or_tests.get("project_mapping", {}).get("PWD", os.getcwd())
            loader.init_pwd(project_working_directory)
            return self.run_tests(path_or_tests)
        else:
            raise exceptions.ParamsError(f"Invalid testcase path or testcases: {path_or_tests}")
