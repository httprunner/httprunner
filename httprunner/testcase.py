from typing import Text, Any

from httprunner.schema import (
    TConfig,
    TStep,
    TRequest,
    MethodEnum,
)


class Config(object):
    def __init__(self, name: Text):
        self.__name = name
        self.__variables = {}
        self.__base_url = ""
        self.__verify = False
        self.__path = ""

    def variables(self, **variables) -> "Config":
        self.__variables.update(variables)
        return self

    def base_url(self, base_url: Text) -> "Config":
        self.__base_url = base_url
        return self

    def verify(self, verify: bool) -> "Config":
        self.__verify = verify
        return self

    def path(self, path: Text) -> "Config":
        self.__path = path
        return self

    def init(self) -> TConfig:
        return TConfig(
            name=self.__name,
            base_url=self.__base_url,
            verify=self.__verify,
            variables=self.__variables,
            path=self.__path,
        )


class RequestOptionalArgs(object):
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

    def with_params(self, **params) -> "RequestOptionalArgs":
        self.__params.update(params)
        return self

    def with_headers(self, **headers) -> "RequestOptionalArgs":
        self.__headers.update(headers)
        return self

    def with_cookies(self, **cookies) -> "RequestOptionalArgs":
        self.__cookies.update(cookies)
        return self

    def with_data(self, data) -> "RequestOptionalArgs":
        self.__data = data
        return self

    def set_timeout(self, timeout: float) -> "RequestOptionalArgs":
        self.__timeout = timeout
        return self

    def set_verify(self, verify: bool) -> "RequestOptionalArgs":
        self.__verify = verify
        return self

    def set_allow_redirects(self, allow_redirects: bool) -> "RequestOptionalArgs":
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
    def get(self, url: Text) -> RequestOptionalArgs:
        return RequestOptionalArgs(MethodEnum.GET, url)

    def post(self, url: Text) -> RequestOptionalArgs:
        return RequestOptionalArgs(MethodEnum.POST, url)

    def put(self, url: Text) -> RequestOptionalArgs:
        return RequestOptionalArgs(MethodEnum.PUT, url)

    def head(self, url: Text) -> RequestOptionalArgs:
        return RequestOptionalArgs(MethodEnum.HEAD, url)

    def delete(self, url: Text) -> RequestOptionalArgs:
        return RequestOptionalArgs(MethodEnum.DELETE, url)

    def options(self, url: Text) -> RequestOptionalArgs:
        return RequestOptionalArgs(MethodEnum.OPTIONS, url)

    def patch(self, url: Text) -> RequestOptionalArgs:
        return RequestOptionalArgs(MethodEnum.PATCH, url)


class Step(object):
    def __init__(self, name):
        self.__name = name
        self.__variables = {}
        self.__request = None
        self.__extract = {}
        self.__validators = []

    def with_variables(self, **variables) -> "Step":
        self.__variables.update(variables)
        return self

    def run_request(self, req_obj: RequestOptionalArgs) -> "Step":
        self.__request = req_obj.perform()
        return self

    def extract(self, var_name: Text, jmes_path: Text) -> "Step":
        self.__extract[var_name] = jmes_path
        return self

    def assert_equal(self, jmes_path: Text, expected_value: Any) -> "Step":
        self.__validators.append({"eq": [jmes_path, expected_value]})
        return self

    def assert_greater_than(self, jmes_path: Text, expected_value: Any) -> "Step":
        self.__validators.append({"gt": [jmes_path, expected_value]})
        return self

    def assert_less_than(self, jmes_path: Text, expected_value: Any) -> "Step":
        self.__validators.append({"lt": [jmes_path, expected_value]})
        return self

    def init(self) -> TStep:
        return TStep(
            name=self.__name,
            variables=self.__variables,
            request=self.__request,
            extract=self.__extract,
            validate=self.__validators,
        )

