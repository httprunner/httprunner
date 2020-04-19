from typing import List

import requests

from loguru import logger

from httprunner.v3.parser import build_url
from httprunner.v3.schema import TestsConfig, TestStep
from httprunner.v3.validator import Validator
from httprunner.v3.response import ResponseObject


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
        logger.info(f"run step: {step.name}")

        # prepare arguments
        request_dict = step.request.dict()
        method = request_dict.pop("method")
        url_path = request_dict.pop("url")
        url = build_url(self.config.base_url, url_path)

        request_dict["json"] = request_dict.pop("req_json", {})

        logger.info(f"{method} {url}")
        logger.debug(f"request kwargs(raw): {request_dict}")

        # request
        session = self.session or requests.Session()
        resp = session.request(method, url, **request_dict)

        # validate
        resp_obj = ResponseObject(resp)
        validator = Validator(resp_obj)
        validators = step.validation
        validator.validate(validators)

    def test_start(self):
        """main entrance"""
        for step in self.teststeps:
            step.variables.update(self.config.variables)
            self.run_step(step)

    def run(self):
        """main entrance alias for test_start"""
        return self.test_start()
