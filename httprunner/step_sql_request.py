# -*- coding: utf-8 -*-
import sys
import time
from typing import Text

from loguru import logger

from httprunner import utils
from httprunner.exceptions import SqlMethodNotSupport, ValidationFailure
from httprunner.models import IStep, SqlMethodEnum, StepResult, TSqlRequest, TStep
from httprunner.response import SqlResponseObject
from httprunner.runner import ALLURE, HttpRunner
from httprunner.step_request import (
    StepRequestExtraction,
    StepRequestValidation,
    call_hooks,
)

try:
    import pymysql
    import sqlalchemy

    SQL_READY = True
except ModuleNotFoundError:
    SQL_READY = False


def ensure_sql_ready():
    if SQL_READY:
        return

    msg = """
    uploader extension dependencies uninstalled, install first and try again.
    install with pip:
    $ pip install sqlalchemy pymysql

    or you can install httprunner with optional upload dependencies:
    $ pip install "httprunner[sql]"
    """
    logger.error(msg)
    sys.exit(1)


def run_step_sql_request(runner: HttpRunner, step: TStep) -> StepResult:
    """run teststep:sql request"""
    start_time = time.time()

    step_result = StepResult(
        name=step.name,
        step_type="sql",
        success=False,
    )
    step_variables = runner.merge_step_variables(step.variables)
    # parse
    request_dict = step.sql_request.dict()
    parsed_request_dict = runner.parser.parse_data(request_dict, step_variables)
    config = runner.get_config()
    parsed_request_dict["db_config"]["psm"] = (
        parsed_request_dict["db_config"]["psm"] or config.db.psm
    )
    parsed_request_dict["db_config"]["user"] = (
        parsed_request_dict["db_config"]["user"] or config.db.user
    )
    parsed_request_dict["db_config"]["password"] = (
        parsed_request_dict["db_config"]["password"] or config.db.password
    )
    parsed_request_dict["db_config"]["ip"] = (
        parsed_request_dict["db_config"]["ip"] or config.db.ip
    )
    parsed_request_dict["db_config"]["port"] = (
        parsed_request_dict["db_config"]["port"] or config.db.port
    )
    parsed_request_dict["db_config"]["database"] = (
        parsed_request_dict["db_config"]["database"] or config.db.database
    )

    if not runner.db_engine:
        ensure_sql_ready()
        from httprunner.database.engine import DBEngine

        runner.db_engine = DBEngine(
            f'mysql+pymysql://{parsed_request_dict["db_config"]["user"]}:'
            f'{parsed_request_dict["db_config"]["password"]}@{parsed_request_dict["db_config"]["ip"]}:'
            f'{parsed_request_dict["db_config"]["port"]}/{parsed_request_dict["db_config"]["database"]}'
            f"?charset=utf8mb4"
        )

    # parsed_request_dict["headers"].setdefault(
    #     "HRUN-Request-ID",
    #     f"HRUN-{self.__case_id}-{str(int(time.time() * 1000))[-6:]}",
    # )

    # setup hooks
    if step.setup_hooks:
        call_hooks(runner, step.setup_hooks, step_variables, "setup request")

    # log request
    sql_request_print = "====== sql request details ======\n"
    sql_request_print += f"sql: {step.sql_request.sql}\n"
    for k, v in parsed_request_dict.items():
        v = utils.omit_long_data(v)
        sql_request_print += f"{k}: {repr(v)}\n"

    sql_request_print += "\n"

    if ALLURE is not None:
        ALLURE.attach(
            sql_request_print,
            name="sql request details",
            attachment_type=ALLURE.attachment_type.TEXT,
        )
    logger.info(f"Executing SQL: {parsed_request_dict['sql']}")
    if step.sql_request.method == SqlMethodEnum.FETCHONE:
        sql_resp = runner.db_engine.fetchone(parsed_request_dict["sql"])
    elif step.sql_request.method == SqlMethodEnum.INSERT:
        sql_resp = runner.db_engine.insert(parsed_request_dict["sql"])
    elif step.sql_request.method == SqlMethodEnum.FETCHMANY:
        sql_resp = runner.db_engine.fetchmany(
            parsed_request_dict["sql"], parsed_request_dict["size"]
        )
    elif step.sql_request.method == SqlMethodEnum.FETCHALL:
        sql_resp = runner.db_engine.fetchall(parsed_request_dict["sql"])
    elif step.sql_request.method == SqlMethodEnum.UPDATE:
        sql_resp = runner.db_engine.update(parsed_request_dict["sql"])
    elif step.sql_request.method == SqlMethodEnum.DELETE:
        sql_resp = runner.db_engine.delete(parsed_request_dict["sql"])
    else:
        raise SqlMethodNotSupport(
            f"step.sql_request.method {parsed_request_dict['method']} not support"
        )

    # log response
    sql_response_print = "====== sql response details ======\n"
    if isinstance(sql_resp, dict):
        for k, v in sql_resp.items():
            v = utils.omit_long_data(v)
            sql_response_print += f"{k}: {repr(v)}\n"
    elif isinstance(sql_resp, list):
        sql_response_print += f"count: {len(sql_resp)}\n"
        sql_response_print += "-" * 34 + "\n"
        for el in sql_resp:
            for k, v in el.items():
                v = utils.omit_long_data(v)
                sql_response_print += f"{k}: {repr(v)}\n"
            sql_response_print += "-" * 34 + "\n"
    elif sql_resp is None:
        sql_response_print += "None\n"
    if ALLURE is not None:
        ALLURE.attach(
            sql_response_print,
            name="sql response details",
            attachment_type=ALLURE.attachment_type.TEXT,
        )

    resp_obj = SqlResponseObject(sql_resp, parser=runner.parser)
    step_variables["sql_response"] = resp_obj

    # teardown hooks
    if step.teardown_hooks:
        call_hooks(runner, step.teardown_hooks, step_variables, "teardown request")

    def log_sql_req_resp_details():
        err_msg = "\n{} SQL DETAILED REQUEST & RESPONSE {}\n".format("*" * 32, "*" * 32)
        err_msg += sql_request_print + sql_response_print
        logger.error(err_msg)

    # extract
    extractors = step.extract
    extract_mapping = resp_obj.extract(extractors)
    step_result.export_vars = extract_mapping

    variables_mapping = step_variables
    variables_mapping.update(extract_mapping)

    # validate
    validators = step.validators
    try:
        resp_obj.validate(validators, variables_mapping)
        step_result.success = True
    except ValidationFailure:
        log_sql_req_resp_details()
        raise
    finally:
        session_data = runner.session.data
        session_data.success = step_result.success
        session_data.validators = resp_obj.validation_results

        # save step data
        step_result.data = session_data
        step_result.elapsed = time.time() - start_time
    return step_result


