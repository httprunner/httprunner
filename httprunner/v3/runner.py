from typing import List

from loguru import logger

from httprunner.client import HttpSession
from httprunner.v3.parser import build_url, parse_data, parse_variables_mapping
from httprunner.v3.response import ResponseObject
from httprunner.v3.schema import TestsConfig, TestStep, VariablesMapping, TestCase


class TestCaseRunner(object):

    config: TestsConfig = {}
    teststeps: List[TestStep] = []
    session: HttpSession = None
    meta_datas: List = []

    def init(self, testcase: TestCase) -> "TestCaseRunner":
        self.config = testcase.config
        self.teststeps = testcase.teststeps
        return self

    def with_session(self, s: HttpSession) -> "TestCaseRunner":
        self.session = s
        return self

    def with_variables(self, **variables: VariablesMapping) -> "TestCaseRunner":
        self.config.variables.update(variables)
        return self

    def run_step(self, step: TestStep):
        logger.info(f"run step: {step.name}")

        # parse
        request_dict = step.request.dict()
        parsed_request_dict = parse_data(request_dict, step.variables, self.config.functions)

        # prepare arguments
        method = parsed_request_dict.pop("method")
        url_path = parsed_request_dict.pop("url")
        url = build_url(self.config.base_url, url_path)

        parsed_request_dict["json"] = parsed_request_dict.pop("req_json", {})

        logger.info(f"{method} {url}")
        logger.debug(f"request kwargs(raw): {parsed_request_dict}")

        # request
        self.session = self.session or HttpSession()
        resp = self.session.request(method, url, **parsed_request_dict)
        resp_obj = ResponseObject(resp)

        # extract
        extractors = step.extract
        extract_mapping = resp_obj.extract(extractors)

        variables_mapping = step.variables
        variables_mapping.update(extract_mapping)

        # validate
        validators = step.validators
        resp_obj.validate(validators, variables_mapping, self.config.functions)

        return extract_mapping

    def test_start(self):
        """main entrance"""
        self.meta_datas.clear()
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
            # save request & response meta data
            self.session.meta_data["name"] = step.name
            self.meta_datas.append(self.session.meta_data)

        return self

    def run(self):
        """main entrance alias for test_start"""
        return self.test_start()
