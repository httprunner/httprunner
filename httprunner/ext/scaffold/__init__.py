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

    demo_api_content = """
name: demo api
variables:
    var1: value1
    var2: value2
request:
    url: /api/path/$var1
    method: POST
    headers:
        Content-Type: "application/json"
    json:
        key: $var2
validate:
    - eq: ["status_code", 200]
"""
    demo_testcase_content = """
config:
    name: "demo testcase"
    variables:
        device_sn: "ABC"
        username: ${ENV(USERNAME)}
        password: ${ENV(PASSWORD)}
    base_url: "http://127.0.0.1:5000"

teststeps:
-
    name: demo step 1
    api: path/to/api1.yml
    variables:
        user_agent: 'iOS/10.3'
        device_sn: $device_sn
    extract:
        token: content.token
    validate:
        - eq: ["status_code", 200]
-
    name: demo step 2
    api: path/to/api2.yml
    variables:
        token: $token
"""
    demo_testsuite_content = """
config:
    name: "demo testsuite"
    variables:
        device_sn: "XYZ"
    base_url: "http://127.0.0.1:5000"

testcases:
-
    name: call demo_testcase with data 1
    testcase: path/to/demo_testcase.yml
    variables:
        device_sn: $device_sn
-
    name: call demo_testcase with data 2
    testcase: path/to/demo_testcase.yml
    variables:
        device_sn: $device_sn
"""
    ignore_content = "\n".join(
        [".env", "reports/*", "__pycache__/*", "*.pyc", ".python-version", "logs/*"]
    )
    demo_debugtalk_content = """
import time

def sleep(n_secs):
    time.sleep(n_secs)
"""
    demo_env_content = "\n".join(["USERNAME=leolee", "PASSWORD=123456"])

    create_folder(project_name)
    create_folder(os.path.join(project_name, "api"))
    create_folder(os.path.join(project_name, "testcases"))
    create_folder(os.path.join(project_name, "testsuites"))
    create_folder(os.path.join(project_name, "reports"))
    create_file(os.path.join(project_name, "api", "demo_api.yml"), demo_api_content)
    create_file(
        os.path.join(project_name, "testcases", "demo_testcase.yml"),
        demo_testcase_content,
    )
    create_file(
        os.path.join(project_name, "testsuites", "demo_testsuite.yml"),
        demo_testsuite_content,
    )
    create_file(os.path.join(project_name, "debugtalk.py"), demo_debugtalk_content)
    create_file(os.path.join(project_name, ".env"), demo_env_content)
    create_file(os.path.join(project_name, ".gitignore"), ignore_content)


def main_scaffold(args):
    create_scaffold(args.project_name)
    sys.exit(0)
