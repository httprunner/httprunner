import hashlib
import hmac
import imp
import importlib
import io
import os.path
import random
import re
import string
import types
from collections import OrderedDict

from httprunner import exception, logger
from requests.structures import CaseInsensitiveDict

try:
    string_type = basestring
    long_type = long
    PYTHON_VERSION = 2
except NameError:
    string_type = str
    long_type = int
    PYTHON_VERSION = 3

SECRET_KEY = "DebugTalk"


def gen_random_string(str_len):
    return ''.join(
        random.choice(string.ascii_letters + string.digits) for _ in range(str_len))

def gen_md5(*str_args):
    return hashlib.md5("".join(str_args).encode('utf-8')).hexdigest()

def get_sign(*args):
    content = ''.join(args).encode('ascii')
    sign_key = SECRET_KEY.encode('ascii')
    sign = hmac.new(sign_key, content, hashlib.sha1).hexdigest()
    return sign

def remove_prefix(text, prefix):
    """ remove prefix from text
    """
    if text.startswith(prefix):
        return text[len(prefix):]
    return text

def load_folder_files(folder_path, recursive=True):
    """ load folder path, return all files in list format.
    @param
        folder_path: specified folder path to load
        recursive: if True, will load files recursively
    """
    if isinstance(folder_path, (list, set)):
        files = []
        for path in set(folder_path):
            files.extend(load_folder_files(path, recursive))

        return files

    if not os.path.exists(folder_path):
        return []

    file_list = []

    for dirpath, dirnames, filenames in os.walk(folder_path):
        filenames_list = []

        for filename in filenames:
            if not filename.endswith(('.yml', '.yaml', '.json')):
                continue

            filenames_list.append(filename)

        for filename in filenames_list:
            file_path = os.path.join(dirpath, filename)
            file_list.append(file_path)

        if not recursive:
            break

    return file_list

def query_json(json_content, query, delimiter='.'):
    """ Do an xpath-like query with json_content.
    @param (json_content) json_content
        json_content = {
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
    @param (str) query
        "person.name.first_name"  =>  "Leo"
        "person.cities.0"         =>  "Guangzhou"
    @return queried result
    """
    if json_content == "":
        raise exception.ResponseError("response content is empty!")

    try:
        for key in query.split(delimiter):
            if isinstance(json_content, list):
                json_content = json_content[int(key)]
            elif isinstance(json_content, (dict, CaseInsensitiveDict)):
                json_content = json_content[key]
            else:
                raise exception.ParseResponseError(
                    "response content is in text format! failed to query key {}!".format(key))
    except (KeyError, ValueError, IndexError):
        raise exception.ParseResponseError("failed to query json when extracting response!")

    return json_content

def get_uniform_comparator(comparator):
    """ convert comparator alias to uniform name
    """
    if comparator in ["eq", "equals", "==", "is"]:
        return "equals"
    elif comparator in ["lt", "less_than"]:
        return "less_than"
    elif comparator in ["le", "less_than_or_equals"]:
        return "less_than_or_equals"
    elif comparator in ["gt", "greater_than"]:
        return "greater_than"
    elif comparator in ["ge", "greater_than_or_equals"]:
        return "greater_than_or_equals"
    elif comparator in ["ne", "not_equals"]:
        return "not_equals"
    elif comparator in ["str_eq", "string_equals"]:
        return "string_equals"
    elif comparator in ["len_eq", "length_equals", "count_eq"]:
        return "length_equals"
    elif comparator in ["len_gt", "count_gt", "length_greater_than", "count_greater_than"]:
        return "length_greater_than"
    elif comparator in ["len_ge", "count_ge", "length_greater_than_or_equals", \
        "count_greater_than_or_equals"]:
        return "length_greater_than_or_equals"
    elif comparator in ["len_lt", "count_lt", "length_less_than", "count_less_than"]:
        return "length_less_than"
    elif comparator in ["len_le", "count_le", "length_less_than_or_equals", \
        "count_less_than_or_equals"]:
        return "length_less_than_or_equals"
    else:
        return comparator

def deep_update_dict(origin_dict, override_dict):
    """ update origin dict with override dict recursively
    e.g. origin_dict = {'a': 1, 'b': {'c': 2, 'd': 4}}
         override_dict = {'b': {'c': 3}}
    return: {'a': 1, 'b': {'c': 3, 'd': 4}}
    """
    if not override_dict:
        return origin_dict

    for key, val in override_dict.items():
        if isinstance(val, dict):
            tmp = deep_update_dict(origin_dict.get(key, {}), val)
            origin_dict[key] = tmp
        elif val is None:
            # fix #64: when headers in test is None, it should inherit from config
            continue
        else:
            origin_dict[key] = override_dict[key]

    return origin_dict

def is_function(tup):
    """ Takes (name, object) tuple, returns True if it is a function.
    """
    name, item = tup
    return isinstance(item, types.FunctionType)

def is_variable(tup):
    """ Takes (name, object) tuple, returns True if it is a variable.
    """
    name, item = tup
    if callable(item):
        # function or class
        return False

    if isinstance(item, types.ModuleType):
        # imported module
        return False

    if name.startswith("_"):
        # private property
        return False

    return True

def get_imported_module(module_name):
    """ import module and return imported module
    """
    return importlib.import_module(module_name)

def get_imported_module_from_file(file_path):
    """ import module from python file path and return imported module
    """

    if PYTHON_VERSION == 3:
        imported_module = importlib.machinery.SourceFileLoader(
            'module_name', file_path).load_module()
    else:
        # Python 2.7
        imported_module = imp.load_source('module_name', file_path)

    return imported_module

