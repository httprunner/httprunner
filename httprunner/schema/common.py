from enum import Enum
from typing import Dict, List, Any, Tuple

from pydantic import BaseModel, HttpUrl, Field

Name = str
Url = HttpUrl
BaseUrl = str
Variables = Dict[str, Any]
Headers = Dict[str, str]
Verify = bool
Hook = List[str]
Export = List[str]
Extract = Dict[str, str]
Validate = List[Dict]
Env = Dict[str, Any]


class MethodEnum(str, Enum):
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
    params: Dict[str, str] = {}
    headers: Headers = {}
    req_json: Dict = Field({}, alias="json")
    cookies: Dict[str, str] = {}
    timeout: int = 120
    allow_redirects: bool = True
    verify: Verify = False
