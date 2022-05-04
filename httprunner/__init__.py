__version__ = "v4.0.0"
__description__ = "One-stop solution for HTTP(S) testing."

from httprunner.config import Config
import platform
from httprunner.parser import parse_parameters as Parameters
from httprunner.runner import HttpRunner
from httprunner.step import Step
from httprunner.step_request import RunRequest
from httprunner.step_testcase import RunTestCase
from httprunner.step_sql_request import (
    RunSqlRequest,
    StepSqlRequestValidation,
    StepSqlRequestExtraction,
)

__all__ = [
    "__version__",
    "__description__",
    "HttpRunner",
    "Config",
    "Step",
    "RunRequest",
    "RunSqlRequest",
    "StepSqlRequestValidation",
    "StepSqlRequestExtraction",
    "RunTestCase",
    "Parameters",
]
if platform.system() != "Windows":
    from httprunner.step_thrift_request import (
        RunThriftRequest,
        StepThriftRequestValidation,
        StepThriftRequestExtraction,
    )

    __all__.extend(
        [
            "RunThriftRequest",
            "StepThriftRequestValidation",
            "StepThriftRequestExtraction",
        ]
    )
