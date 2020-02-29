from typing import Dict, List, Text

from pydantic import BaseModel, Field

from httprunner.schema import common


class ProjectMeta(BaseModel):
    debugtalk_py: Text = ""
    variables: common.Variables = {}
    env: common.Env = {}


class TestStep(BaseModel):
    name: common.Name
    request: common.Request
    extract: Dict[str, str] = {}
    validation: common.Validate = Field([], alias="validate")


class TestCase(BaseModel):
    config: common.TestsConfig
    teststeps: List[TestStep]

    class Config:
        schema_extra = {
            "examples": [
                {
                    "config": {
                        "name": "testcase name"
                    },
                    "teststeps": [
                        {
                            "name": "api 1",
                            "api": "/path/to/api1"
                        },
                        {
                            "name": "api 2",
                            "api": "/path/to/api2"
                        }
                    ]
                },
                {
                    "config": {
                        "name": "demo testcase",
                        "variables": {
                            "device_sn": "ABC",
                            "username": "${ENV(USERNAME)}",
                            "password": "${ENV(PASSWORD)}"
                        },
                        "base_url": "http://127.0.0.1:5000"
                    },
                    "teststeps": [
                        {
                            "name": "demo step 1",
                            "api": "path/to/api1.yml",
                            "variables": {
                                "user_agent": "iOS/10.3",
                                "device_sn": "$device_sn"
                            },
                            "extract": [
                                {
                                    "token": "content.token"
                                }
                            ],
                            "validate": [
                                {
                                    "eq": ["status_code", 200]
                                }
                            ]
                        },
                        {
                            "name": "demo step 2",
                            "api": "path/to/api2.yml",
                            "variables": {
                                "token": "$token"
                            }
                        }
                    ]
                }
            ]
        }


TestCases = List[TestCase]
