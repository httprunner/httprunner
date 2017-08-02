#coding: utf-8

class MyBaseError(BaseException):
    pass

class ParamsError(MyBaseError):
    pass

class ParseResponseError(MyBaseError):
    pass

class ValidationError(MyBaseError):
    pass
