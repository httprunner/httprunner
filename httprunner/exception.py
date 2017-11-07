#coding: utf-8
import json

try:
    FileNotFoundError = FileNotFoundError
except NameError:
    FileNotFoundError = IOError

try:
    JSONDecodeError = json.decoder.JSONDecodeError
except AttributeError:
    JSONDecodeError = ValueError

class MyBaseError(BaseException):
    pass

class FileFormatError(MyBaseError):
    pass

class ParamsError(MyBaseError):
    pass

class ResponseError(MyBaseError):
    pass

class ParseResponseError(MyBaseError):
    pass

class ValidationError(MyBaseError):
    pass

class NotFoundError(MyBaseError):
    pass

class FunctionNotFound(NotFoundError):
    pass

class VariableNotFound(NotFoundError):
    pass

class ApiNotFound(NotFoundError):
    pass

class SuiteNotFound(NotFoundError):
    pass

class TestcaseNotFound(NotFoundError):
    pass
