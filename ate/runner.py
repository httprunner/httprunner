import requests

from ate import exception, utils
from ate.context import Context
from ate.testcase import TestcaseParser


class TestRunner(object):

    def __init__(self):
        self.client = requests.Session()
        self.context = Context()
        self.testcase_parser = TestcaseParser()

    def prepare(self, testcase):
        """ prepare work before running test.
        parse testcase with variables binds if it is a template.
        """
        requires = testcase.get('requires', [])
        self.context.import_requires(requires)

        function_binds = testcase.get('function_binds', {})
        self.context.bind_functions(function_binds)

        variable_binds = testcase.get('variable_binds', [])
        self.context.bind_variables(variable_binds)

        parsed_testcase = self.testcase_parser.parse(
            testcase,
            variables_binds=self.context.variables
        )
        return parsed_testcase

    def run_single_testcase(self, testcase):
        testcase = self.prepare(testcase)

        req_kwargs = testcase['request']

        try:
            url = req_kwargs.pop('url')
            method = req_kwargs.pop('method')
        except KeyError:
            raise exception.ParamsError("URL or METHOD missed!")

        resp_obj = self.client.request(url=url, method=method, **req_kwargs)
        diff_content = utils.diff_response(resp_obj, testcase['response'])
        success = False if diff_content else True
        return success, diff_content

    def run_testcase_suite(self, testcase_sets):
        return [
            self.run_single_testcase(testcase)
            for testcase in testcase_sets
        ]
