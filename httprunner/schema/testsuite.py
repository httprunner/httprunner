from typing import List, Text

from pydantic import BaseModel

from httprunner.schema import common


class TestCase(BaseModel):
    name: common.Name
    testcase: Text
    weight: int = 1
    variables: common.Variables = {}


class TestSuite(BaseModel):
    config: common.TestsConfig
    testcases: List[TestCase]
