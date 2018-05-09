# encoding: utf-8

import copy
import sys
import unittest

from httprunner import exception, logger, runner, testcase, utils
from httprunner.compat import is_py3
from httprunner.report import HtmlTestResult, get_summary, render_html_report
from httprunner.testcase import TestcaseLoader
from httprunner.utils import load_dot_env_file


class TestCase(unittest.TestCase):
    """ create a testcase.
    """
    def __init__(self, test_runner, testcase_dict):
        super(TestCase, self).__init__()
        self.test_runner = test_runner
        self.testcase_dict = copy.copy(testcase_dict)

    def runTest(self):
        """ run testcase and check result.
        """
        try:
            self.test_runner.run_test(self.testcase_dict)
        finally:
            self.meta_data = getattr(self.test_runner.http_client_session, "meta_data", {})

class TestSuite(unittest.TestSuite):
    """ create test suite with a testset, it may include one or several testcases.
        each suite should initialize a separate Runner() with testset config.
    @param
        (dict) testset
            {
                "name": "testset description",
                "config": {
                    "name": "testset description",
                    "requires": [],
                    "function_binds": {},
                    "parameters": {},
                    "variables": [],
                    "request": {},
                    "output": []
                },
                "testcases": [
                    {
                        "name": "testcase description",
                        "parameters": {},
                        "variables": [],    # optional, override
                        "request": {},
                        "extract": {},      # optional
                        "validate": {}      # optional
                    },
                    testcase12
                ]
            }
        (dict) variables_mapping:
            passed in variables mapping, it will override variables in config block
    """
    def __init__(self, testset, variables_mapping=None, http_client_session=None):
        super(TestSuite, self).__init__()
        self.test_runner_list = []

        config_dict = testset.get("config", {})
        self.output_variables_list = config_dict.get("output", [])
        self.testset_file_path = config_dict.get("path")
        config_dict_parameters = config_dict.get("parameters", [])

        config_dict_variables = config_dict.get("variables", [])
        variables_mapping = variables_mapping or {}
        config_dict_variables = utils.override_variables_binds(config_dict_variables, variables_mapping)

        config_parametered_variables_list = self._get_parametered_variables(
            config_dict_variables,
            config_dict_parameters
        )
        self.testcase_parser = testcase.TestcaseParser()
        testcases = testset.get("testcases", [])

        for config_variables in config_parametered_variables_list:
            # config level
            config_dict["variables"] = config_variables
            test_runner = runner.Runner(config_dict, http_client_session)

            for testcase_dict in testcases:
                testcase_dict = copy.copy(testcase_dict)
                # testcase level
                testcase_parametered_variables_list = self._get_parametered_variables(
                    testcase_dict.get("variables", []),
                    testcase_dict.get("parameters", [])
                )
                for testcase_variables in testcase_parametered_variables_list:
                    testcase_dict["variables"] = testcase_variables

                    # eval testcase name with bind variables
                    variables = utils.override_variables_binds(
                        config_variables,
                        testcase_variables
                    )
                    self.testcase_parser.update_binded_variables(variables)
                    try:
                        testcase_name = self.testcase_parser.eval_content_with_bindings(testcase_dict["name"])
                    except (AssertionError, exception.ParamsError):
                        logger.log_warning("failed to eval testcase name: {}".format(testcase_dict["name"]))
                        testcase_name = testcase_dict["name"]
                    self.test_runner_list.append((test_runner, variables))

                    self._add_test_to_suite(testcase_name, test_runner, testcase_dict)

    def _get_parametered_variables(self, variables, parameters):
        """ parameterize varaibles with parameters
        """
        cartesian_product_parameters = testcase.parse_parameters(
            parameters,
            self.testset_file_path
        ) or [{}]

        parametered_variables_list = []
        for parameter_mapping in cartesian_product_parameters:
            parameter_mapping = parameter_mapping or {}
            variables = utils.override_variables_binds(
                variables,
                parameter_mapping
            )

            parametered_variables_list.append(variables)

        return parametered_variables_list

    def _add_test_to_suite(self, testcase_name, test_runner, testcase_dict):
        if is_py3:
            TestCase.runTest.__doc__ = testcase_name
        else:
            TestCase.runTest.__func__.__doc__ = testcase_name

        test = TestCase(test_runner, testcase_dict)
        [self.addTest(test) for _ in range(int(testcase_dict.get("times", 1)))]

    @property
    def output(self):
        outputs = []

        for test_runner, variables in self.test_runner_list:
            out = test_runner.extract_output(self.output_variables_list)
            if not out:
                continue

            outputs.append({"in": variables, "out": out})

        return outputs

