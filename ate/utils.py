import codecs
import fnmatch
import hashlib
import hmac
import importlib
import json
import os.path
import random
import re
import string
import types

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

def load_foler_files(folder_path, match_filter_list=["*"]):
    """ load folder path, return all files in list format.
    """
    file_list = []

    for dirpath, dirnames, filenames in os.walk(folder_path):
        filenames_list = []
        for match_filter in match_filter_list:
            filenames_list.extend(fnmatch.filter(filenames, match_filter))

        for filename in filenames_list:
            file_path = os.path.join(dirpath, filename)
            file_list.append(file_path)

    return file_list

def load_testcases_by_path(path):
    """ load testcases from file path
    @param path
        path could be in several type:
            - absolute/relative file path
            - absolute/relative folder path
            - list/set container with file(s) and/or folder(s)
    @return testcase sets list, each testset is corresponding to a file
        [
            {"name": "desc1", "config": {}, "testcases": [testcase11, testcase12]},
            {"name": "desc2", "config": {}, "testcases": [testcase21, testcase22, testcase23]},
        ]
    """
    if isinstance(path, (list, set)):
        testsets_list = []

        for file_path in set(path):
            _testsets_list = load_testcases_by_path(file_path)
            testsets_list.extend(_testsets_list)

        return testsets_list

    if not os.path.isabs(path):
        path = os.path.join(os.getcwd(), path)

    if os.path.isdir(path):
        files_list = load_foler_files(path, ["*.yml", "*.yaml", "*.json"])
        return load_testcases_by_path(files_list)

    elif os.path.isfile(path):
        testset = {
            "name": "",
            "config": {
                "path": path
            },
            "testcases": []
        }
        testcases_list = load_testcases(path)

        for item in testcases_list:
            for key in item:
                if key == "config":
                    testset["config"].update(item["config"])
                    testset["name"] = item["config"].get("name", "")
                elif key == "test":
                    testset["testcases"].append(item["test"])

        return [testset] if testset["testcases"] else []

    else:
        return []

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

def get_module_functions(module_name):
    """ import module and return filtered functions
    """
    imported = importlib.import_module(module_name)
    module_functions_dict = dict(filter(is_function, vars(imported).items()))
    return module_functions_dict
