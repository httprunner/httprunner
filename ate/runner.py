import requests
from ate import utils, exception

class TestRunner(object):

    def __init__(self):
        self.client = requests.Session()

    def run_single_testcase(self, testcase):
        req_kwargs = testcase['request']

        try:
            url = req_kwargs.pop('url')
            method = req_kwargs.pop('method')
        except KeyError:
            raise exception.ParamsError("Params Error")

        resp_obj = self.client.request(url=url, method=method, **req_kwargs)
        diff_content = utils.diff_response(resp_obj, testcase['response'])
        success = False if diff_content else True
        return success, diff_content

    def run_testcase_suite(self, testcase_sets):
        return [
            self.run_single_testcase(testcase)
            for testcase in testcase_sets
        ]