class TaskSuite(unittest.TestSuite):
    """ create task suite with specified testcase path.
        each task suite may include one or several test suite.
    """
    def __init__(self, testsets, mapping=None, http_client_session=None):
        """
        @params
            testsets (dict/list): testset or list of testset
                testset_dict
                or
                [
                    testset_dict_1,
                    testset_dict_2,
                    {
                        "name": "desc1",
                        "config": {},
                        "api": {},
                        "testcases": [testcase11, testcase12]
                    }
                ]
            mapping (dict):
                passed in variables mapping, it will override variables in config block
        """
        super(TaskSuite, self).__init__()
        mapping = mapping or {}

        if not testsets:
            raise exception.TestcaseNotFound

        if isinstance(testsets, dict):
            testsets = [testsets]

        self.suite_list = []
        for testset in testsets:
            suite = TestSuite(testset, mapping, http_client_session)
            self.addTest(suite)
            self.suite_list.append(suite)

    @property
    def tasks(self):
        return self.suite_list


def init_task_suite(path_or_testsets, mapping=None, http_client_session=None):
    """ initialize task suite
    """
    if not testcase.is_testsets(path_or_testsets):
        TestcaseLoader.load_test_dependencies()
        testsets = TestcaseLoader.load_testsets_by_path(path_or_testsets)
    else:
        testsets = path_or_testsets

    mapping = mapping or {}
    return TaskSuite(testsets, mapping, http_client_session)


class HttpRunner(object):

    def __init__(self, **kwargs):
        """ initialize test runner
        @param (dict) kwargs: key-value arguments used to initialize TextTestRunner
            - resultclass: HtmlTestResult or TextTestResult
            - failfast: False/True, stop the test run on the first error or failure.
            - dot_env_path: .env file path
        """
        dot_env_path = kwargs.pop("dot_env_path", None)
        load_dot_env_file(dot_env_path)

        kwargs.setdefault("resultclass", HtmlTestResult)
        self.runner = unittest.TextTestRunner(**kwargs)

    def run(self, path_or_testsets, mapping=None):
        """ start to run test with varaibles mapping
        @param path_or_testsets: YAML/JSON testset file path or testset list
            path: path could be in several type
                - absolute/relative file path
                - absolute/relative folder path
                - list/set container with file(s) and/or folder(s)
            testsets: testset or list of testset
                - (dict) testset_dict
                - (list) list of testset_dict
                    [
                        testset_dict_1,
                        testset_dict_2
                    ]
        @param (dict) mapping:
            if mapping specified, it will override variables in config block
        """
        try:
            task_suite = init_task_suite(path_or_testsets, mapping)
        except exception.TestcaseNotFound:
            logger.log_error("Testcases not found in {}".format(path_or_testsets))
            sys.exit(1)

        result = self.runner.run(task_suite)
        self.summary = get_summary(result)

        output = []
        for task in task_suite.tasks:
            output.extend(task.output)

        self.summary["output"] = output
        return self

    def gen_html_report(self, html_report_name=None, html_report_template=None):
        """ generate html report and return report path
        @param (str) html_report_name:
            output html report file name
        @param (str) html_report_template:
            report template file path, template should be in Jinja2 format
        """
        return render_html_report(
            self.summary,
            html_report_name,
            html_report_template
        )


class LocustTask(object):

    def __init__(self, path_or_testsets, locust_client, mapping=None):
        self.task_suite = init_task_suite(path_or_testsets, mapping, locust_client)

    def run(self):
        for suite in self.task_suite:
            for test in suite:
                try:
                    test.runTest()
                except exception.MyBaseError as ex:
                    from locust.events import request_failure
                    request_failure.fire(
                        request_type=test.testcase_dict.get("request", {}).get("method"),
                        name=test.testcase_dict.get("request", {}).get("url"),
                        response_time=0,
                        exception=ex
                    )
