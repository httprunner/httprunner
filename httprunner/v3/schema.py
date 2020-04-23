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


class TestCaseTime(BaseModel):
    start_at: float
    duration: float
    start_datetime: Text = ""


class TestCaseInOut(BaseModel):
    vars: VariablesMapping = {}
    out: Export = []


class RequestStat(BaseModel):
    content_size: Text = "N/A"
    response_time_ms: Text = "N/A"
    elapsed_ms: Text = "N/A"


class MetaData(BaseModel):
    name: Text = ""
    data: List[Dict]
    stat: RequestStat
    validators: Dict = {}


class TestCaseSummary(BaseModel):
    name: Text = ""
    success: bool
    status: Text = ""
    attachment: Text = ""
    time: TestCaseTime
    in_out: TestCaseInOut = {}
    log: Text = ""
    step_datas: List[MetaData] = []
    total_response_time: Text = "N/A"


class PlatformInfo(BaseModel):
    httprunner_version: Text
    python_version: Text
    platform: Text


class Stat(BaseModel):
    total: int = 0
    success: int = 0
    fail: int = 0


class TestSuiteSummary(BaseModel):
    success: bool
    stat: Stat
    time: TestCaseTime
    platform: PlatformInfo
    details: List[TestCaseSummary]
