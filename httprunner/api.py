import os
import sys
import unittest
from typing import List

from loguru import logger

from httprunner import report, loader, utils, exceptions, __version__
from httprunner.report import gen_html_report
from httprunner.runner import HttpRunner as TestCaseRunner
from httprunner.schema import TestsMapping, TestCaseSummary, TestSuiteSummary


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

    def __init__(self, save_tests=False, log_level="WARNING", log_file=None):
        """ initialize HttpRunner.

        Args:
            save_tests (bool): save loaded/parsed tests to JSON file.
            log_level (str): logging level.
            log_file (str): log file path.

        """
        self.exception_stage = "initialize HttpRunner()"
        kwargs = {"failfast": True, "resultclass": report.HtmlTestResult}

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
                    self.step_datas = test_runner.step_datas

            test.__doc__ = test_runner.config.name
            return test

        project_meta = tests.project_meta
        testcases = tests.testcases

        prepared_testcases: List[unittest.TestSuite] = []

        for testcase in testcases:
            testcase.config.variables.update(project_meta.variables)
            test_runner = TestCaseRunner(testcase.config, testcase.teststeps)

            TestSequense = type("TestSequense", (unittest.TestCase,), {})
            test_method = _add_test(test_runner)
            setattr(TestSequense, "test_method_name", test_method)

            loaded_testcase = self.test_loader.loadTestsFromTestCase(TestSequense)
            setattr(loaded_testcase, "config", testcase.config)
            prepared_testcases.append(loaded_testcase)

        return prepared_testcases

    def _run_suite(
        self, prepared_testcases: List[unittest.TestSuite]
    ) -> List[TestCaseSummary]:
        """ run prepared testcases
        """
        tests_results: List[TestCaseSummary] = []

        for index, testcase in enumerate(prepared_testcases):
            log_handler = None
            if self.save_tests:
                logs_file_abs_path = utils.prepare_log_file_abs_path(
                    self.test_path, f"testcase_{index+1}.log"
                )
                log_handler = logger.add(logs_file_abs_path, level="DEBUG")

            logger.info(f"Start to run testcase: {testcase.config.name}")

            result = self.unittest_runner.run(testcase)
            testcase_summary = report.get_summary(result)
            testcase_summary.in_out.vars = testcase.config.variables
            testcase_summary.in_out.out = testcase.config.export

            if self.save_tests and log_handler:
                logger.remove(log_handler)
                logs_file_abs_path = utils.prepare_log_file_abs_path(
                    self.test_path, f"testcase_{index+1}.log"
                )
                testcase_summary.log = logs_file_abs_path

            if result.wasSuccessful():
                tests_results.append(testcase_summary)
            else:
                tests_results.insert(0, testcase_summary)

        return tests_results

    def _aggregate(self, tests_results: List[TestCaseSummary]) -> TestSuiteSummary:
        """ aggregate multiple testcase results

        Args:
            tests_results (list): list of testcase summary

        """
        testsuite_summary = TestSuiteSummary(
            success=True, platform=report.get_platform(), testcases=[]
        )
        testsuite_summary.stat.total = len(tests_results)
        testsuite_summary.stat.success = 0
        testsuite_summary.stat.fail = 0

        for testcase_summary in tests_results:
            if testcase_summary.success:
                testsuite_summary.stat.success += 1
            else:
                testsuite_summary.stat.fail += 1

            testsuite_summary.success &= testcase_summary.success
            testsuite_summary.testcases.append(testcase_summary)

        total_duration = (
            tests_results[-1].time.start_at
            + tests_results[-1].time.duration
            - tests_results[0].time.start_at
        )

        testsuite_summary.time.start_at = tests_results[0].time.start_at
        testsuite_summary.time.start_at_iso_format = tests_results[
            0
        ].time.start_at_iso_format
        testsuite_summary.time.duration = total_duration

        return testsuite_summary

    def run_tests(self, tests_mapping) -> TestSuiteSummary:
        """ run testcase/testsuite data
        """
        tests = TestsMapping.parse_obj(tests_mapping)
        self.test_path = tests.project_meta.test_path

        if self.save_tests:
            utils.dump_json_file(
                tests_mapping,
                utils.prepare_log_file_abs_path(self.test_path, "loaded.json"),
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

        if self.save_tests:
            utils.dump_json_file(
                self._summary.dict(),
                utils.prepare_log_file_abs_path(self.test_path, "summary.json"),
            )
            # save variables and export data
            vars_out = self.get_vars_out()
            utils.dump_json_file(
                vars_out, utils.prepare_log_file_abs_path(self.test_path, "io.json")
            )

        return self._summary

    def get_vars_out(self):
        """ get variables and output
        Returns:
            list: list of variables and output.
                if tests are parameterized, list items are corresponded to parameters.

                [
                    {
                        "in": {
                            "user1": "leo"
                        },
                        "out": {
                            "out1": "out_value_1"
                        }
                    },
                    {...}
                ]

            None: returns None if tests not started or finished or corrupted.

        """
        if not self._summary:
            return None

        return [
            testcase_summary.in_out.dict()
            for testcase_summary in self._summary.testcases
        ]

    def run_path(self, path, dot_env_path=None, mapping=None) -> TestSuiteSummary:
        """ run testcase/testsuite file or folder.

        Args:
            path (str): testcase/testsuite file/foler path.
            dot_env_path (str): specified .env file path.
            mapping (dict): if mapping is specified, it will override variables in config block.

        Returns:
            dict: result summary

        """
        # load tests
        logger.info(f"HttpRunner version: {__version__}")
        self.exception_stage = "load tests"
        tests_mapping = loader.load_cases(path, dot_env_path)

        if mapping:
            tests_mapping["project_meta"]["variables"] = mapping

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
        if loader.is_test_path(path_or_tests):
            return self.run_path(path_or_tests, dot_env_path, mapping)

        project_working_directory = path_or_tests.get("project_meta", {}).get(
            "PWD", os.getcwd()
        )
        loader.init_pwd(project_working_directory)
        return self.run_tests(path_or_tests)

    def gen_html_report(self, report_template=None, report_dir=None, report_file=None):
        if not self._summary:
            return None

        return gen_html_report(self._summary, report_template, report_dir, report_file)
