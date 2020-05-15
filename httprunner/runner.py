import os
import time
from datetime import datetime
from typing import List, Dict, Text

from loguru import logger

from httprunner import utils, exceptions
from httprunner.client import HttpSession
from httprunner.exceptions import ValidationFailure, ParamsError
from httprunner.loader import load_project_meta, load_testcase_file
from httprunner.parser import build_url, parse_data, parse_variables_mapping
from httprunner.response import ResponseObject
from httprunner.schema import (
    TConfig,
    TStep,
    VariablesMapping,
    StepData,
    TestCaseSummary,
    TestCaseTime,
    TestCaseInOut,
    ProjectMeta,
    TestCase,
)


class HttpRunner(object):
    config: TConfig
    teststeps: List[TStep]

    session: HttpSession = None
    variables: VariablesMapping = {}
    step_datas: List[StepData] = None
    validation_results: Dict = {}
    session_variables: Dict = {}
    success: bool = True  # indicate testcase execution result
    project_meta: ProjectMeta = None
    start_at = 0
    duration = 0

    def with_project_meta(self, project_meta: ProjectMeta) -> "HttpRunner":
        self.project_meta = project_meta
        return self

    def with_session(self, session: HttpSession) -> "HttpRunner":
        self.session = session
        return self

    def with_variables(self, variables: VariablesMapping) -> "HttpRunner":
        self.variables = variables
        return self

    def __run_step_request(self, step: TStep):
        """run teststep: request"""
        step_data = StepData(name=step.name)

        # parse
        request_dict = step.request.dict()
        parsed_request_dict = parse_data(
            request_dict, step.variables, self.project_meta.functions
        )

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
            err_msg += f"status_code: {resp.status_code}\n"
            err_msg += f"headers: {resp.headers}\n"
            err_msg += f"body: {repr(resp.text)}\n"
            logger.error(err_msg)

        # extract
        extractors = step.extract
        extract_mapping = resp_obj.extract(extractors)
        step_data.export = extract_mapping

        variables_mapping = step.variables
        variables_mapping.update(extract_mapping)

        # validate
        validators = step.validators
        try:
            resp_obj.validate(
                validators, variables_mapping, self.project_meta.functions
            )
            self.session.data.success = True
        except ValidationFailure:
            self.session.data.success = False
            log_req_resp_details()
            raise
        finally:
            self.validation_results = resp_obj.validation_results
            # save request & response meta data
            self.session.data.validators = self.validation_results
            self.success &= self.session.data.success

        step_data.success = self.session.data.success
        step_data.data = self.session.data
        return step_data

    def __run_step_testcase(self, step):
        """run teststep: referenced testcase"""
        step_data = StepData(name=step.name)
        step_variables = step.variables

        ref_testcase_path = os.path.join(self.project_meta.PWD, step.testcase)
        case_result = (
            HttpRunner()
            .with_session(self.session)
            .with_variables(step_variables)
            .run_path(ref_testcase_path)
        )
        step_data.data = case_result.step_datas  # list of step data
        step_data.export = case_result.get_export_variables()
        step_data.success = case_result.success
        self.success &= case_result.success

        return step_data

    def __run_step(self, step: TStep):
        """run teststep, teststep maybe a request or referenced testcase"""
        logger.info(f"run step: {step.name}")

        if step.request:
            step_data = self.__run_step_request(step)
        elif step.testcase:
            step_data = self.__run_step_testcase(step)
        else:
            raise ParamsError(
                f"teststep is neither a request nor a referenced testcase: {step.dict()}"
            )

        self.step_datas.append(step_data)
        return step_data.export

    def run(self, testcase: TestCase):
        """main entrance"""
        self.config = testcase.config
        self.teststeps = testcase.teststeps
        self.config.variables.update(self.variables)

        if self.config.path:
            self.project_meta = load_project_meta(self.config.path)
        else:
            self.project_meta = ProjectMeta()

        self.start_at = time.time()
        self.step_datas: List[StepData] = []
        self.session_variables = {}
        for step in self.teststeps:
            # update with config variables
            step.variables.update(self.config.variables)
            # update with session variables extracted from former step
            step.variables.update(self.session_variables)
            # parse variables
            step.variables = parse_variables_mapping(
                step.variables, self.project_meta.functions
            )
            # run step
            extract_mapping = self.__run_step(step)
            # save extracted variables to session variables
            self.session_variables.update(extract_mapping)

        self.duration = time.time() - self.start_at
        return self

    def run_path(self, path: Text) -> "HttpRunner":
        if not os.path.isfile(path):
            raise exceptions.ParamsError(f"Invalid testcase path: {path}")

        _, testcase_obj = load_testcase_file(path)
        return self.run(testcase_obj)

    def get_export_variables(self) -> Dict:
        export_vars_mapping = {}
        for var_name in self.config.export:
            if var_name not in self.session_variables:
                raise ParamsError(
                    f"failed to export variable {var_name} from session variables {self.session_variables}"
                )

            export_vars_mapping[var_name] = self.session_variables[var_name]

        return export_vars_mapping

    def get_summary(self) -> TestCaseSummary:
        """get testcase result summary"""
        start_at_timestamp = self.start_at
        start_at_iso_format = datetime.utcfromtimestamp(start_at_timestamp).isoformat()
        return TestCaseSummary(
            name=self.config.name,
            success=self.success,
            time=TestCaseTime(
                start_at=self.start_at,
                start_at_iso_format=start_at_iso_format,
                duration=self.duration,
            ),
            # status=result.status,
            # attachment=result.attachment,
            in_out=TestCaseInOut(
                vars=self.config.variables, export=self.get_export_variables()
            ),
            step_datas=self.step_datas,
        )

    def test_start(self):
        """discovered by pytest"""
        self.run(TestCase(config=self.config, teststeps=self.teststeps))
