from typing import Union

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
