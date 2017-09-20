import codecs
import hashlib
import hmac
import imp
import importlib
import json
import os.path
import random
import re
import string
import types
from collections import OrderedDict

import yaml
from ate import exception
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

def load_yaml_file(yaml_file):
    with codecs.open(yaml_file, 'r+', encoding='utf-8') as stream:
        return yaml.load(stream)

def load_json_file(json_file):
    with codecs.open(json_file, encoding='utf-8') as data_file:
        return json.load(data_file)

def load_testcases(testcase_file_path):
    file_suffix = os.path.splitext(testcase_file_path)[1]
    if file_suffix == '.json':
        return load_json_file(testcase_file_path)
    elif file_suffix in ['.yaml', '.yml']:
        return load_yaml_file(testcase_file_path)
    else:
        # '' or other suffix
        return []

def load_folder_files(folder_path, file_type, recursive=False):
    """ load folder path, return all files in list format.
    @param
        folder_path: specified folder path to load
        file_type: "test" or "api"
        recursive: if True, will load files recursively
    """
    file_list = []

    for dirpath, dirnames, filenames in os.walk(folder_path):
        filenames_list = []

        for filename in filenames:

            if not filename.endswith(('.yml', '.yaml', '.json')):
                continue

            if file_type == "api" and not filename.startswith(('api.', 'api-')):
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

def match_expected(value, expected, comparator="eq", check_item=""):
    """ check if value matches expected value.
    @param value: actual value that get from response.
    @param expected: expected result described in testcase
    @param comparator: compare method
    @param check_item: check item name
    """
    try:
        if value is None or expected is None:
            assert comparator in ["is", "eq", "equals", "=="]
            assert value is None
            assert expected is None

        if comparator in ["eq", "equals", "=="]:
            assert value == expected
        elif comparator in ["lt", "less_than"]:
            assert value < expected
        elif comparator in ["le", "less_than_or_equals"]:
            assert value <= expected
        elif comparator in ["gt", "greater_than"]:
            assert value > expected
        elif comparator in ["ge", "greater_than_or_equals"]:
            assert value >= expected
        elif comparator in ["ne", "not_equals"]:
            assert value != expected
        elif comparator in ["str_eq", "string_equals"]:
            assert str(value) == str(expected)
        elif comparator in ["len_eq", "length_equals", "count_eq"]:
            assert isinstance(expected, int)
            assert len(value) == expected
        elif comparator in ["len_gt", "count_gt", "length_greater_than", "count_greater_than"]:
            assert isinstance(expected, int)
            assert len(value) > expected
        elif comparator in ["len_ge", "count_ge", "length_greater_than_or_equals", \
            "count_greater_than_or_equals"]:
            assert isinstance(expected, int)
            assert len(value) >= expected
        elif comparator in ["len_lt", "count_lt", "length_less_than", "count_less_than"]:
            assert isinstance(expected, int)
            assert len(value) < expected
        elif comparator in ["len_le", "count_le", "length_less_than_or_equals", \
            "count_less_than_or_equals"]:
            assert isinstance(expected, int)
            assert len(value) <= expected
        elif comparator in ["contains"]:
            assert isinstance(value, (list, tuple, dict, string_type))
            assert expected in value
        elif comparator in ["contained_by"]:
            assert isinstance(expected, (list, tuple, dict, string_type))
            assert value in expected
        elif comparator in ["type"]:
            assert isinstance(value, expected)
        elif comparator in ["regex"]:
            assert isinstance(expected, string_type)
            assert isinstance(value, string_type)
            assert re.match(expected, value)
        elif comparator in ["startswith"]:
            assert str(value).startswith(str(expected))
        elif comparator in ["endswith"]:
            assert str(value).endswith(str(expected))
        else:
            raise exception.ParamsError("comparator not supported!")

        return True

    except (AssertionError, TypeError):
        err_msg = "\n".join([
            "check item name: %s;" % check_item,
            "check item value: %s (%s);" % (value, type(value).__name__),
            "comparator: %s;" % comparator,
            "expected value: %s (%s)." % (expected, type(expected).__name__)
        ])
        raise exception.ValidationError(err_msg)

def deep_update_dict(origin_dict, override_dict):
    """ update origin dict with override dict recursively
    e.g. origin_dict = {'a': 1, 'b': {'c': 2, 'd': 4}}
         override_dict = {'b': {'c': 3}}
    return: {'a': 1, 'b': {'c': 3, 'd': 4}}
    """
    for key, val in override_dict.items():
        if isinstance(val, dict):
            tmp = deep_update_dict(origin_dict.get(key, {}), val)
            origin_dict[key] = tmp
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

def lower_dict_key(origin_dict, depth=1):
    """ convert dict key to lower case, with depth control supported.
    """
    new_dict = {}

    for key, value in origin_dict.items():
        if depth > 2:
            new_dict[key] = value
            continue

        if isinstance(value, dict):
            value = lower_dict_key(value, depth+1)

        new_dict[key.lower()] = value

    return new_dict

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

def override_variables_binds(variable_binds, new_mapping):
    """ convert variable_binds in testcase to ordered mapping, with new_mapping overrided
    """
    return update_ordered_dict(
        convert_to_order_dict(variable_binds),
        new_mapping
    )

def print_output(output):
    if not output:
        return

    print("\n================== Output ==================")
    print('{:<16}:  {:<}'.format("Variable", "Value"))
    print('{:<16}:  {:<}'.format("--------", "-----"))

    for variable, value in output.items():
        print('{:<16}:  {:<}'.format(
            variable.encode("utf-8"), value.encode("utf-8")))

    print("============================================\n")
