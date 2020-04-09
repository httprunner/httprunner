from typing import List

from pydantic import BaseModel

from httprunner.schema import common, TestCase


class TestSuite(BaseModel):
    config: common.TestsConfig
    testcases: List[TestCase]