def filter_module(module, filter_type):
    """ filter functions or variables from import module
    @params
        module: imported module
        filter_type: "function" or "variable"
    """
    filter_type = is_function if filter_type == "function" else is_variable
    module_functions_dict = dict(filter(filter_type, vars(module).items()))
    return module_functions_dict

def search_conf_item(start_path, item_type, item_name):
    """ search expected function or variable recursive upward
    @param
        start_path: search start path
        item_type: "function" or "variable"
        item_name: function name or variable name
    """
    dir_path = os.path.dirname(os.path.abspath(start_path))
    target_file = os.path.join(dir_path, "debugtalk.py")

    if os.path.isfile(target_file):
        imported_module = get_imported_module_from_file(target_file)
        items_dict = filter_module(imported_module, item_type)
        if item_name in items_dict:
            return items_dict[item_name]
        else:
            return search_conf_item(dir_path, item_type, item_name)

    if dir_path == start_path:
        # system root path
        err_msg = "{} not found in recursive upward path!".format(item_name)
        if item_type == "function":
            raise exception.FunctionNotFound(err_msg)
        else:
            raise exception.VariableNotFound(err_msg)

    return search_conf_item(dir_path, item_type, item_name)

def lower_dict_keys(origin_dict):
    """ convert keys in dict to lower case
    e.g.
        Name => name, Request => request
        URL => url, METHOD => method, Headers => headers, Data => data
    """
    if not origin_dict or not isinstance(origin_dict, dict):
        return origin_dict

    return {
        key.lower(): value
        for key, value in origin_dict.items()
    }

def lower_config_dict_key(config_dict):
    """ convert key in config dict to lower case, convertion will occur in three places:
        1, all keys in config dict;
        2, all keys in config["request"]
        3, all keys in config["request"]["headers"]
    """
    # convert keys in config dict
    config_dict = lower_dict_keys(config_dict)

    if "request" in config_dict:
        # convert keys in config["request"]
        config_dict["request"] = lower_dict_keys(config_dict["request"])

        # convert keys in config["request"]["headers"]
        if "headers" in config_dict["request"]:
            config_dict["request"]["headers"] = lower_dict_keys(config_dict["request"]["headers"])

    return config_dict

def convert_to_order_dict(map_list):
    """ convert mapping in list to ordered dict
    @param (list) map_list
        [
            {"a": 1},
            {"b": 2}
        ]
    @return (OrderDict)
        OrderDict({
            "a": 1,
            "b": 2
        })
    """
    ordered_dict = OrderedDict()
    for map_dict in map_list:
        ordered_dict.update(map_dict)

    return ordered_dict

def update_ordered_dict(ordered_dict, override_mapping):
    """ override ordered_dict with new mapping
    @param
        (OrderDict) ordered_dict
            OrderDict({
                "a": 1,
                "b": 2
            })
        (dict) override_mapping
            {"a": 3, "c": 4}
    @return (OrderDict)
        OrderDict({
            "a": 3,
            "b": 2,
            "c": 4
        })
    """
    for var, value in override_mapping.items():
        ordered_dict.update({var: value})

    return ordered_dict

def override_variables_binds(variables, new_mapping):
    """ convert variables in testcase to ordered mapping, with new_mapping overrided
    """
    if isinstance(variables, list):
        variables_ordered_dict = convert_to_order_dict(variables)
    elif isinstance(variables, OrderedDict):
        variables_ordered_dict = variables
    else:
        raise exception.ParamsError("variables error!")

    return update_ordered_dict(
        variables_ordered_dict,
        new_mapping
    )

def print_output(output):
    if not output:
        return

    content = "\n================== Output ==================\n"
    content += '{:<16}:  {:<}\n'.format("Variable", "Value")
    content += '{:<16}:  {:<}\n'.format("--------", "-----")

    for variable, value in output.items():

        if PYTHON_VERSION == 2:
            if isinstance(variable, unicode):
                variable = variable.encode("utf-8")
            if isinstance(value, unicode):
                value = value.encode("utf-8")

        content += '{:<16}:  {:<}\n'.format(variable, value)

    content += "============================================\n"

    logger.log_debug(content)

def create_scaffold(project_path):
    if os.path.isdir(project_path):
        folder_name = os.path.basename(project_path)
        logger.log_warning(u"Folder {} exists, please specify a new folder name.".format(folder_name))
        return

    logger.color_print("Start to create new project: {}\n".format(project_path), "GREEN")

    def create_path(path, ptype):
        if ptype == "folder":
            os.makedirs(path)
        elif ptype == "file":
            open(path, 'w').close()

        return "created {}: {}\n".format(ptype, path)

    path_list = [
        (project_path, "folder"),
        (os.path.join(project_path, "tests"), "folder"),
        (os.path.join(project_path, "tests", "api"), "folder"),
        (os.path.join(project_path, "tests", "suite"), "folder"),
        (os.path.join(project_path, "tests", "testcases"), "folder"),
        (os.path.join(project_path, "tests", "debugtalk.py"), "file")
    ]

    msg = ""
    for p in path_list:
        msg += create_path(p[0], p[1])

    logger.color_print(msg, "BLUE")

def load_dot_env_file(path):
    """ load .env file and set to os.environ
    """
    if not os.path.isfile(path):
        return

    logger.log_info("Loading environment variables from {}".format(path))
    with io.open(path, 'r', encoding='utf-8') as fp:
        for line in fp:
            variable, value = line.split("=")
            os.environ[variable] = value
            logger.log_debug("Loaded variable: {}".format(variable))
