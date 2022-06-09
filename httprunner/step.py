from typing import Union

from httprunner import HttpRunner
from httprunner.models import StepResult, TRequest, TStep, TestCase
from httprunner.step_request import (
    RequestWithOptionalArgs,
    StepRequestExtraction,
    StepRequestValidation,
)
from httprunner.step_sql_request import (
    RunSqlRequest,
    StepSqlRequestExtraction,
    StepSqlRequestValidation,
)
from httprunner.step_testcase import StepRefCase
from httprunner.step_thrift_request import (
    RunThriftRequest,
    StepThriftRequestExtraction,
    StepThriftRequestValidation,
)


class Step(object):
    def __init__(
        self,
        step: Union[
            StepRequestValidation,
            StepRequestExtraction,
            RequestWithOptionalArgs,
            StepRefCase,
            RunSqlRequest,
            StepSqlRequestValidation,
            StepSqlRequestExtraction,
            RunThriftRequest,
            StepThriftRequestValidation,
            StepThriftRequestExtraction,
        ],
    ):
        self.__step = step

    @property
    def request(self) -> TRequest:
        return self.__step.struct().request

    @property
    def testcase(self) -> TestCase:
        return self.__step.struct().testcase

    @property
    def retry_times(self) -> int:
        return self.__step.struct().retry_times

    @property
    def retry_interval(self) -> int:
        return self.__step.struct().retry_interval

    def struct(self) -> TStep:
        return self.__step.struct()

    def name(self) -> str:
        return self.__step.name()

    def type(self) -> str:
        return self.__step.type()

    def run(self, runner: HttpRunner) -> StepResult:
        return self.__step.run(runner)
