import collections
import copy
import json
import os.path
import platform
import uuid
from multiprocessing import Queue
import itertools
from typing import Dict, List, Any, Union, Text

import sentry_sdk
from loguru import logger

from httprunner import __version__
from httprunner import exceptions
from httprunner.models import VariablesMapping


def init_sentry_sdk():
    sentry_sdk.init(
        dsn="https://460e31339bcb428c879aafa6a2e78098@sentry.io/5263855",
        release="httprunner@{}".format(__version__),
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

    return {key.lower(): value for key, value in origin_dict.items()}


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


def get_platform():
    return {
        "httprunner_version": __version__,
        "python_version": "{} {}".format(
            platform.python_implementation(), platform.python_version()
        ),
        "platform": platform.platform(),
    }


def sort_dict_by_custom_order(raw_dict: Dict, custom_order: List):
    def get_index_from_list(lst: List, item: Any):
        try:
            return lst.index(item)
        except ValueError:
            # item is not in lst
            return len(lst) + 1

    return dict(
        sorted(raw_dict.items(), key=lambda i: get_index_from_list(custom_order, i[0]))
    )


class ExtendJSONEncoder(json.JSONEncoder):
    """ especially used to safely dump json data with python object, such as MultipartEncoder
    """

    def default(self, obj):
        try:
            return super(ExtendJSONEncoder, self).default(obj)
        except (UnicodeDecodeError, TypeError):
            return repr(obj)


def merge_variables(
    variables: VariablesMapping, variables_to_be_overridden: VariablesMapping
) -> VariablesMapping:
    """ merge two variables mapping, the first variables have higher priority
    """
    step_new_variables = {}
    for key, value in variables.items():
        if f"${key}" == value or "${" + key + "}" == value:
            # e.g. {"base_url": "$base_url"}
            # or {"base_url": "${base_url}"}
            continue

        step_new_variables[key] = value

    merged_variables = copy.copy(variables_to_be_overridden)
    merged_variables.update(step_new_variables)
    return merged_variables


def is_support_multiprocessing() -> bool:
    try:
        Queue()
        return True
    except (ImportError, OSError):
        # system that does not support semaphores(dependency of multiprocessing), like Android termux
        return False


def gen_cartesian_product(*args: List[Dict]) -> List[Dict]:
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
