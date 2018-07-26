# encoding: utf-8

from httprunner.compat import JSONDecodeError, FileNotFoundError

""" failure type exceptions
    these exceptions will mark test as failure
"""

class MyBaseFailure(BaseException):
    pass

class ValidationFailure(MyBaseFailure):
    pass

class ExtractFailure(MyBaseFailure):
    pass


""" error type exceptions
    these exceptions will mark test as error
"""

class MyBaseError(BaseException):
    pass

class FileFormatError(MyBaseError):
    pass

class ParamsError(MyBaseError):
    pass

class NotFoundError(MyBaseError):
    pass

class FileNotFound(FileNotFoundError, NotFoundError):
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
