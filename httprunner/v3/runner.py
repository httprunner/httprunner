from typing import List

import requests
from loguru import logger

from httprunner.v3.parser import build_url, parse_content, parse_variables_mapping
from httprunner.v3.response import ResponseObject
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
        logger.info(f"run step: {step.name}")

        # parse
        request_dict = step.request.dict()
        parsed_request_dict = parse_content(request_dict, step.variables, self.config.functions)

        # prepare arguments
        method = parsed_request_dict.pop("method")
        url_path = parsed_request_dict.pop("url")
        url = build_url(self.config.base_url, url_path)

        parsed_request_dict["json"] = parsed_request_dict.pop("req_json", {})

        logger.info(f"{method} {url}")
        logger.debug(f"request kwargs(raw): {parsed_request_dict}")

        # request
        session = self.session or requests.Session()
        resp = session.request(method, url, **parsed_request_dict)
        resp_obj = ResponseObject(resp)

        # validate
        validators = step.validation
        resp_obj.validate(validators)

        # extract
        extractors = step.extract
        extract_mapping = resp_obj.extract(extractors)
        return extract_mapping

    def test_start(self):
        """main entrance"""
        session_variables = {}
        for step in self.teststeps:
            # update with config variables
            step.variables.update(self.config.variables)
            # update with session variables extracted from former step
            step.variables.update(session_variables)
            # parse variables
            step.variables = parse_variables_mapping(step.variables, self.config.functions)
            # run step
            extract_mapping = self.run_step(step)
            # save extracted variables to session variables
            session_variables.update(extract_mapping)

    def run(self):
        """main entrance alias for test_start"""
        return self.test_start()
