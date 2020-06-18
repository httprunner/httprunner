import importlib.util
import inspect
import os
import sys
from typing import List

if sys.argv[0].endswith("locusts"):
    try:
        # monkey patch ssl at beginning to avoid RecursionError when running locust.
        from gevent import monkey

        monkey.patch_ssl()
        from locust import main as locust_main
    except ImportError:
        msg = """
Locust is not installed, install first and try again.
install with pip:
$ pip install locust
"""
        print(msg)
        sys.exit(1)


""" converted pytest files from YAML/JSON testcases
"""
pytest_files: List = []


def is_httprunner_testcase(item):
    """ check if a variable is a HttpRunner testcase class
    """
    from httprunner import HttpRunner

    # TODO: skip referenced testcase
    return bool(
        inspect.isclass(item)
        and issubclass(item, HttpRunner)
        and item.__name__ != "HttpRunner"
    )


def prepare_locust_tests() -> List:
    """ prepare locust testcases

    Returns:
        list: testcase class list
    """

    locust_tests = []

    for pytest_file in pytest_files:
        spec = importlib.util.spec_from_file_location("module.name", pytest_file)
        module = importlib.util.module_from_spec(spec)
        spec.loader.exec_module(module)

        for name, item in vars(module).items():

            if not is_httprunner_testcase(item):
                continue

            for _ in range(item.config.weight):
                locust_tests.append(item)

    return locust_tests


def main_locusts():
    """ locusts entrance
    """
    from httprunner.utils import init_sentry_sdk
    from sentry_sdk import capture_message
    init_sentry_sdk()
    capture_message("start to run locusts")

    sys.argv[0] = "locust"
    if len(sys.argv) == 1:
        sys.argv.extend(["-h"])

    if sys.argv[1] in ["-h", "--help", "-V", "--version"]:
        locust_main.main()

    def get_arg_index(*target_args):
        for arg in target_args:
            if arg not in sys.argv:
                continue

            return sys.argv.index(arg) + 1

        return None

    # get testcase file path
    testcase_index = get_arg_index("-f", "--locustfile")
    if not testcase_index:
        print("Testcase file is not specified, exit 1.")
        sys.exit(1)

    from httprunner.make import main_make

    global pytest_files
    testcase_file_path = sys.argv[testcase_index]
    pytest_files = main_make([testcase_file_path])
    if not pytest_files:
        print("No valid testcases found, exit 1.")
        sys.exit(1)

    sys.argv[testcase_index] = os.path.join(os.path.dirname(__file__), "locustfile.py")

    locust_main.main()
