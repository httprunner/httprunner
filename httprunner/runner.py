import os
import time
import uuid
from datetime import datetime
from typing import List, Dict, Text

try:
    import allure

    USE_ALLURE = True
except ModuleNotFoundError:
    USE_ALLURE = False

from loguru import logger

from httprunner import utils, exceptions
from httprunner.client import HttpSession
from httprunner.exceptions import ValidationFailure, ParamsError
from httprunner.ext.uploader import prepare_upload_step
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

    success: bool = True  # indicate testcase execution result
    __project_meta: ProjectMeta = None
    __case_id: Text = ""
    __step_datas: List[StepData] = None
    __session: HttpSession = None
    __session_variables: VariablesMapping = {}
    # time
    __start_at: float = 0
    __duration: float = 0
    # log
    __log_path: Text = ""

    def with_project_meta(self, project_meta: ProjectMeta) -> "HttpRunner":
        self.__project_meta = project_meta
        return self

    def with_session(self, session: HttpSession) -> "HttpRunner":
        self.__session = session
        return self

    def with_case_id(self, case_id: Text) -> "HttpRunner":
        self.__case_id = case_id
        return self

    def with_variables(self, variables: VariablesMapping) -> "HttpRunner":
        self.__session_variables = variables
        return self

    def __run_step_request(self, step: TStep):
        """run teststep: request"""
        step_data = StepData(name=step.name)

        # parse
        prepare_upload_step(step, self.__project_meta.functions)
        request_dict = step.request.dict()
        request_dict.pop("upload", None)
        parsed_request_dict = parse_data(
            request_dict, step.variables, self.__project_meta.functions
        )
        parsed_request_dict["headers"].setdefault(
            "HRUN-Request-ID",
            f"HRUN-{self.__case_id}-{str(int(time.time() * 1000))[-6:]}",
        )

        # prepare arguments
        method = parsed_request_dict.pop("method")
        url_path = parsed_request_dict.pop("url")
        url = build_url(self.config.base_url, url_path)
        parsed_request_dict["json"] = parsed_request_dict.pop("req_json", {})

        # request
        resp = self.__session.request(method, url, **parsed_request_dict)
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
                validators, variables_mapping, self.__project_meta.functions
            )
            self.__session.data.success = True
        except ValidationFailure:
            self.__session.data.success = False
            log_req_resp_details()
            raise
        finally:
            # save request & response meta data
            self.__session.data.validators = resp_obj.validation_results
            self.success &= self.__session.data.success
            # save step data
            step_data.success = self.__session.data.success
            step_data.data = self.__session.data

        return step_data

    def __run_step_testcase(self, step):
        """run teststep: referenced testcase"""
        step_data = StepData(name=step.name)
        step_variables = step.variables

        if hasattr(step.testcase, "config") and hasattr(step.testcase, "teststeps"):
            testcase_cls = step.testcase
            case_result = (
                testcase_cls()
                .with_session(self.__session)
                .with_case_id(self.__case_id)
                .with_variables(step_variables)
                .run()
            )

        elif isinstance(step.testcase, Text):
            if os.path.isabs(step.testcase):
                ref_testcase_path = step.testcase
            else:
                ref_testcase_path = os.path.join(self.__project_meta.PWD, step.testcase)

            case_result = (
                HttpRunner()
                .with_session(self.__session)
                .with_case_id(self.__case_id)
                .with_variables(step_variables)
                .run_path(ref_testcase_path)
            )

        else:
            raise exceptions.ParamsError(
                f"Invalid teststep referenced testcase: {step.dict()}"
            )

        step_data.data = case_result.get_step_datas()  # list of step data
        step_data.export = case_result.get_export_variables()
        step_data.success = case_result.success
        self.success &= case_result.success

        return step_data

    def __run_step(self, step: TStep):
        """run teststep, teststep maybe a request or referenced testcase"""
        logger.info(f"run step begin: {step.name} >>>>>>")

        if step.request:
            step_data = self.__run_step_request(step)
        elif step.testcase:
            step_data = self.__run_step_testcase(step)
        else:
            raise ParamsError(
                f"teststep is neither a request nor a referenced testcase: {step.dict()}"
            )

        self.__step_datas.append(step_data)
        logger.info(f"run step end: {step.name} <<<<<<\n")
        return step_data.export

    def __parse_config(self, config: TConfig):
        config.variables.update(self.__session_variables)
        config.variables = parse_variables_mapping(
            config.variables, self.__project_meta.functions
        )
        config.name = parse_data(
            config.name, config.variables, self.__project_meta.functions
        )
        config.base_url = parse_data(
            config.base_url, config.variables, self.__project_meta.functions
        )

    def run_testcase(self, testcase: TestCase):
        """run specified testcase

        Examples:
            >>> testcase_obj = TestCase(config=TConfig(...), teststeps=[TStep(...)])
            >>> HttpRunner().with_project_meta(project_meta).run_testcase(testcase_obj)

        """
        self.config = testcase.config
        self.teststeps = testcase.teststeps

        # prepare
        self.__project_meta = self.__project_meta or load_project_meta(self.config.path)
        self.__parse_config(self.config)
        self.__start_at = time.time()
        self.__step_datas: List[StepData] = []
        self.__session = self.__session or HttpSession()
        self.__session_variables = {}

        # run teststeps
        for step in self.teststeps:
            # update with config variables
            step.variables.update(self.config.variables)
            # update with session variables extracted from pre step
            step.variables.update(self.__session_variables)
            # parse variables
            step.variables = parse_variables_mapping(
                step.variables, self.__project_meta.functions
            )
            # run step
            if USE_ALLURE:
                with allure.step(f"step: {step.name}"):
                    extract_mapping = self.__run_step(step)
            else:
                extract_mapping = self.__run_step(step)
            # save extracted variables to session variables
            self.__session_variables.update(extract_mapping)

        self.__duration = time.time() - self.__start_at
        return self

    def run_path(self, path: Text) -> "HttpRunner":
        if not os.path.isfile(path):
            raise exceptions.ParamsError(f"Invalid testcase path: {path}")

        testcase_obj = load_testcase_file(path)
        return self.run_testcase(testcase_obj)

    def run(self) -> "HttpRunner":
        """ run current testcase

        Examples:
            >>> TestCaseRequestWithFunctions().run()

        """
        testcase_obj = TestCase(config=self.config, teststeps=self.teststeps)
        return self.run_testcase(testcase_obj)

    def get_step_datas(self) -> List[StepData]:
        return self.__step_datas

    def get_export_variables(self) -> Dict:
        export_vars_mapping = {}
        for var_name in self.config.export:
            if var_name not in self.__session_variables:
                raise ParamsError(
                    f"failed to export variable {var_name} from session variables {self.__session_variables}"
                )

            export_vars_mapping[var_name] = self.__session_variables[var_name]

        return export_vars_mapping

    def get_summary(self) -> TestCaseSummary:
        """get testcase result summary"""
        start_at_timestamp = self.__start_at
        start_at_iso_format = datetime.utcfromtimestamp(start_at_timestamp).isoformat()
        return TestCaseSummary(
            name=self.config.name,
            success=self.success,
            case_id=self.__case_id,
            time=TestCaseTime(
                start_at=self.__start_at,
                start_at_iso_format=start_at_iso_format,
                duration=self.__duration,
            ),
            in_out=TestCaseInOut(
                vars=self.config.variables, export=self.get_export_variables()
            ),
            log=self.__log_path,
            step_datas=self.__step_datas,
        )

    def test_start(self):
        """main entrance, discovered by pytest"""
        self.__project_meta = self.__project_meta or load_project_meta(self.config.path)
        self.__case_id = self.__case_id or str(uuid.uuid4())
        self.__log_path = self.__log_path or os.path.join(
            self.__project_meta.PWD, "logs", f"{self.__case_id}.run.log"
        )
        log_handler = logger.add(self.__log_path, level="DEBUG")

        # parse config name
        variables = self.config.variables
        variables.update(self.__session_variables)
        self.config.name = parse_data(
            self.config.name, variables, self.__project_meta.functions
        )

        if USE_ALLURE:
            # update allure report meta
            allure.dynamic.title(self.config.name)
            allure.dynamic.description(f"TestCase ID: {self.__case_id}")

        logger.info(
            f"Start to run testcase: {self.config.name}, TestCase ID: {self.__case_id}"
        )

        try:
            return self.run_testcase(
                TestCase(config=self.config, teststeps=self.teststeps)
            )
        finally:
            logger.remove(log_handler)
            logger.info(f"generate testcase log: {self.__log_path}")
