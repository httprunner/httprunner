# encoding: utf-8

import os
import unittest

from httprunner import (exceptions, loader, logger, parser, report, runner,
                        utils, validator)


class HttpRunner(object):

    def __init__(self, **kwargs):
        """ initialize HttpRunner.

        Args:
            kwargs (dict): key-value arguments used to initialize TextTestRunner.
            Commonly used arguments:

            resultclass (class): HtmlTestResult or TextTestResult
            failfast (bool): False/True, stop the test run on the first error or failure.
            http_client_session (instance): requests.Session(), or locust.client.Session() instance.

        Attributes:
            project_mapping (dict): save project loaded api/testcases, environments and debugtalk.py module.
                {
                    "debugtalk": {
                        "variables": {},
                        "functions": {}
                    },
                    "env": {},
                    "def-api": {},
                    "def-testcase": {}
                }

        """
        self.exception_stage = "initialize HttpRunner()"
        self.http_client_session = kwargs.pop("http_client_session", None)
        kwargs.setdefault("resultclass", report.HtmlTestResult)
        self.unittest_runner = unittest.TextTestRunner(**kwargs)
        self.test_loader = unittest.TestLoader()
        self.summary = None

    def _add_tests(self, testcases):
        """ initialize testcase with Runner() and add to test suite.

        Args:
            testcases (list): parsed testcases list

        Returns:
            tuple: unittest.TestSuite()

        """
        def _add_teststep(test_runner, config, teststep_dict):
            """ add teststep to testcase.
            """
            def test(self):
                try:
                    test_runner.run_test(teststep_dict)
                except exceptions.MyBaseFailure as ex:
                    self.fail(str(ex))
                finally:
                    if hasattr(test_runner.http_client_session, "meta_data"):
                        self.meta_data = test_runner.http_client_session.meta_data
                        self.meta_data["validators"] = test_runner.evaluated_validators
                        test_runner.http_client_session.init_meta_data()

            try:
                teststep_dict["name"] = parser.parse_data(
                    teststep_dict["name"],
                    config.get("variables", {}),
                    config.get("functions", {})
                )
            except exceptions.VariableNotFound:
                pass

            test.__doc__ = teststep_dict["name"]
            return test

        test_suite = unittest.TestSuite()
        for testcase in testcases:
            config = testcase.get("config", {})
            test_runner = runner.Runner(config, self.http_client_session)
            TestSequense = type('TestSequense', (unittest.TestCase,), {})

            teststeps = testcase.get("teststeps", [])
            for index, teststep_dict in enumerate(teststeps):
                for times_index in range(int(teststep_dict.get("times", 1))):
                    # suppose one testcase should not have more than 9999 steps,
                    # and one step should not run more than 999 times.
                    test_method_name = 'test_{:04}_{:03}'.format(index, times_index)
                    test_method = _add_teststep(test_runner, config, teststep_dict)
                    setattr(TestSequense, test_method_name, test_method)

            loaded_testcase = self.test_loader.loadTestsFromTestCase(TestSequense)
            setattr(loaded_testcase, "config", config)
            setattr(loaded_testcase, "teststeps", testcase.get("teststeps", []))
            setattr(loaded_testcase, "runner", test_runner)
            test_suite.addTest(loaded_testcase)

        return test_suite

    def _run_suite(self, test_suite):
        """ run tests in test_suite

        Args:
            test_suite: unittest.TestSuite()

        Returns:
            list: tests_results

        """
        tests_results = []

        for testcase in test_suite:
            testcase_name = testcase.config.get("name")
            logger.log_info("Start to run testcase: {}".format(testcase_name))

            result = self.unittest_runner.run(testcase)
            tests_results.append((testcase, result))

        return tests_results

    def _aggregate(self, tests_results):
        """ aggregate results

        Args:
            tests_results (list): list of (testcase, result)

        """
        self.summary = {
            "success": True,
            "stat": {},
            "time": {},
            "platform": report.get_platform(),
            "details": []
        }

        for tests_result in tests_results:
            testcase, result = tests_result
            testcase_summary = report.get_summary(result)

            self.summary["success"] &= testcase_summary["success"]
            testcase_summary["name"] = testcase.config.get("name")
            testcase_summary["base_url"] = testcase.config.get("request", {}).get("base_url", "")

            in_out = utils.get_testcase_io(testcase)
            utils.print_io(in_out)
            testcase_summary["in_out"] = in_out

            report.aggregate_stat(self.summary["stat"], testcase_summary["stat"])
            report.aggregate_stat(self.summary["time"], testcase_summary["time"])

            self.summary["details"].append(testcase_summary)

    def _run_tests(self, testcases, mapping=None):
        """ start to run test with variables mapping.

        Args:
            testcases (list): list of testcase_dict, each testcase is corresponding to a YAML/JSON file
                [
                    {   # testcase data structure
                        "config": {
                            "name": "desc1",
                            "path": "testcase1_path",
                            "variables": [],        # optional
                            "request": {}           # optional
                            "refs": {
                                "debugtalk": {
                                    "variables": {},
                                    "functions": {}
                                },
                                "env": {},
                                "def-api": {},
                                "def-testcase": {}
                            }
                        },
                        "teststeps": [
                            # teststep data structure
                            {
                                'name': 'test step desc2',
                                'variables': [],    # optional
                                'extract': [],      # optional
                                'validate': [],
                                'request': {},
                                'function_meta': {}
                            },
                            teststep2   # another teststep dict
                        ]
                    },
                    testcase_dict_2     # another testcase dict
                ]
            mapping (dict): if mapping is specified, it will override variables in config block.

        Returns:
            instance: HttpRunner() instance

        """
        self.exception_stage = "parse tests"
        parsed_testcases_list = parser.parse_tests(testcases, mapping)

        self.exception_stage = "add tests to test suite"
        test_suite = self._add_tests(parsed_testcases_list)

        self.exception_stage = "run test suite"
        results = self._run_suite(test_suite)

        self.exception_stage = "aggregate results"
        self._aggregate(results)

        return self

    def run(self, path_or_testcases, dot_env_path=None, mapping=None):
        """ main interface, run testcases with variables mapping.

        Args:
            path_or_testcases (str/list/dict): testcase file/foler path, or valid testcases.
            dot_env_path (str): specified .env file path.
            mapping (dict): if mapping is specified, it will override variables in config block.

        Returns:
            instance: HttpRunner() instance

        """
        self.exception_stage = "load tests"

        if validator.is_testcases(path_or_testcases):
            if isinstance(path_or_testcases, dict):
                testcases = [path_or_testcases]
            else:
                testcases = path_or_testcases
        elif validator.is_testcase_path(path_or_testcases):
            testcases = loader.load_tests(path_or_testcases, dot_env_path)
        else:
            raise exceptions.ParamsError("invalid testcase path or testcases.")

        return self._run_tests(testcases, mapping)

    def gen_html_report(self, html_report_name=None, html_report_template=None):
        """ generate html report and return report path.

        Args:
            html_report_name (str): output html report file name
            html_report_template (str): report template file path, template should be in Jinja2 format

        Returns:
            str: generated html report path

        """
        if not self.summary:
            raise exceptions.MyBaseError("run method should be called before gen_html_report.")

        self.exception_stage = "generate report"
        return report.render_html_report(
            self.summary,
            html_report_name,
            html_report_template
        )
