import os
import time
import uuid
from datetime import datetime
from typing import Dict, List, Text

try:
    import allure

    USE_ALLURE = True
except ModuleNotFoundError:
    USE_ALLURE = False

from loguru import logger

from httprunner.client import HttpSession
from httprunner.config import Config
from httprunner.exceptions import ParamsError
from httprunner.loader import load_project_meta
from httprunner.models import (ProjectMeta, StepData, TConfig, TestCaseInOut,
                               TestCaseSummary, TestCaseTime, VariablesMapping)
from httprunner.parser import Parser
from httprunner.utils import merge_variables


class HttpRunner(object):
    config: Config
    teststeps: List[object]     # list of Step

    parser: Parser = None
    session: HttpSession = None
    case_id: Text = ""
    root_dir: Text = ""

    __config: TConfig
    __project_meta: ProjectMeta = None
    __export: List[Text] = []
    __step_datas: List[StepData] = []
    __session_variables: VariablesMapping = {}
    # time
    __start_at: float = 0
    __duration: float = 0
    # log
    __log_path: Text = ""

    def __init(self):
        self.__config = self.config.struct()
        self.__session_variables = {}
        self.__start_at = 0
        self.__duration = 0

        self.__project_meta = self.__project_meta or load_project_meta(
            self.__config.path
        )
        self.case_id = self.case_id or str(uuid.uuid4())
        self.root_dir = self.root_dir or self.__project_meta.RootDir
        self.__log_path = os.path.join(
            self.root_dir, "logs", f"{self.case_id}.run.log"
        )

        self.__step_datas.clear()
        self.session = self.session or HttpSession()
        self.parser = self.parser or Parser(self.__project_meta.functions)

    def with_session(self, session: HttpSession) -> "HttpRunner":
        self.session = session
        return self

    def get_config(self) -> TConfig:
        return self.__config

    def with_case_id(self, case_id: Text) -> "HttpRunner":
        self.case_id = case_id
        return self

    def with_variables(self, variables: VariablesMapping) -> "HttpRunner":
        self.__session_variables = variables
        return self

    def with_export(self, export: List[Text]) -> "HttpRunner":
        self.__export = export
        return self

    def __parse_config(self, param: Dict = None) -> None:
        # parse config variables
        self.__config.variables.update(self.__session_variables)
        if param:
            self.__config.variables.update(param)
        self.__config.variables = self.parser.parse_variables(
            self.__config.variables
        )

        # parse config name
        self.__config.name = self.parser.parse_data(
            self.__config.name, self.__config.variables
        )

        # parse config base url
        self.__config.base_url = self.parser.parse_data(
            self.__config.base_url, self.__config.variables
        )

    def get_export_variables(self) -> Dict:
        # override testcase export vars with step export
        export_var_names = self.__export or self.__config.export
        export_vars_mapping = {}
        for var_name in export_var_names:
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

        summary_success = True
        for step_data in self.__step_datas:
            if not step_data.success:
                summary_success = False
                break

        return TestCaseSummary(
            name=self.__config.name,
            success=summary_success,
            case_id=self.case_id,
            time=TestCaseTime(
                start_at=self.__start_at,
                start_at_iso_format=start_at_iso_format,
                duration=self.__duration,
            ),
            in_out=TestCaseInOut(
                config_vars=self.__config.variables,
                export_vars=self.get_export_variables(),
            ),
            log=self.__log_path,
            step_datas=self.__step_datas,
        )

    def merge_step_variables(self, variables: VariablesMapping) -> VariablesMapping:
        # override variables
        # step variables > extracted variables from previous steps
        variables = merge_variables(variables, self.__session_variables)
        # step variables > testcase config variables
        variables = merge_variables(variables, self.__config.variables)

        # parse variables
        return self.parser.parse_variables(variables)

    def __run_step(self, step) -> Dict:
        """run teststep, step maybe any kind that implements IStep interface

        Args:
            step (Step): teststep

        """
        logger.info(f"run step begin: {step.name()} >>>>>>")

        # run step
        if USE_ALLURE:
            with allure.step(f"step: {step.name()}"):
                step_result = step.run(self)
        else:
            step_result = step.run(self)

        # save extracted variables to session variables
        self.__session_variables.update(step_result.export_vars)
        # update testcase summary
        self.__step_datas.append(step_result)

        logger.info(f"run step end: {step.name()} <<<<<<\n")

    def test_start(self, param: Dict = None):
        """main entrance, discovered by pytest"""
        self.__init()
        log_handler = logger.add(self.__log_path, level="DEBUG")

        self.__parse_config(param)

        if USE_ALLURE:
            # update allure report meta
            allure.dynamic.title(self.__config.name)
            allure.dynamic.description(f"TestCase ID: {self.case_id}")

        logger.info(
            f"Start to run testcase: {self.__config.name}, TestCase ID: {self.case_id}"
        )

        self.__start_at = time.time()
        try:
            # run step in sequential order
            for step in self.teststeps:
                self.__run_step(step)
        finally:
            logger.remove(log_handler)
            logger.info(f"generate testcase log: {self.__log_path}")

        self.__duration = time.time() - self.__start_at
        return self
