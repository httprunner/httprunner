# encoding: utf-8

try:
    # monkey patch at beginning to avoid RecursionError when running locust.
    from gevent import monkey
    if not monkey.is_module_patched('socket'):
        print("========== monkey patch all ==========")
        monkey.patch_all()
    else:
        print("========== monkey patched ==========")
except ImportError:
    pass

from httprunner.api import HttpRunner
