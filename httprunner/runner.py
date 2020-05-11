from typing import List, Dict

from loguru import logger

from httprunner import utils
from httprunner.client import HttpSession
from httprunner.exceptions import ValidationFailure, ParamsError
from httprunner.parser import build_url, parse_data, parse_variables_mapping
from httprunner.response import ResponseObject
from httprunner.schema import TestsConfig, TestStep, VariablesMapping, TestCase, SessionData


class TestCaseRunner(object):

    config: TestsConfig = {}
    teststeps: List[TestStep] = []
    session: HttpSession = None
    step_datas: List[SessionData] = []
    validation_results: Dict = {}
    session_variables: Dict = {}

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

    def __run_step_request(self, step: TestStep):
        """run teststep: request"""
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

        def log_req_resp_details():
            err_msg = "\n{} DETAILED REQUEST & RESPONSE {}\n".format("*" * 32, "*" * 32)

            # log request
            err_msg += "====== request details ======\n"
            err_msg += f"url: {url}\n"
            err_msg += f"method: {method}\n"
            headers = parsed_request_dict.pop("headers", {})
            err_msg += f"headers: {headers}\n"
            for k, v in parsed_request_dict.items():
                v = utils.omit_long_data(v)
                err_msg += f"{k}: {repr(v)}\n"

            err_msg += "\n"

            # log response
            err_msg += "====== response details ======\n"
            err_msg += f"status_code: {resp_obj.status_code}\n"
            err_msg += f"headers: {resp_obj.headers}\n"
            err_msg += f"body: {repr(resp_obj.text)}\n"
            logger.error(err_msg)

        # extract
        extractors = step.extract
        extract_mapping = resp_obj.extract(extractors)

        variables_mapping = step.variables
        variables_mapping.update(extract_mapping)

        # validate
        validators = step.validators
        try:
            resp_obj.validate(validators, variables_mapping, self.config.functions)
            self.session.data.status = "passed"
        except ValidationFailure:
            self.session.data.status = "failed"
            log_req_resp_details()
            raise
        finally:
            self.validation_results = resp_obj.validation_results
            # save request & response meta data
            self.session.data.validators = self.validation_results
            self.session.data.name = step.name
            self.step_datas.append(self.session.data)

        return extract_mapping

    def __run_step_testcase(self, step):
        """run teststep: referenced testcase"""
        step_variables = step.variables
        testcase: TestCaseRunner = step.testcase
        res = testcase.with_variables(**step_variables).run()
        return res.get_export_variables()

    def __run_step(self, step: TestStep):
        logger.info(f"run step: {step.name}")
        if step.request:
            return self.__run_step_request(step)
        elif step.testcase:
            return self.__run_step_testcase(step)

    def test_start(self):
        """main entrance"""
        self.step_datas.clear()
        self.session_variables.clear()
        for step in self.teststeps:
            # update with config variables
            step.variables.update(self.config.variables)
            # update with session variables extracted from former step
            step.variables.update(self.session_variables)
            # parse variables
            step.variables = parse_variables_mapping(step.variables, self.config.functions)
            # run step
            extract_mapping = self.__run_step(step)
            # save extracted variables to session variables
            self.session_variables.update(extract_mapping)

        return self

    def run(self):
        """main entrance alias for test_start"""
        return self.test_start()

    def get_export_variables(self):
        export_vars_mapping = {}
        for var_name in self.config.export:
            if var_name not in self.session_variables:
                raise ParamsError(
                    f"failed to export variable {var_name} from session variables {self.session_variables}")

            export_vars_mapping[var_name] = self.session_variables[var_name]

        return export_vars_mapping
