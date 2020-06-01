import inspect
from typing import Text, Any, Dict, Callable

from httprunner.schema import (
    TConfig,
    TStep,
    TRequest,
    MethodEnum,
    TestCase,
)


class Config(object):
    def __init__(self, name: Text):
        self.__name = name
        self.__variables = {}
        self.__base_url = ""
        self.__verify = False

        caller_frame = inspect.stack()[1]
        self.__path = caller_frame.filename

    @property
    def name(self):
        return self.__name

    @property
    def path(self):
        return self.__path

    def variables(self, **variables) -> "Config":
        self.__variables.update(variables)
        return self

    def base_url(self, base_url: Text) -> "Config":
        self.__base_url = base_url
        return self

    def verify(self, verify: bool) -> "Config":
        self.__verify = verify
        return self

    def perform(self) -> TConfig:
        return TConfig(
            name=self.__name,
            base_url=self.__base_url,
            verify=self.__verify,
            variables=self.__variables,
            path=self.__path,
        )


class RequestWithOptionalArgs(object):
    def __init__(self, method: MethodEnum, url: Text):
        self.__method = method
        self.__url = url
        self.__params = {}
        self.__headers = {}
        self.__cookies = {}
        self.__data = ""
        self.__timeout = 120
        self.__allow_redirects = True
        self.__verify = False

    def with_params(self, **params) -> "RequestWithOptionalArgs":
        self.__params.update(params)
        return self

    def with_headers(self, **headers) -> "RequestWithOptionalArgs":
        self.__headers.update(headers)
        return self

    def with_cookies(self, **cookies) -> "RequestWithOptionalArgs":
        self.__cookies.update(cookies)
        return self

    def with_data(self, data) -> "RequestWithOptionalArgs":
        self.__data = data
        return self

    def set_timeout(self, timeout: float) -> "RequestWithOptionalArgs":
        self.__timeout = timeout
        return self

    def set_verify(self, verify: bool) -> "RequestWithOptionalArgs":
        self.__verify = verify
        return self

    def set_allow_redirects(self, allow_redirects: bool) -> "RequestWithOptionalArgs":
        self.__allow_redirects = allow_redirects
        return self

    def perform(self) -> TRequest:
        """build TRequest object with configs"""
        return TRequest(
            method=self.__method,
            url=self.__url,
            params=self.__params,
            headers=self.__headers,
            data=self.__data,
            timeout=self.__timeout,
            verify=self.__verify,
            allow_redirects=self.__allow_redirects,
        )


class Request(object):
    def get(self, url: Text) -> RequestWithOptionalArgs:
        return RequestWithOptionalArgs(MethodEnum.GET, url)

    def post(self, url: Text) -> RequestWithOptionalArgs:
        return RequestWithOptionalArgs(MethodEnum.POST, url)

    def put(self, url: Text) -> RequestWithOptionalArgs:
        return RequestWithOptionalArgs(MethodEnum.PUT, url)

    def head(self, url: Text) -> RequestWithOptionalArgs:
        return RequestWithOptionalArgs(MethodEnum.HEAD, url)

    def delete(self, url: Text) -> RequestWithOptionalArgs:
        return RequestWithOptionalArgs(MethodEnum.DELETE, url)

    def options(self, url: Text) -> RequestWithOptionalArgs:
        return RequestWithOptionalArgs(MethodEnum.OPTIONS, url)

    def patch(self, url: Text) -> RequestWithOptionalArgs:
        return RequestWithOptionalArgs(MethodEnum.PATCH, url)


class StepValidation(object):
    def __init__(
        self,
        name: Text,
        variables: Dict,
        extractors: Dict,
        request: TRequest = None,
        testcase: Callable = None,
    ):
        self.__name = name
        self.__variables = variables
        self.__extractors = extractors
        self.__request: TRequest = request
        self.__testcase: Callable = testcase
        self.__validators = []

    @property
    def request(self) -> TRequest:
        return self.__request

    @property
    def testcase(self) -> TestCase:
        return self.__testcase

    def assert_equal(self, jmes_path: Text, expected_value: Any) -> "StepValidation":
        self.__validators.append({"eq": [jmes_path, expected_value]})
        return self

    def assert_greater_than(
        self, jmes_path: Text, expected_value: Any
    ) -> "StepValidation":
        self.__validators.append({"gt": [jmes_path, expected_value]})
        return self

    def assert_less_than(
        self, jmes_path: Text, expected_value: Any
    ) -> "StepValidation":
        self.__validators.append({"lt": [jmes_path, expected_value]})
        return self

    def perform(self) -> TStep:
        return TStep(
            name=self.__name,
            variables=self.__variables,
            request=self.__request,
            testcase=self.__testcase,
            extract=self.__extractors,
            validate=self.__validators,
        )


class Step(object):
    def __init__(self, name: Text):
        self.__name = name
        self.__variables = {}
        self.__extractors = {}

    def with_variables(self, **variables) -> "Step":
        self.__variables.update(variables)
        return self

    def set_extractor(self, var_name: Text, jmes_path: Text) -> "Step":
        self.__extractors[var_name] = jmes_path
        return self

    def run_request(self, req_obj: RequestWithOptionalArgs) -> "StepValidation":
        return StepValidation(
            self.__name, self.__variables, self.__extractors, request=req_obj.perform()
        )

    def run_testcase(self, testcase: Callable) -> "StepValidation":
        return StepValidation(
            self.__name, self.__variables, self.__extractors, testcase=testcase
        )
