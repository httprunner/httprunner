import os.path
import subprocess
import sys

from loguru import logger
from sentry_sdk import capture_message


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

    def show_tree(prj_name):
        try:
            print(f"\n$ tree {prj_name} -a")
            subprocess.run(["tree", prj_name, "-a"])
            print("")
        except FileNotFoundError:
            logger.warning("tree command not exists, ignore.")

    if os.path.isdir(project_name):
        logger.warning(
            f"Project folder {project_name} exists, please specify a new project name."
        )
        show_tree(project_name)
        return 1
    elif os.path.isfile(project_name):
        logger.warning(
            f"Project name {project_name} conflicts with existed file, please specify a new one."
        )
        return 1

    logger.info(f"Create new project: {project_name}")
    print(f"Project Root Dir: {os.path.join(os.getcwd(), project_name)}\n")

    def create_folder(path):
        os.makedirs(path)
        msg = f"created folder: {path}"
        print(msg)

    def create_file(path, file_content=""):
        with open(path, "w", encoding="utf-8") as f:
            f.write(file_content)
        msg = f"created file: {path}"
        print(msg)

    demo_testcase_request_content = """
config:
    name: "request methods testcase with functions"
    variables:
        foo1: config_bar1
        foo2: config_bar2
        expect_foo1: config_bar1
        expect_foo2: config_bar2
    base_url: "https://postman-echo.com"
    verify: False
    export: ["foo3"]

teststeps:
-
    name: get with params
    variables:
        foo1: bar11
        foo2: bar21
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
        foo3: "body.args.foo2"
    validate:
        - eq: ["status_code", 200]
        - eq: ["body.args.foo1", "bar11"]
        - eq: ["body.args.sum_v", "3"]
        - eq: ["body.args.foo2", "bar21"]
-
    name: post raw text
    variables:
        foo1: "bar12"
        foo3: "bar32"
    request:
        method: POST
        url: /post
        headers:
            User-Agent: HttpRunner/${get_httprunner_version()}
            Content-Type: "text/plain"
        data: "This is expected to be sent back as part of response body: $foo1-$foo2-$foo3."
    validate:
        - eq: ["status_code", 200]
        - eq: ["body.data", "This is expected to be sent back as part of response body: bar12-$expect_foo2-bar32."]
-
    name: post form data
    variables:
        foo2: bar23
    request:
        method: POST
        url: /post
        headers:
            User-Agent: HttpRunner/${get_httprunner_version()}
            Content-Type: "application/x-www-form-urlencoded"
        data: "foo1=$foo1&foo2=$foo2&foo3=$foo3"
    validate:
        - eq: ["status_code", 200]
        - eq: ["body.form.foo1", "$expect_foo1"]
        - eq: ["body.form.foo2", "bar23"]
        - eq: ["body.form.foo3", "bar21"]
"""
    demo_testcase_with_ref_content = """
config:
    name: "request methods testcase: reference testcase"
    variables:
        foo1: testsuite_config_bar1
        expect_foo1: testsuite_config_bar1
        expect_foo2: config_bar2
    base_url: "https://postman-echo.com"
    verify: False

teststeps:
-
    name: request with functions
    variables:
        foo1: testcase_ref_bar1
        expect_foo1: testcase_ref_bar1
    testcase: testcases/demo_testcase_request.yml
    export:
        - foo3
-
    name: post form data
    variables:
        foo1: bar1
    request:
        method: POST
        url: /post
        headers:
            User-Agent: HttpRunner/${get_httprunner_version()}
            Content-Type: "application/x-www-form-urlencoded"
        data: "foo1=$foo1&foo2=$foo3"
    validate:
        - eq: ["status_code", 200]
        - eq: ["body.form.foo1", "bar1"]
        - eq: ["body.form.foo2", "bar21"]
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

    show_tree(project_name)
    return 0


def main_scaffold(args):
    capture_message("startproject with scaffold")
    sys.exit(create_scaffold(args.project_name))
