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
    """requests.Request model"""
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
    request: Request = None
    testcase: Union[Text, Callable] = None
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
    project_meta: ProjectMeta
    testcases: List[TestCase]


class TestCaseTime(BaseModel):
    start_at: float = 0
    start_at_iso_format: Text = ""
    duration: float = 0


class TestCaseInOut(BaseModel):
    vars: VariablesMapping = {}
    out: Export = []


class RequestStat(BaseModel):
    content_size: float = 0
    response_time_ms: float = 0
    elapsed_ms: float = 0


class RequestData(BaseModel):
    method: MethodEnum = MethodEnum.GET
    url: Url
    headers: Headers = {}
    # TODO: add cookies
    body: Union[Text, Dict] = {}


class ResponseData(BaseModel):
    status_code: int
    cookies: Dict
    encoding: Text
    headers: Dict
    content_type: Text
    body: Union[Text, bytes, Dict]


class ReqRespData(BaseModel):
    request: RequestData
    response: ResponseData


class SessionData(BaseModel):
    """request session data, including request, response, validators and stat data"""
    success: bool = False
    # in most cases, req_resps only contains one request & response
    # while when 30X redirect occurs, req_resps will contain multiple request & response
    req_resps: List[ReqRespData] = []
    stat: RequestStat = RequestStat()
    validators: Dict = {}


class StepData(BaseModel):
    """teststep data, each step maybe corresponding to one request or one testcase"""
    success: bool = False
    name: Text = ""     # teststep name
    data: Union[SessionData, List[SessionData]] = None
    export: Dict = {}


class TestCaseSummary(BaseModel):
    name: Text = ""
    success: bool
    status: Text = ""
    attachment: Text = ""
    time: TestCaseTime
    in_out: TestCaseInOut = {}
    log: Text = ""
    step_datas: List[StepData] = []


class PlatformInfo(BaseModel):
    httprunner_version: Text
    python_version: Text
    platform: Text


class Stat(BaseModel):
    total: int = 0
    success: int = 0
    fail: int = 0


class TestSuiteSummary(BaseModel):
    success: bool = False
    stat: Stat = Stat()
    time: TestCaseTime = TestCaseTime()
    platform: PlatformInfo
    testcases: List[TestCaseSummary]
