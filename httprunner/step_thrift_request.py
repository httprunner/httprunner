# -*- coding: utf-8 -*-
import platform
import sys
import time
from typing import Text, Union

from loguru import logger

from httprunner import utils
from httprunner.exceptions import ValidationFailure
from httprunner.models import (
    IStep,
    ProtoType,
    StepResult,
    TransType,
    TStep,
    TThriftRequest,
)
from httprunner.response import ThriftResponseObject
from httprunner.runner import ALLURE, HttpRunner
from httprunner.step_request import (
    StepRequestExtraction,
    StepRequestValidation,
    call_hooks,
)

try:
    import thriftpy2

    from thrift.Thrift import TType

    THRIFT_READY = True
except ModuleNotFoundError:
    THRIFT_READY = False


def ensure_thrift_ready():
    assert platform.system() != "Windows", "Sorry,thrift not support Windows for now"
    if THRIFT_READY:
        return

    msg = """
    uploader extension dependencies uninstalled, install first and try again.
    install with pip:
    $ pip install cython thriftpy2 thrift

    or you can install httprunner with optional upload dependencies:
    $ pip install "httprunner[thrift]"
    """
    logger.error(msg)
    sys.exit(1)


def run_step_thrift_request(runner: HttpRunner, step: TStep) -> StepResult:
    """run teststep:thrift request"""
    start_time = time.time()

    step_result = StepResult(
        name=step.name,
        step_type="thrift",
        success=False,
    )
    step_variables = runner.merge_step_variables(step.variables)
    # parse
    request_dict = step.thrift_request.dict()
    parsed_request_dict = runner.parser.parse_data(request_dict, step_variables)
    config = runner.get_config()
    parsed_request_dict["psm"] = parsed_request_dict["psm"] or config.thrift.psm
    parsed_request_dict["env"] = parsed_request_dict["env"] or config.thrift.env
    parsed_request_dict["cluster"] = (
        parsed_request_dict["cluster"] or config.thrift.cluster
    )
    parsed_request_dict["idl_path"] = (
        parsed_request_dict["idl_path"] or config.thrift.idl_path
    )
    parsed_request_dict["include_dirs"] = (
        parsed_request_dict["include_dirs"] or config.thrift.include_dirs
    )
    parsed_request_dict["method"] = (
        parsed_request_dict["method"] or config.thrift.method
    )
    parsed_request_dict["service_name"] = (
        parsed_request_dict["service_name"] or config.thrift.service_name
    )
    parsed_request_dict["ip"] = parsed_request_dict["ip"] or config.thrift.ip
    parsed_request_dict["port"] = parsed_request_dict["port"] or config.thrift.port
    parsed_request_dict["proto_type"] = (
        parsed_request_dict["proto_type"] or config.thrift.proto_type
    )
    parsed_request_dict["trans_port"] = (
        parsed_request_dict["trans_type"] or config.thrift.trans_type
    )
    parsed_request_dict["timeout"] = (
        parsed_request_dict["timeout"] or config.thrift.timeout
    )
    parsed_request_dict["thrift_client"] = parsed_request_dict["thrift_client"]

    # parsed_request_dict["headers"].setdefault(
    #     "HRUN-Request-ID",
    #     f"HRUN-{self.__case_id}-{str(int(time.time() * 1000))[-6:]}",
    # )
    step_variables["thrift_request"] = parsed_request_dict

    psm = parsed_request_dict["psm"]
    if not runner.thrift_client:
        runner.thrift_client = parsed_request_dict["thrift_client"]
    if not runner.thrift_client:
        ensure_thrift_ready()
        from httprunner.thrift.thrift_client import ThriftClient

        runner.thrift_client = ThriftClient(
            thrift_file=parsed_request_dict["idl_path"],
            service_name=parsed_request_dict["service_name"],
            ip=parsed_request_dict["ip"],
            port=parsed_request_dict["port"],
            include_dirs=parsed_request_dict["include_dirs"],
            timeout=parsed_request_dict["timeout"],
            proto_type=parsed_request_dict["proto_type"],
            trans_type=parsed_request_dict["trans_port"],
        )

    # setup hooks
    if step.setup_hooks:
        call_hooks(runner, step.setup_hooks, step_variables, "setup request")

    # log request
    thrift_request_print = "====== thrift request details ======\n"
    thrift_request_print += f"psm: {psm}\n"
    for k, v in parsed_request_dict.items():
        v = utils.omit_long_data(v)
        thrift_request_print += f"{k}: {repr(v)}\n"
    thrift_request_print += "\n"
    if ALLURE is not None:
        ALLURE.attach(
            thrift_request_print,
            name="thrift request details",
            attachment_type=ALLURE.attachment_type.TEXT,
        )

    # thrift request
    resp = runner.thrift_client.send_request(
        parsed_request_dict["params"], parsed_request_dict["method"]
    )
    resp_obj = ThriftResponseObject(resp, parser=runner.parser)
    step_variables["thrift_response"] = resp_obj

    # log response
    thrift_response_print = "====== thrift response details ======\n"
    for k, v in resp.items():
        v = utils.omit_long_data(v)
        thrift_response_print += f"{k}: {repr(v)}\n"
    if ALLURE is not None:
        ALLURE.attach(
            thrift_request_print,
            name="thrift response details",
            attachment_type=ALLURE.attachment_type.TEXT,
        )

    # teardown hooks
    if step.teardown_hooks:
        call_hooks(runner, step.teardown_hooks, step_variables, "teardown request")

    def log_thrift_req_resp_details():
        err_msg = "\n{} THRIFT DETAILED REQUEST & RESPONSE {}\n".format(
            "*" * 32, "*" * 32
        )
        err_msg += thrift_request_print + thrift_response_print
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
        log_thrift_req_resp_details()
        raise
    finally:
        session_data = runner.session.data
        session_data.success = step_result.success
        session_data.validators = resp_obj.validation_results

        # save step data
        step_result.data = session_data
        step_result.elapsed = time.time() - start_time
    return step_result


