# encoding: utf-8

try:
    # monkey patch at beginning to avoid RecursionError when running locust.
    from gevent import monkey; monkey.patch_all()
except ImportError:
    pass

from httprunner.api import HttpRunner
