from enum import Enum
from typing import Any
from typing import Dict, Text, Union, Callable
from typing import List

from pydantic import BaseModel, Field
from pydantic import HttpUrl

Name = Text
Url = Text
BaseUrl = Union[HttpUrl, Text]
VariablesMapping = Dict[Text, Any]
FunctionsMapping = Dict[Text, Callable]
Headers = Dict[Text, Text]
Verify = bool
Hook = List[Text]
Export = List[Text]
Validators = List[Dict]
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
    variables: VariablesMapping = {}
    functions: FunctionsMapping = {}
    setup_hooks: Hook = []
    teardown_hooks: Hook = []
    export: Export = []


class Request(BaseModel):
    method: MethodEnum = MethodEnum.GET
    url: Url
    params: Dict[Text, Text] = {}
    headers: Headers = {}
    req_json: Dict = Field({}, alias="json")
    data: Union[Text, Dict[Text, Any]] = ""
    cookies: Dict[Text, Text] = {}
    timeout: int = 120
    allow_redirects: bool = True
    verify: Verify = False


class TestStep(BaseModel):
    name: Name
    request: Request
    variables: VariablesMapping = {}
    extract: Dict[Text, Text] = {}
    validators: Validators = Field([], alias="validate")


class TestCase(BaseModel):
    config: TestsConfig
    teststeps: List[TestStep]


class ProjectMeta(BaseModel):
    debugtalk_py: Text = ""
    variables: VariablesMapping = {}
    functions: FunctionsMapping = {}
    env: Env = {}
    PWD: Text
    test_path: Text


class TestsMapping(BaseModel):
    project_mapping: ProjectMeta    # TODO: rename to project_meta
    testcases: List[TestCase]
