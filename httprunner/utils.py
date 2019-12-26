# encoding: utf-8

import collections
import copy
import io
import itertools
import json
import os.path
import re
from datetime import datetime

from httprunner import exceptions, logger
from httprunner.compat import basestring, bytes, is_py2
from httprunner.exceptions import ParamsError

absolute_http_url_regexp = re.compile(r"^https?://", re.I)


def set_os_environ(variables_mapping):
    """ set variables mapping to os.environ
    """
    for variable in variables_mapping:
        os.environ[variable] = variables_mapping[variable]
        logger.log_debug("Set OS environment variable: {}".format(variable))


def unset_os_environ(variables_mapping):
    """ set variables mapping to os.environ
    """
    for variable in variables_mapping:
        os.environ.pop(variable)
        logger.log_debug("Unset OS environment variable: {}".format(variable))


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
    response_body = u"response body: {}\n".format(json_content)
    try:
        for key in query.split(delimiter):
            if isinstance(json_content, (list, basestring)):
                json_content = json_content[int(key)]
            elif isinstance(json_content, dict):
                json_content = json_content[key]
            else:
                logger.log_error(
                    "invalid type value: {}({})".format(json_content, type(json_content)))
                raise_flag = True
    except (KeyError, ValueError, IndexError):
        raise_flag = True

    if raise_flag:
        err_msg = u"Failed to extract! => {}\n".format(query)
        err_msg += response_body
        logger.log_error(err_msg)
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

        if is_py2:
            if isinstance(key, unicode):
                key = key.encode("utf-8")
            if isinstance(value, unicode):
                value = value.encode("utf-8")

        content += content_format.format(key, value)

    content += "-" * 48 + "\n"
    logger.log_info(content)


def create_scaffold(project_name):
    """ create scaffold with specified project name.
    """
    if os.path.isdir(project_name):
        logger.log_warning(u"Folder {} exists, please specify a new folder name.".format(project_name))
        return

    logger.color_print("Start to create new project: {}".format(project_name), "GREEN")
    logger.color_print("CWD: {}\n".format(os.getcwd()), "BLUE")

    def create_folder(path):
        os.makedirs(path)
        msg = "created folder: {}".format(path)
        logger.color_print(msg, "BLUE")

    def create_file(path, file_content=""):
        with open(path, 'w') as f:
            f.write(file_content)
        msg = "created file: {}".format(path)
        logger.color_print(msg, "BLUE")

    def create_path(path, ptype, file_content=""):
        if ptype == "folder":
            os.makedirs(path)
        elif ptype == "file":
            with open(path, 'w') as f:
                f.write(file_content)

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


def prettify_json_file(file_list):
    """ prettify JSON testcase format
    """
    for json_file in set(file_list):
        if not json_file.endswith(".json"):
            logger.log_warning("Only JSON file format can be prettified, skip: {}".format(json_file))
            continue

        logger.color_print("Start to prettify JSON file: {}".format(json_file), "GREEN")

        dir_path = os.path.dirname(json_file)
        file_name, file_suffix = os.path.splitext(os.path.basename(json_file))
        outfile = os.path.join(dir_path, "{}.pretty.json".format(file_name))

        with io.open(json_file, 'r', encoding='utf-8') as stream:
            try:
                obj = json.load(stream)
            except ValueError as e:
                raise SystemExit(e)

        with io.open(outfile, 'w', encoding='utf-8') as out:
            json.dump(obj, out, indent=4, separators=(',', ': '))
            out.write('\n')

        print("success: {}".format(outfile))


def omit_long_data(body, omit_len=512):
    """ omit too long str/bytes
    """
    if not isinstance(body, basestring):
        return body

    body_len = len(body)
    if body_len <= omit_len:
        return body

    omitted_body = body[0:omit_len]

    appendix_str = " ... OMITTED {} CHARACTORS ...".format(body_len - omit_len)
    if isinstance(body, bytes):
        appendix_str = appendix_str.encode("utf-8")

    return omitted_body + appendix_str


def dump_json_file(json_data, json_file_abs_path):
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
            if is_py2:
                outfile.write(
                    unicode(json.dumps(
                        json_data,
                        indent=4,
                        separators=(',', ':'),
                        encoding="utf8",
                        ensure_ascii=False,
                        cls=PythonObjectEncoder
                    ))
                )
            else:
                json.dump(
                    json_data,
                    outfile,
                    indent=4,
                    separators=(',', ':'),
                    ensure_ascii=False,
                    cls=PythonObjectEncoder
                )

        msg = "dump file: {}".format(json_file_abs_path)
        logger.color_print(msg, "BLUE")

    except TypeError as ex:
        msg = "Failed to dump json file: {}\nReason: {}".format(json_file_abs_path, ex)
        logger.color_print(msg, "RED")


def prepare_dump_json_file_abs_path(project_mapping, tag_name):
    """ prepare dump json file absolute path.
    """
    pwd_dir_path = project_mapping.get("PWD") or os.getcwd()
    test_path = project_mapping.get("test_path")

    if not test_path:
        # running passed in testcase/testsuite data structure
        dump_file_name = "tests_mapping.{}.json".format(tag_name)
        dumped_json_file_abs_path = os.path.join(pwd_dir_path, "logs", dump_file_name)
        return dumped_json_file_abs_path

    # both test_path and pwd_dir_path are absolute path
    logs_dir_path = os.path.join(pwd_dir_path, "logs")
    test_path_relative_path = test_path[len(pwd_dir_path)+1:]

    if os.path.isdir(test_path):
        file_foder_path = os.path.join(logs_dir_path, test_path_relative_path)
        dump_file_name = "all.{}.json".format(tag_name)
    else:
        file_relative_folder_path, test_file = os.path.split(test_path_relative_path)
        file_foder_path = os.path.join(logs_dir_path, file_relative_folder_path)
        test_file_name, _file_suffix = os.path.splitext(test_file)
        dump_file_name = "{}.{}.json".format(test_file_name, tag_name)

    dumped_json_file_abs_path = os.path.join(file_foder_path, dump_file_name)
    return dumped_json_file_abs_path


def dump_logs(json_data, project_mapping, tag_name):
    """ dump tests data to json file.
        the dumped file is located in PWD/logs folder.

    Args:
        json_data (list/dict): json data to dump
        project_mapping (dict): project info
        tag_name (str): tag name, loaded/parsed/summary

    """
    json_file_abs_path = prepare_dump_json_file_abs_path(project_mapping, tag_name)
    dump_json_file(json_data, json_file_abs_path)


def get_python2_retire_msg():
    retire_day = datetime(2020, 1, 1)
    today = datetime.now()
    left_days = (retire_day - today).days

    if left_days > 0:
        retire_msg = "Python 2 will retire in {} days, why not move to Python 3?".format(left_days)
    else:
        retire_msg = "Python 2 has been retired, you should move to Python 3."

    return retire_msg
