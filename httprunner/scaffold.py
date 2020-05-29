import os.path
import sys

from loguru import logger


def init_parser_scaffold(subparsers):
    sub_parser_scaffold = subparsers.add_parser(
        "startproject", help="Create a new project with template structure."
    )
    sub_parser_scaffold.add_argument(
        "project_name", type=str, nargs="?", help="Specify new project name."
    )
    return sub_parser_scaffold


def create_scaffold(project_name):
    """ create scaffold with specified project name.
    """
    if os.path.isdir(project_name):
        logger.warning(
            f"Folder {project_name} exists, please specify a new folder name."
        )
        return

    logger.info(f"Start to create new project: {project_name}")
    logger.info(f"CWD: {os.getcwd()}")

    def create_folder(path):
        os.makedirs(path)
        msg = f"created folder: {path}"
        logger.info(msg)

    def create_file(path, file_content=""):
        with open(path, "w") as f:
            f.write(file_content)
        msg = f"created file: {path}"
        logger.info(msg)

    demo_testcase_request_content = """
config:
    name: "request methods testcase with functions"
    variables:
        foo1: session_bar1
    base_url: "https://postman-echo.com"
    verify: False

teststeps:
-
    name: get with params
    variables:
        foo1: bar1
        foo2: session_bar2
        sum_v: "${sum_two(1, 2)}"
    request:
        method: GET
        url: /get
        params:
            foo1: $foo1
            foo2: $foo2
            sum_v: $sum_v
        headers:
            User-Agent: HttpRunner/${get_httprunner_version()}
    extract:
        session_foo2: "body.args.foo2"
    validate:
        - eq: ["status_code", 200]
        - eq: ["body.args.foo1", "session_bar1"]
        - eq: ["body.args.sum_v", 3]
        - eq: ["body.args.foo2", "session_bar2"]
-
    name: post raw text
    variables:
        foo1: "hello world"
        foo3: "$session_foo2"
    request:
        method: POST
        url: /post
        headers:
            User-Agent: HttpRunner/${get_httprunner_version()}
            Content-Type: "text/plain"
        data: "This is expected to be sent back as part of response body: $foo1-$foo3."
    validate:
        - eq: ["status_code", 200]
        - eq: ["body.data", "This is expected to be sent back as part of response body: session_bar1-session_bar2."]
"""
    demo_testcase_with_ref_content = """
config:
    name: "request methods testcase: reference testcase"
    variables:
        foo1: session_bar1
    base_url: "https://postman-echo.com"
    verify: False

teststeps:
-
    name: request with referenced testcase
    variables:
        foo1: override_bar1
    # NOTICE: relative testcase path based on debugtalk.py
    testcase: testcases/demo_testcase_request.yml
"""
    ignore_content = "\n".join(
        [".env", "reports/*", "__pycache__/*", "*.pyc", ".python-version", "logs/*"]
    )
    demo_debugtalk_content = """import time

from httprunner import __version__


def get_httprunner_version():
    return __version__


def sum_two(m, n):
    return m + n


def sleep(n_secs):
    time.sleep(n_secs)
"""
    demo_env_content = "\n".join(["USERNAME=leolee", "PASSWORD=123456"])

    create_folder(project_name)
    create_folder(os.path.join(project_name, "har"))
    create_folder(os.path.join(project_name, "testcases"))
    create_folder(os.path.join(project_name, "reports"))

    create_file(
        os.path.join(project_name, "testcases", "demo_testcase_request.yml"),
        demo_testcase_request_content,
    )
    create_file(
        os.path.join(project_name, "testcases", "demo_testcase_ref.yml"),
        demo_testcase_with_ref_content,
    )
    create_file(os.path.join(project_name, "debugtalk.py"), demo_debugtalk_content)
    create_file(os.path.join(project_name, ".env"), demo_env_content)
    create_file(os.path.join(project_name, ".gitignore"), ignore_content)


def main_scaffold(args):
    create_scaffold(args.project_name)
    sys.exit(0)
