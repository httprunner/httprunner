import os
from enum import Enum
from typing import Any, Callable, Dict, List, Text, Union

from pydantic import BaseModel, Field, HttpUrl

Name = Text
Url = Text
BaseUrl = Union[HttpUrl, Text]
VariablesMapping = Dict[Text, Any]
FunctionsMapping = Dict[Text, Callable]
Headers = Dict[Text, Text]
Cookies = Dict[Text, Text]
Verify = bool
Hooks = List[Union[Text, Dict[Text, Text]]]
Export = List[Text]
Validators = List[Dict]
Env = Dict[Text, Any]


class MethodEnum(Text, Enum):
    GET = "GET"
    POST = "POST"
    PUT = "PUT"
    DELETE = "DELETE"
    HEAD = "HEAD"
    OPTIONS = "OPTIONS"
    PATCH = "PATCH"


class ProtoType(Enum):
    Binary = 1
    CyBinary = 2
    Compact = 3
    Json = 4


class TransType(Enum):
    Buffered = 1
    CyBuffered = 2
    Framed = 3
    CyFramed = 4


# configs for thrift rpc
class TConfigThrift(BaseModel):
    psm: Text = None
    env: Text = None
    cluster: Text = None
    target: Text = None
    include_dirs: List[Text] = None
    thrift_client: Any = None
    timeout: int = 10
    idl_path: Text = None
    method: Text = None
    ip: Text = "127.0.0.1"
    port: int = 9000
    service_name: Text = None
    proto_type: ProtoType = ProtoType.Binary
    trans_type: TransType = TransType.Buffered


# configs for db
class TConfigDB(BaseModel):
    psm: Text = None
    user: Text = None
    password: Text = None
    ip: Text = None
    port: int = 3306
    database: Text = None


class TransportEnum(Text, Enum):
    BUFFERED = "buffered"
    FRAMED = "framed"


class TThriftRequest(BaseModel):
    """rpc request model"""

    method: Text = ""
    params: Dict = {}
    thrift_client: Any = None
    idl_path: Text = ""  # idl local path
    timeout: int = 10  # sec
    transport: TransportEnum = TransportEnum.BUFFERED
    include_dirs: List[Union[Text, None]] = []  # param of thriftpy2.load
    target: Text = ""  # tcp://{ip}:{port} or sd://psm?cluster=xx&env=xx
    env: Text = "prod"
    cluster: Text = "default"
    psm: Text = ""
    service_name: Text = None
    ip: Text = None
    port: int = None
    proto_type: ProtoType = None
    trans_type: TransType = None


class SqlMethodEnum(Text, Enum):
    FETCHONE = "FETCHONE"
    FETCHMANY = "FETCHMANY"
    FETCHALL = "FETCHALL"
    INSERT = "INSERT"
    UPDATE = "UPDATE"
    DELETE = "DELETE"


class TSqlRequest(BaseModel):
    """sql request model"""

    db_config: TConfigDB = TConfigDB()
    method: SqlMethodEnum = None
    sql: Text = None
    size: int = 0  # limit nums of sql result


class TConfig(BaseModel):
    name: Name
    verify: Verify = False
    base_url: BaseUrl = ""
    # Text: prepare variables in debugtalk.py, ${gen_variables()}
    variables: Union[VariablesMapping, Text] = {}
    parameters: Union[VariablesMapping, Text] = {}
    # setup_hooks: Hooks = []
    # teardown_hooks: Hooks = []
    export: Export = []
    path: Text = None
    # configs for other protocols
    thrift: TConfigThrift = None
    db: TConfigDB = TConfigDB()


class TRequest(BaseModel):
    """requests.Request model"""

    method: MethodEnum
    url: Url
    params: Dict[Text, Text] = {}
    headers: Headers = {}
    req_json: Union[Dict, List, Text] = Field(None, alias="json")
    data: Union[Text, Dict[Text, Any]] = None
    cookies: Cookies = {}
    timeout: float = 120
    allow_redirects: bool = True
    verify: Verify = False
    upload: Dict = {}  # used for upload files


class TStep(BaseModel):
    name: Name
    request: Union[TRequest, None] = None
    testcase: Union[Text, Callable, None] = None
    variables: VariablesMapping = {}
    setup_hooks: Hooks = []
    teardown_hooks: Hooks = []
    # used to extract request's response field
    extract: VariablesMapping = {}
    # used to export session variables from referenced testcase
    export: Export = []
    validators: Validators = Field([], alias="validate")
    validate_script: List[Text] = []
    retry_times: int = 0
    retry_interval: int = 0  # sec
    thrift_request: Union[TThriftRequest, None] = None
    sql_request: Union[TSqlRequest, None] = None


class TestCase(BaseModel):
    config: TConfig
    teststeps: List[TStep]


class ProjectMeta(BaseModel):
    debugtalk_py: Text = ""  # debugtalk.py file content
    debugtalk_path: Text = ""  # debugtalk.py file path
    dot_env_path: Text = ""  # .env file path
    functions: FunctionsMapping = {}  # functions defined in debugtalk.py
    env: Env = {}
    RootDir: Text = (
        os.getcwd()
    )  # project root directory (ensure absolute), the path debugtalk.py located


class TestsMapping(BaseModel):
    project_meta: ProjectMeta
    testcases: List[TestCase]


class TestCaseTime(BaseModel):
    start_at: float = 0
    start_at_iso_format: Text = ""
    duration: float = 0


class TestCaseInOut(BaseModel):
    config_vars: VariablesMapping = {}
    export_vars: Dict = {}


class RequestStat(BaseModel):
    content_size: float = 0
    response_time_ms: float = 0
    elapsed_ms: float = 0


class AddressData(BaseModel):
    client_ip: Text = "N/A"
    client_port: int = 0
    server_ip: Text = "N/A"
    server_port: int = 0


class RequestData(BaseModel):
    method: MethodEnum = MethodEnum.GET
    url: Url
    headers: Headers = {}
    cookies: Cookies = {}
    body: Union[Text, bytes, List, Dict, None] = {}


class ResponseData(BaseModel):
    status_code: int
    headers: Dict
    cookies: Cookies
    encoding: Union[Text, None] = None
    content_type: Text
    body: Union[Text, bytes, List, Dict, None]


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
    address: AddressData = AddressData()
    validators: Dict = {}


class StepResult(BaseModel):
    """teststep data, each step maybe corresponding to one request or one testcase"""

    name: Text = ""  # teststep name
    step_type: Text = ""  # teststep type, request or testcase
    success: bool = False
    data: Union[SessionData, List["StepResult"]] = None
    elapsed: float = 0.0  # teststep elapsed time
    content_size: float = 0  # response content size
    export_vars: VariablesMapping = {}
    attachment: Text = ""  # teststep attachment


StepResult.update_forward_refs()


class IStep(object):
    def name(self) -> str:
        raise NotImplementedError

    def type(self) -> str:
        raise NotImplementedError

    def struct(self) -> TStep:
        raise NotImplementedError

    def run(self, runner) -> StepResult:
        # runner: HttpRunner
        raise NotImplementedError


class TestCaseSummary(BaseModel):
    name: Text
    success: bool
    case_id: Text
    time: TestCaseTime
    in_out: TestCaseInOut = {}
    log: Text = ""
    step_results: List[StepResult] = []


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
