import copy
import sys
import unittest

from httprunner import exception, logger, runner, testcase, utils
from httprunner.compat import is_py3
from httprunner.report import HtmlTestResult, get_summary


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
        self.testset_file_path = config_dict["path"]
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
                    testcase_name = self.testcase_parser.eval_content_with_bindings(testcase_dict["name"])
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
            outputs.append(
                {
                    "in": variables,
                    "out": test_runner.extract_output(self.output_variables_list)
                }
            )

        return outputs

class TaskSuite(unittest.TestSuite):
    """ create task suite with specified testcase path.
        each task suite may include one or several test suite.
    """
    def __init__(self, path, mapping=None, http_client_session=None):
        """
        @params
            path: path could be in several type
                - absolute/relative file path
                - absolute/relative folder path
                - list/set container with file(s) and/or folder(s)
            (dict) mapping:
                passed in variables mapping, it will override variables in config block
        """
        super(TaskSuite, self).__init__()
        mapping = mapping or {}

        if not isinstance(path, list):
            # absolute/relative file/folder path
            path = [path]

        # remove duplicate path
        path = set(path)

        testsets = testcase.load_testcases_by_path(path)
        if not testsets:
            raise exception.TestcaseNotFound

        self.suite_list = []
        for testset in testsets:
            suite = TestSuite(testset, mapping, http_client_session)
            self.addTest(suite)
            self.suite_list.append(suite)

    @property
    def tasks(self):
        return self.suite_list


class HttpRunner(object):

    def __init__(self, path, gen_html_report=True, **kwargs):
        """ initialize HttpRunner with specified testset file path and test runner
        @param (str) path:
            YAML/JSON testset file path
        @param (boolean) gen_html_report:
            True: use HtmlTestResult and generate html report
            False: use TextTestResult and do not generate report file
        @param (dict) kwargs:
            key-value arguments used to initialize TextTestRunner
                - failfast: False/True, stop the test run on the first error or failure.
        """
        self.path = path

        self.gen_html_report = gen_html_report
        if self.gen_html_report:
            kwargs["resultclass"] = HtmlTestResult

        self.runner = unittest.TextTestRunner(**kwargs)

    def run(self, mapping=None, html_report_name=None, html_report_template=None):
        """ start to run suite
        @param (dict) mapping:
            if mapping specified, it will override variables in config block
        @param (str) html_report_name:
            output html report file name
        @param (str) html_report_template:
            report template file path, template should be in Jinja2 format
        """
        try:
            mapping = mapping or {}
            task_suite = TaskSuite(self.path, mapping)
        except exception.TestcaseNotFound:
            logger.log_error("Testcases not found in {}".format(self.path))
            sys.exit(1)

        result = self.runner.run(task_suite)

        output = []
        for task in task_suite.tasks:
            output.extend(task.output)

        if self.gen_html_report:
            summary = result.summary
            summary["report_path"] = result.render_html_report(
                html_report_name,
                html_report_template
            )
        else:
            summary = get_summary(result)

        summary["output"] = output
        return summary


class LocustTask(object):

    def __init__(self, path, locust_client, mapping=None):
        mapping = mapping or {}
        self.task_suite = TaskSuite(path, mapping, locust_client)

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
