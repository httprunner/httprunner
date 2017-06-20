#coding: utf-8
from termcolor import colored

class MyBaseError(BaseException):
    def __init__(self, msg):
        self.msg = msg
        self.color_msg = colored(msg, 'red', attrs=['bold'])

    def __repr__(self):
        return self.msg

    def __str__(self):
        return self.color_msg

class ParamsError(MyBaseError):
    pass
