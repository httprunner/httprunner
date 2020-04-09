from enum import Enum
from typing import Dict, List, Any, Text, Union

from pydantic import BaseModel, HttpUrl, Field

Name = Text
Url = Text
BaseUrl = Union[HttpUrl, Text]
Variables = Dict[Text, Any]
Headers = Dict[Text, Text]
Verify = bool
Hook = List[Text]
Export = List[Text]
Validate = List[Dict]
Env = Dict[Text, Any]


class MethodEnum(Text, Enum):
    GET = 'GET'
    POST = 'POST'
    PUT = "PUT"
    DELETE = "DELETE"
    HEAD = "HEAD"
    OPTIONS = "OPTIONS"
    PATCH = "PATCH"
    CONNECT = "CONNECT"
    TRACE = "TRACE"


class TestsConfig(BaseModel):
    name: Name
    verify: Verify = False
    base_url: BaseUrl = ""
    variables: Variables = {}
    setup_hooks: Hook = []
    teardown_hooks: Hook = []
    export: Export = []

    class Config:
        schema_extra = {
            "examples": [
                {
                    "name": "used in testcase/testsuite to configure common fields",
                    "verify": False,
                    "base_url": "https://httpbin.org"
                }
            ]
        }


class Request(BaseModel):
    method: MethodEnum = MethodEnum.GET
    url: Url
    params: Dict[Text, Text] = {}
    headers: Headers = {}
    req_json: Dict = Field({}, alias="json")
    cookies: Dict[Text, Text] = {}
    timeout: int = 120
    allow_redirects: bool = True
    verify: Verify = False
