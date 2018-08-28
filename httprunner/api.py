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
            dot_env_path (str): .env file path.

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
        loader.dot_env_path = kwargs.pop("dot_env_path", None)
        self.http_client_session = kwargs.pop("http_client_session", None)
        self.kwargs = kwargs

    def load_tests(self, path_or_testcases):
        """ load testcases, extend and merge with api/testcase definitions.

        Args:
            path_or_testcases (str/dict/list): YAML/JSON testcase file path or testcase list
                path (str): testcase file/folder path
                testcases (dict/list): testcase dict or list of testcases

        Returns:
            list: valid testcases list.

                [
                    # testcase data structure
                    {
                        "config": {
                            "name": "desc1",
                            "path": "",         # optional
                            "variables": [],    # optional
                            "request": {}       # optional
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
                    {}  # another testcase dict
                ]

        """
        if validator.is_testcases(path_or_testcases):
            # TODO: refactor
            if isinstance(path_or_testcases, list):
                for testcase in path_or_testcases:
                    try:
                        test_path = os.path.dirname(testcase["config"]["path"])
                    except KeyError:
                        test_path = os.getcwd()
                    loader.load_project_tests(test_path)
            else:
                try:
                    test_path = os.path.dirname(path_or_testcases["config"]["path"])
                except KeyError:
                    test_path = os.getcwd()
                loader.load_project_tests(test_path)

            testcases = path_or_testcases
        else:
            testcases = loader.load_testcases(path_or_testcases)

        self.project_mapping = loader.project_mapping

        if not testcases:
            raise exceptions.TestcaseNotFound

        if isinstance(testcases, dict):
            testcases = [testcases]

        return testcases

    def parse_tests(self, testcases, variables_mapping=None):
        """ parse testcases configs, including variables/parameters/name/request.

        Args:
            testcases (list): testcase list, with config unparsed.
            variables_mapping (dict): if variables_mapping is specified, it will override variables in config block.

        Returns:
            list: parsed testcases list, with config variables/parameters/name/request parsed.

        """
        variables_mapping = variables_mapping or {}

        parsed_testcases_list = []
        for testcase in testcases:

            config = testcase.setdefault("config", {})

            # parse config parameters
            config_parameters = config.pop("parameters", [])
            cartesian_product_parameters_list = parser.parse_parameters(
                config_parameters,
                self.project_mapping["debugtalk"]["variables"],
                self.project_mapping["debugtalk"]["functions"]
            ) or [{}]

            for parameter_mapping in cartesian_product_parameters_list:
                # parse config variables
                raw_config_variables = config.get("variables", [])
                parsed_config_variables = parser.parse_data(
                    raw_config_variables,
                    self.project_mapping["debugtalk"]["variables"],
                    self.project_mapping["debugtalk"]["functions"]
                )

                # priority: passed in > debugtalk.py > parameters > variables
                # override variables mapping with parameters mapping
                config_variables = utils.override_mapping_list(
                    parsed_config_variables, parameter_mapping)
                # merge debugtalk.py module variables
                config_variables.update(self.project_mapping["debugtalk"]["variables"])
                # override variables mapping with passed in variables_mapping
                config_variables = utils.override_mapping_list(
                    config_variables, variables_mapping)

                testcase["config"]["variables"] = config_variables

                # parse config name
                testcase["config"]["name"] = parser.parse_data(
                    testcase["config"].get("name", ""),
                    config_variables,
                    self.project_mapping["debugtalk"]["functions"]
                )

                # parse config request
                testcase["config"]["request"] = parser.parse_data(
                    testcase["config"].get("request", {}),
                    config_variables,
                    self.project_mapping["debugtalk"]["functions"]
                )

                # put loaded project functions to config
                testcase["config"]["functions"] = self.project_mapping["debugtalk"]["functions"]
                parsed_testcases_list.append(testcase)

        return parsed_testcases_list

    def __initialize(self, testcases):
        """ initialize test runner with parsed testcases.

        Args:
            testcases (list): testcases list

        Returns:
            tuple: (unittest.TextTestRunner(), unittest.TestSuite())

        """
        self.kwargs.setdefault("resultclass", report.HtmlTestResult)
        unittest_runner = unittest.TextTestRunner(**self.kwargs)

        testcases_list = []
        loader = unittest.TestLoader()
        loaded_testcases = []
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
                    test_method = utils.add_teststep(test_runner, teststep_dict)
                    setattr(TestSequense, test_method_name, test_method)

            loaded_testcase = loader.loadTestsFromTestCase(TestSequense)
            setattr(loaded_testcase, "config", config)
            setattr(loaded_testcase, "teststeps", testcase.get("teststeps", []))
            setattr(loaded_testcase, "runner", test_runner)
            loaded_testcases.append(loaded_testcase)

        test_suite = unittest.TestSuite(loaded_testcases)
        return (unittest_runner, test_suite)

    def run(self, path_or_testcases, mapping=None):
        """ start to run test with variables mapping.

        Args:
            path_or_testcases (str/list/dict): YAML/JSON testcase file path or testcase list
                path: path could be in several type
                    - absolute/relative file path
                    - absolute/relative folder path
                    - list/set container with file(s) and/or folder(s)
                testcases: testcase dict or list of testcases
                    - (dict) testset_dict
                    - (list) list of testset_dict
                        [
                            testset_dict_1,
                            testset_dict_2
                        ]
            mapping (dict): if mapping specified, it will override variables in config block.

        Returns:
            instance: HttpRunner() instance

        """
        # loader
        testcases_list = self.load_tests(path_or_testcases)
        # parser
        parsed_testcases_list = self.parse_tests(testcases_list)

        # initialize
        unittest_runner, test_suite = self.__initialize(parsed_testcases_list)

        # aggregate
        self.summary = {
            "success": True,
            "stat": {},
            "time": {},
            "platform": report.get_platform(),
            "details": []
        }

        # execution
        for testcase in test_suite:
            testcase_name = testcase.config.get("name")
            logger.log_info("Start to run testcase: {}".format(testcase_name))

            result = unittest_runner.run(testcase)
            testcase_summary = report.get_summary(result)

            self.summary["success"] &= testcase_summary["success"]
            testcase_summary["name"] = testcase_name
            testcase_summary["base_url"] = testcase.config.get("request", {}).get("base_url", "")

            in_out = utils.get_testcase_io(testcase)
            utils.print_io(in_out)
            testcase_summary["in_out"] = in_out

            report.aggregate_stat(self.summary["stat"], testcase_summary["stat"])
            report.aggregate_stat(self.summary["time"], testcase_summary["time"])

            self.summary["details"].append(testcase_summary)

        return self

    def gen_html_report(self, html_report_name=None, html_report_template=None):
        """ generate html report and return report path.

        Args:
            html_report_name (str): output html report file name
            html_report_template (str): report template file path, template should be in Jinja2 format

        Returns:
            str: generated html report path

        """
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