class StepThriftRequestValidation(StepRequestValidation):
    def __init__(self, step: TStep):
        self.__step = step
        super().__init__(step)

    def run(self, runner: HttpRunner):
        return run_step_thrift_request(runner, self.__step)


class StepThriftRequestExtraction(StepRequestExtraction):
    def __init__(self, step: TStep):
        self.__step = step
        super().__init__(step)

    def run(self, runner: HttpRunner):
        return run_step_thrift_request(runner, self.__step)

    def validate(self) -> StepThriftRequestValidation:
        return StepThriftRequestValidation(self.__step)


class RunThriftRequest(IStep):
    def __init__(self, name: Text):
        self.__step = TStep(name=name)
        self.__step.thrift_request = TThriftRequest()

    def with_variables(self, **variables) -> "RunThriftRequest":
        self.__step.variables.update(variables)
        return self

    def with_retry(self, retry_times, retry_interval) -> "RunThriftRequest":
        self.__step.retry_times = retry_times
        self.__step.retry_interval = retry_interval
        return self

    def teardown_hook(
        self, hook: Text, assign_var_name: Text = None
    ) -> "RunThriftRequest":
        if assign_var_name:
            self.__step.teardown_hooks.append({assign_var_name: hook})
        else:
            self.__step.teardown_hooks.append(hook)

        return self

    def setup_hook(
        self, hook: Text, assign_var_name: Text = None
    ) -> "RunThriftRequest":
        if assign_var_name:
            self.__step.setup_hooks.append({assign_var_name: hook})
        else:
            self.__step.setup_hooks.append(hook)

        return self

    def with_params(self, **params) -> "RunThriftRequest":
        self.__step.thrift_request.params.update(params)
        return self

    def with_method(self, method) -> "RunThriftRequest":
        self.__step.thrift_request.method = method
        return self

    def with_idl_path(self, idl_path, idl_root_path) -> "RunThriftRequest":
        self.__step.thrift_request.idl_path = idl_path
        self.__step.thrift_request.include_dirs = [idl_root_path]
        return self

    def with_thrift_client(
        self, thrift_client: Union["ThriftClient", str]
    ) -> "RunThriftRequest":
        self.__step.thrift_request.thrift_client = thrift_client
        return self

    def with_ip(self, ip: str) -> "RunThriftRequest":
        self.__step.thrift_request.ip = ip
        return self

    def with_port(self, port: int) -> "RunThriftRequest":
        self.__step.thrift_request.port = port
        return self

    def with_proto_type(self, proto_type: ProtoType) -> "RunThriftRequest":
        self.__step.thrift_request.proto_type = proto_type
        return self

    def with_trans_type(self, trans_type: TransType) -> "RunThriftRequest":
        self.__step.thrift_request.proto_type = trans_type
        return self

    def struct(self) -> TStep:
        return self.__step

    def name(self) -> Text:
        return self.__step.name

    def type(self) -> Text:
        return f"thrift-request-{self.__step.thrift_request.psm}-{self.__step.thrift_request.method}"

    def run(self, runner) -> StepResult:
        return run_step_thrift_request(runner, self.__step)

    def extract(self) -> StepThriftRequestExtraction:
        return StepThriftRequestExtraction(self.__step)

    def validate(self) -> StepThriftRequestValidation:
        return StepThriftRequestValidation(self.__step)

    def with_jmespath(
        self, jmes_path: Text, var_name: Text
    ) -> "StepThriftRequestExtraction":
        self.__step.extract[var_name] = jmes_path
        return StepThriftRequestExtraction(self.__step)
