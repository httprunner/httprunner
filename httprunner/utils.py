# encoding: utf-8

import collections
import copy
import io
import itertools
import json
import os.path
import re
import uuid
from typing import Union

import sentry_sdk
from loguru import logger

from httprunner import exceptions, __version__
from httprunner.exceptions import ParamsError

absolute_http_url_regexp = re.compile(r"^https?://", re.I)


def init_sentry_sdk():
    sentry_sdk.init(
        dsn="https://cc6dd86fbe9f4e7fbd95248cfcff114d@sentry.io/1862849",
        release=f"httprunner@{__version__}"
    )

    with sentry_sdk.configure_scope() as scope:
        scope.set_user({"id": uuid.getnode()})


def set_os_environ(variables_mapping):
    """ set variables mapping to os.environ
    """
    for variable in variables_mapping:
        os.environ[variable] = variables_mapping[variable]
        logger.debug(f"Set OS environment variable: {variable}")


def unset_os_environ(variables_mapping):
    """ set variables mapping to os.environ
    """
    for variable in variables_mapping:
        os.environ.pop(variable)
        logger.debug(f"Unset OS environment variable: {variable}")


def get_os_environ(variable_name):
    """ get value of environment variable.

    Args:
        variable_name(str): variable name

    Returns:
        value of environment variable.

    Raises:
        exceptions.EnvNotFound: If environment variable not found.

    """
    try:
        return os.environ[variable_name]
    except KeyError:
        raise exceptions.EnvNotFound(variable_name)


def build_url(base_url, path):
    """ prepend url with base_url unless it's already an absolute URL """
    if absolute_http_url_regexp.match(path):
        return path
    elif base_url:
        return "{}/{}".format(base_url.rstrip("/"), path.lstrip("/"))
    else:
        raise ParamsError("base url missed!")


def query_json(json_content, query, delimiter='.'):
    """ Do an xpath-like query with json_content.

    Args:
        json_content (dict/list/string): content to be queried.
        query (str): query string.
        delimiter (str): delimiter symbol.

    Returns:
        str: queried result.

    Examples:
        >>> json_content = {
            "ids": [1, 2, 3, 4],
            "person": {
                "name": {
                    "first_name": "Leo",
                    "last_name": "Lee",
                },
                "age": 29,
                "cities": ["Guangzhou", "Shenzhen"]
            }
        }
        >>>
        >>> query_json(json_content, "person.name.first_name")
        >>> Leo
        >>>
        >>> query_json(json_content, "person.name.first_name.0")
        >>> L
        >>>
        >>> query_json(json_content, "person.cities.0")
        >>> Guangzhou

    """
    raise_flag = False
    response_body = f"response body: {json_content}\n"
    try:
        for key in query.split(delimiter):
            if isinstance(json_content, (list, str, bytes)):
                json_content = json_content[int(key)]
            elif isinstance(json_content, dict):
                json_content = json_content[key]
            else:
                logger.error(
                    f"invalid type value: {json_content}({type(json_content)})")
                raise_flag = True
    except (KeyError, ValueError, IndexError):
        raise_flag = True

    if raise_flag:
        err_msg = f"Failed to extract! => {query}\n"
        err_msg += response_body
        logger.error(err_msg)
        raise exceptions.ExtractFailure(err_msg)

    return json_content


def lower_dict_keys(origin_dict):
    """ convert keys in dict to lower case

    Args:
        origin_dict (dict): mapping data structure

    Returns:
        dict: mapping with all keys lowered.

    Examples:
        >>> origin_dict = {
            "Name": "",
            "Request": "",
            "URL": "",
            "METHOD": "",
            "Headers": "",
            "Data": ""
        }
        >>> lower_dict_keys(origin_dict)
            {
                "name": "",
                "request": "",
                "url": "",
                "method": "",
                "headers": "",
                "data": ""
            }

    """
    if not origin_dict or not isinstance(origin_dict, dict):
        return origin_dict

    return {
        key.lower(): value
        for key, value in origin_dict.items()
    }


def lower_test_dict_keys(test_dict):
    """ convert keys in test_dict to lower case, convertion will occur in two places:
        1, all keys in test_dict;
        2, all keys in test_dict["request"]
    """
    # convert keys in test_dict
    test_dict = lower_dict_keys(test_dict)

    if "request" in test_dict:
        # convert keys in test_dict["request"]
        test_dict["request"] = lower_dict_keys(test_dict["request"])

    return test_dict


