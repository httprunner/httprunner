""" failure type exceptions
    these exceptions will mark test as failure
"""


class MyBaseFailure(Exception):
    pass


class ParseTestsFailure(MyBaseFailure):
    pass


class ValidationFailure(MyBaseFailure):
    pass


class ExtractFailure(MyBaseFailure):
    pass


class SetupHooksFailure(MyBaseFailure):
    pass


class TeardownHooksFailure(MyBaseFailure):
    pass


""" error type exceptions
    these exceptions will mark test as error
"""


class MyBaseError(Exception):
    pass


class FileFormatError(MyBaseError):
    pass


class TestCaseFormatError(FileFormatError):
    pass


class TestSuiteFormatError(FileFormatError):
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


class EnvNotFound(NotFoundError):
    pass


class CSVNotFound(NotFoundError):
    pass


class ApiNotFound(NotFoundError):
    pass


class TestcaseNotFound(NotFoundError):
    pass


class SummaryEmpty(MyBaseError):
    """test result summary data is empty"""
