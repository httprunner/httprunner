# encoding: utf-8

import copy
import os
import unittest

from httprunner import (exceptions, loader, logger, parser, report, runner,
                        utils, validator)


def parse_tests(testcases, variables_mapping=None):
    """ parse testcases configs, including variables/parameters/name/request.

    Args:
        testcases (list): testcase list, with config unparsed.
            [
                {   # testcase data structure
                    "config": {
                        "name": "desc1",
                        "path": "testcase1_path",
                        "variables": [],         # optional
                        "request": {}            # optional
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
        variables_mapping (dict): if variables_mapping is specified, it will override variables in config block.

    Returns:
        list: parsed testcases list, with config variables/parameters/name/request parsed.

    """
    # exception_stage = "parse tests"
    variables_mapping = variables_mapping or {}
    parsed_testcases_list = []

    for testcase in testcases:
        testcase_config = testcase.setdefault("config", {})
        project_mapping = testcase_config.pop(
            "refs",
            {
                "debugtalk": {
                    "variables": {},
                    "functions": {}
                },
                "env": {},
                "def-api": {},
                "def-testcase": {}
            }
        )

        # parse config parameters
        config_parameters = testcase_config.pop("parameters", [])
        cartesian_product_parameters_list = parser.parse_parameters(
            config_parameters,
            project_mapping["debugtalk"]["variables"],
            project_mapping["debugtalk"]["functions"]
        ) or [{}]

        for parameter_mapping in cartesian_product_parameters_list:
            testcase_dict = copy.deepcopy(testcase)
            config = testcase_dict.get("config")

            # parse config variables
            raw_config_variables = config.get("variables", [])
            parsed_config_variables = parser.parse_data(
                raw_config_variables,
                project_mapping["debugtalk"]["variables"],
                project_mapping["debugtalk"]["functions"]
            )

            # priority: passed in > debugtalk.py > parameters > variables
            # override variables mapping with parameters mapping
            config_variables = utils.override_mapping_list(
                parsed_config_variables, parameter_mapping)
            # merge debugtalk.py module variables
            config_variables.update(project_mapping["debugtalk"]["variables"])
            # override variables mapping with passed in variables_mapping
            config_variables = utils.override_mapping_list(
                config_variables, variables_mapping)

            testcase_dict["config"]["variables"] = config_variables

            # parse config name
            testcase_dict["config"]["name"] = parser.parse_data(
                testcase_dict["config"].get("name", ""),
                config_variables,
                project_mapping["debugtalk"]["functions"]
            )

            # parse config request
            testcase_dict["config"]["request"] = parser.parse_data(
                testcase_dict["config"].get("request", {}),
                config_variables,
                project_mapping["debugtalk"]["functions"]
            )

            # put loaded project functions to config
            testcase_dict["config"]["functions"] = project_mapping["debugtalk"]["functions"]
            parsed_testcases_list.append(testcase_dict)

    return parsed_testcases_list


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

    def _add_tests(self, testcases):
        """ initialize testcase with Runner() and add to test suite.

        Args:
            testcases (list): parsed testcases list

        Returns:
            tuple: unittest.TestSuite()

        """
        def __add_teststep(test_runner, config, teststep_dict):
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

        self.exception_stage = "add tests to test suite"

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
                    test_method = __add_teststep(test_runner, config, teststep_dict)
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
        self.exception_stage = "run test suite"
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
        self.exception_stage = "aggregate results"
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

    def run_tests(self, testcases, mapping=None):
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
        # parser
        parsed_testcases_list = parse_tests(testcases, mapping)

        # initialize
        test_suite = self._add_tests(parsed_testcases_list)

        # running test suite
        results = self._run_suite(test_suite)

        # aggregate
        self._aggregate(results)

        return self

    def run(self, testcase_path, mapping=None):
        """ main entrance, run testcase path with variables mapping.

        Args:
            testcase_path (str/list): testcase file/foler path.
            mapping (dict): if mapping is specified, it will override variables in config block.

        Returns:
            instance: HttpRunner() instance

        """
        testcases = loader.load_tests(testcase_path)
        return self.run_tests(testcases)

    def gen_html_report(self, html_report_name=None, html_report_template=None):
        """ generate html report and return report path.

        Args:
            html_report_name (str): output html report file name
            html_report_template (str): report template file path, template should be in Jinja2 format

        Returns:
            str: generated html report path

        """
        self.exception_stage = "generate report"
        return report.render_html_report(
            self.summary,
            html_report_name,
            html_report_template
        )


class LocustRunner(object):

    def __init__(self, locust_client):
        self.runner = HttpRunner(http_client_session=locust_client)

    def run(self, path):
        try:
            self.runner.run(path)
        except exceptions.MyBaseError as ex:
            # TODO: refactor
            from locust.events import request_failure
            request_failure.fire(
                request_type=test.testcase_dict.get("request", {}).get("method"),
                name=test.testcase_dict.get("request", {}).get("url"),
                response_time=0,
                exception=ex
            )