def deepcopy_dict(data):
    """ deepcopy dict data, ignore file object (_io.BufferedReader)

    Args:
        data (dict): dict data structure
            {
                'a': 1,
                'b': [2, 4],
                'c': lambda x: x+1,
                'd': open('LICENSE'),
                'f': {
                    'f1': {'a1': 2},
                    'f2': io.open('LICENSE', 'rb'),
                }
            }

    Returns:
        dict: deep copied dict data, with file object unchanged.

    """
    try:
        return copy.deepcopy(data)
    except TypeError:
        copied_data = {}
        for key, value in data.items():
            if isinstance(value, dict):
                copied_data[key] = deepcopy_dict(value)
            else:
                try:
                    copied_data[key] = copy.deepcopy(value)
                except TypeError:
                    copied_data[key] = value

        return copied_data


def ensure_mapping_format(variables):
    """ ensure variables are in mapping format.

    Args:
        variables (list/dict): original variables

    Returns:
        dict: ensured variables in dict format

    Examples:
        >>> variables = [
                {"a": 1},
                {"b": 2}
            ]
        >>> print(ensure_mapping_format(variables))
            {
                "a": 1,
                "b": 2
            }

    """
    if isinstance(variables, list):
        variables_dict = {}
        for map_dict in variables:
            variables_dict.update(map_dict)

        return variables_dict

    elif isinstance(variables, dict):
        return variables

    else:
        raise exceptions.ParamsError("variables format error!")


def extend_variables(raw_variables, override_variables):
    """ extend raw_variables with override_variables.
        override_variables will merge and override raw_variables.

    Args:
        raw_variables (list):
        override_variables (list):

    Returns:
        dict: extended variables mapping

    Examples:
        >>> raw_variables = [{"var1": "val1"}, {"var2": "val2"}]
        >>> override_variables = [{"var1": "val111"}, {"var3": "val3"}]
        >>> extend_variables(raw_variables, override_variables)
            {
                'var1', 'val111',
                'var2', 'val2',
                'var3', 'val3'
            }

    """
    if not raw_variables:
        override_variables_mapping = ensure_mapping_format(override_variables)
        return override_variables_mapping

    elif not override_variables:
        raw_variables_mapping = ensure_mapping_format(raw_variables)
        return raw_variables_mapping

    else:
        raw_variables_mapping = ensure_mapping_format(raw_variables)
        override_variables_mapping = ensure_mapping_format(override_variables)
        raw_variables_mapping.update(override_variables_mapping)
        return raw_variables_mapping


def get_testcase_io(testcase):
    """ get and print testcase input(variables) and output(export).

    Args:
        testcase (unittest.suite.TestSuite): corresponding to one YAML/JSON file, it has been set two attributes:
            config: parsed config block
            runner: initialized runner.Runner() with config
    Returns:
        dict: input(variables) and output mapping.

    """
    test_runner = testcase.runner
    variables = testcase.config.get("variables", {})
    output_list = testcase.config.get("export") \
        or testcase.config.get("output", [])
    export_mapping = test_runner.export_variables(output_list)

    return {
        "in": variables,
        "out": export_mapping
    }


def print_info(info_mapping):
    """ print info in mapping.

    Args:
        info_mapping (dict): input(variables) or output mapping.

    Examples:
        >>> info_mapping = {
                "var_a": "hello",
                "var_b": "world"
            }
        >>> info_mapping = {
                "status_code": 500
            }
        >>> print_info(info_mapping)
        ==================== Output ====================
        Key              :  Value
        ---------------- :  ----------------------------
        var_a            :  hello
        var_b            :  world
        ------------------------------------------------

    """
    if not info_mapping:
        return

    content_format = "{:<16} : {:<}\n"
    content = "\n==================== Output ====================\n"
    content += content_format.format("Variable", "Value")
    content += content_format.format("-" * 16, "-" * 29)

    for key, value in info_mapping.items():
        if isinstance(value, (tuple, collections.deque)):
            continue
        elif isinstance(value, (dict, list)):
            value = json.dumps(value)
        elif value is None:
            value = "None"

        content += content_format.format(key, value)

    content += "-" * 48 + "\n"
    logger.info(content)


