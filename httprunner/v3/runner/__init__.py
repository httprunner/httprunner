from typing import List

import requests

from httprunner.v3.parser import build_url
from httprunner.v3.schema import TestsConfig, TestStep


class TestCaseRunner(object):

    config: TestsConfig = {}
    teststeps: List[TestStep] = []
    session: requests.Session = None

    def with_session(self, s: requests.Session) -> "TestCaseRunner":
        self.session = s
        return self

    def with_variables(self, **variables) -> "TestCaseRunner":
        self.config.variables.update(variables)
        return self

    def run_step(self, step):
        request_dict = step.request.dict()

        method = request_dict.pop("method")
        url_path = request_dict.pop("url")
        url = build_url(self.config.base_url, url_path)

        request_dict["json"] = request_dict.pop("req_json", {})

        session = self.session or requests.Session()
        resp = session.request(method, url, **request_dict)

    def run(self):
        for step in self.teststeps:
            step.variables.update(self.config.variables)
            self.run_step(step)
