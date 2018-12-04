# encoding: utf-8

try:
    # monkey patch ssl at beginning to avoid RecursionError when running locust.
    from gevent import monkey
    if not monkey.is_module_patched('ssl'):
        monkey.patch_ssl()
except ImportError:
    pass

from httprunner.api import HttpRunner
