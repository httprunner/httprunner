__version__ = "3.1.0"
__description__ = "One-stop solution for HTTP(S) testing."

# import firstly for monkey patch if needed
from httprunner.ext.locust import main_locusts
from httprunner.runner import HttpRunner
from httprunner.testcase import Config, Step, RunRequest, RunTestCase

__all__ = [
    "__version__",
    "__description__",
    "HttpRunner",
    "Config",
    "Step",
    "RunRequest",
    "RunTestCase",
]