class StepSqlRequestValidation(StepRequestValidation):
    def __init__(self, step: TStep):
        self.__step = step
        super().__init__(step)

    def run(self, runner: HttpRunner):
        return run_step_sql_request(runner, self.__step)


class StepSqlRequestExtraction(StepRequestExtraction):
    def __init__(self, step: TStep):
        self.__step = step
        super().__init__(step)

    def run(self, runner: HttpRunner):
        return run_step_sql_request(runner, self.__step)

    def validate(self) -> StepSqlRequestValidation:
        return StepSqlRequestValidation(self.__step)


class RunSqlRequest(IStep):
    def __init__(self, name: Text):
        self.__step = TStep(name=name)
        self.__step.sql_request = TSqlRequest()

    def with_variables(self, **variables) -> "RunSqlRequest":
        self.__step.variables.update(variables)
        return self

    def with_db_config(
        self, user=None, password=None, ip=None, port=None, database=None, psm=None
    ):
        if user:
            self.__step.sql_request.db_config.user = user
        if password:
            self.__step.sql_request.db_config.password = password
        if ip:
            self.__step.sql_request.db_config.ip = ip
        if port:
            self.__step.sql_request.db_config.port = port
        if database:
            self.__step.sql_request.db_config.database = database
        if psm:
            self.__step.sql_request.db_config.psm = psm
        return self

    def fetchone(self, sql) -> "RunSqlRequest":
        self.__step.sql_request.method = SqlMethodEnum.FETCHONE
        self.__step.sql_request.sql = sql
        return self

    def fetchmany(self, sql, size) -> "RunSqlRequest":
        self.__step.sql_request.method = SqlMethodEnum.FETCHMANY
        self.__step.sql_request.sql = sql
        self.__step.sql_request.size = size
        return self

    def fetchall(self, sql) -> "RunSqlRequest":
        self.__step.sql_request.method = SqlMethodEnum.FETCHALL
        self.__step.sql_request.sql = sql
        return self

    def update(self, sql) -> "RunSqlRequest":
        self.__step.sql_request.method = SqlMethodEnum.UPDATE
        self.__step.sql_request.sql = sql
        return self

    def delete(self, sql) -> "RunSqlRequest":
        self.__step.sql_request.method = SqlMethodEnum.DELETE
        self.__step.sql_request.sql = sql
        return self

    def insert(self, sql) -> "RunSqlRequest":
        self.__step.sql_request.method = SqlMethodEnum.INSERT
        self.__step.sql_request.sql = sql
        return self

    def with_retry(self, retry_times, retry_interval) -> "RunSqlRequest":
        self.__step.retry_times = retry_times
        self.__step.retry_interval = retry_interval
        return self

    def teardown_hook(
        self, hook: Text, assign_var_name: Text = None
    ) -> "RunSqlRequest":
        if assign_var_name:
            self.__step.teardown_hooks.append({assign_var_name: hook})
        else:
            self.__step.teardown_hooks.append(hook)

        return self

    def setup_hook(self, hook: Text, assign_var_name: Text = None) -> "RunSqlRequest":
        if assign_var_name:
            self.__step.setup_hooks.append({assign_var_name: hook})
        else:
            self.__step.setup_hooks.append(hook)

        return self

    def struct(self) -> TStep:
        return self.__step

    def name(self) -> Text:
        return self.__step.name

    def type(self) -> Text:
        return f"sql-request-{self.__step.sql_request.sql}"

    def run(self, runner) -> StepResult:
        return run_step_sql_request(runner, self.__step)

    def extract(self) -> StepSqlRequestExtraction:
        return StepSqlRequestExtraction(self.__step)

    def validate(self) -> StepSqlRequestValidation:
        return StepSqlRequestValidation(self.__step)

    def with_jmespath(
        self, jmes_path: Text, var_name: Text
    ) -> "StepSqlRequestExtraction":
        self.__step.extract[var_name] = jmes_path
        return StepSqlRequestExtraction(self.__step)