def create_scaffold(project_name):
    """ create scaffold with specified project name.
    """
    if os.path.isdir(project_name):
        logger.warning(f"Folder {project_name} exists, please specify a new folder name.")
        return

    logger.info(f"Start to create new project: {project_name}")
    logger.info(f"CWD: {os.getcwd()}")

    def create_folder(path):
        os.makedirs(path)
        msg = f"created folder: {path}"
        logger.info(msg)

    def create_file(path, file_content=""):
        with open(path, 'w') as f:
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
        - token: content.token
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
    ignore_content = "\n".join([
        ".env",
        "reports/*",
        "__pycache__/*",
        "*.pyc",
        ".python-version",
        "logs/*"
    ])
    demo_debugtalk_content = """
import time

def sleep(n_secs):
    time.sleep(n_secs)
"""
    demo_env_content = "\n".join([
        "USERNAME=leolee",
        "PASSWORD=123456"
    ])

    create_folder(project_name)
    create_folder(os.path.join(project_name, "api"))
    create_folder(os.path.join(project_name, "testcases"))
    create_folder(os.path.join(project_name, "testsuites"))
    create_folder(os.path.join(project_name, "reports"))
    create_file(os.path.join(project_name, "api", "demo_api.yml"), demo_api_content)
    create_file(os.path.join(project_name, "testcases", "demo_testcase.yml"), demo_testcase_content)
    create_file(os.path.join(project_name, "testsuites", "demo_testsuite.yml"), demo_testsuite_content)
    create_file(os.path.join(project_name, "debugtalk.py"), demo_debugtalk_content)
    create_file(os.path.join(project_name, ".env"), demo_env_content)
    create_file(os.path.join(project_name, ".gitignore"), ignore_content)


def gen_cartesian_product(*args):
    """ generate cartesian product for lists

    Args:
        args (list of list): lists to be generated with cartesian product

    Returns:
        list: cartesian product in list

    Examples:

        >>> arg1 = [{"a": 1}, {"a": 2}]
        >>> arg2 = [{"x": 111, "y": 112}, {"x": 121, "y": 122}]
        >>> args = [arg1, arg2]
        >>> gen_cartesian_product(*args)
        >>> # same as below
        >>> gen_cartesian_product(arg1, arg2)
            [
                {'a': 1, 'x': 111, 'y': 112},
                {'a': 1, 'x': 121, 'y': 122},
                {'a': 2, 'x': 111, 'y': 112},
                {'a': 2, 'x': 121, 'y': 122}
            ]

    """
    if not args:
        return []
    elif len(args) == 1:
        return args[0]

    product_list = []
    for product_item_tuple in itertools.product(*args):
        product_item_dict = {}
        for item in product_item_tuple:
            product_item_dict.update(item)

        product_list.append(product_item_dict)

    return product_list


def omit_long_data(body, omit_len=512):
    """ omit too long str/bytes
    """
    if not isinstance(body, (str, bytes)):
        return body

    body_len = len(body)
    if body_len <= omit_len:
        return body

    omitted_body = body[0:omit_len]

    appendix_str = f" ... OMITTED {body_len - omit_len} CHARACTORS ..."
    if isinstance(body, bytes):
        appendix_str = appendix_str.encode("utf-8")

    return omitted_body + appendix_str


def dump_json_file(json_data: Union[dict, list], json_file_abs_path: str) -> None:
    """ dump json data to file
    """
    class PythonObjectEncoder(json.JSONEncoder):
        def default(self, obj):
            try:
                return super().default(self, obj)
            except TypeError:
                return str(obj)

    file_foder_path = os.path.dirname(json_file_abs_path)
    if not os.path.isdir(file_foder_path):
        os.makedirs(file_foder_path)

    try:
        with io.open(json_file_abs_path, 'w', encoding='utf-8') as outfile:
            json.dump(
                json_data,
                outfile,
                indent=4,
                separators=(',', ':'),
                ensure_ascii=False,
                cls=PythonObjectEncoder
            )

        msg = f"dump file: {json_file_abs_path}"
        logger.info(msg)

    except TypeError as ex:
        msg = f"Failed to dump json file: {json_file_abs_path}\nReason: {ex}"
        logger.error(msg)


def prepare_log_file_abs_path(test_path: str, file_name: str) -> str:
    """ prepare dump json file absolute path.
    """
    current_working_dir = os.getcwd()

    if not test_path:
        # running passed in testcase/testsuite data structure
        dump_file_name = f"tests_mapping.{file_name}"
        dumped_json_file_abs_path = os.path.join(current_working_dir, "logs", dump_file_name)
        return dumped_json_file_abs_path

    # both test_path and pwd_dir_path are absolute path
    logs_dir_path = os.path.join(current_working_dir, "logs")

    if os.path.isdir(test_path):
        file_foder_path = os.path.join(logs_dir_path, test_path)
        dump_file_name = f"all.{file_name}"
    else:
        file_relative_folder_path, test_file = os.path.split(test_path)
        file_foder_path = os.path.join(logs_dir_path, file_relative_folder_path)
        test_file_name, _file_suffix = os.path.splitext(test_file)
        dump_file_name = f"{test_file_name}.{file_name}"

    dumped_json_file_abs_path = os.path.join(file_foder_path, dump_file_name)
    return dumped_json_file_abs_path
