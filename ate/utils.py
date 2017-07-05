import ast
import hashlib
import json
import os.path
import random
import re
import string
import yaml

from ate.exception import ParamsError

try:
    string_type = basestring
    PYTHON_VERSION = 2
except NameError:
    string_type = str
    PYTHON_VERSION = 3

variable_regexp = re.compile(r"^\$(\w+)$")
function_regexp = re.compile(r"^\$\{(\w+)\(([\$\w =,]*)\)\}$")

def gen_random_string(str_len):
    return ''.join(
        random.choice(string.ascii_letters + string.digits) for _ in range(str_len))

def gen_md5(*str_args):
    return hashlib.md5("".join(str_args).encode('utf-8')).hexdigest()

def handle_req_data(data):

    if PYTHON_VERSION == 3 and isinstance(data, bytes):
        # In Python3, convert bytes to str
        data = data.decode('utf-8')

    if not data:
        return data

    if isinstance(data, str):
        # check if data in str can be converted to dict
        try:
            data = json.loads(data)
        except ValueError:
            pass

    if isinstance(data, dict):
        # sort data in dict with keys, then convert to str
        data = json.dumps(data, sort_keys=True)

    return data

def load_yaml_file(yaml_file):
    with open(yaml_file, 'r+') as stream:
        return yaml.load(stream)

def load_json_file(json_file):
    with open(json_file) as data_file:
        return json.load(data_file)

def load_testcases(testcase_file_path):
    file_suffix = os.path.splitext(testcase_file_path)[1]
    if file_suffix == '.json':
        return load_json_file(testcase_file_path)
    elif file_suffix in ['.yaml', '.yml']:
        return load_yaml_file(testcase_file_path)
    else:
        # '' or other suffix
        raise ParamsError("Bad testcase file name!")

def load_foler_files(folder_path):
    """ load folder path, return all files in list format.
    """
    file_list = []

    for dirpath, dirnames, filenames in os.walk(folder_path):
        for filename in filenames:
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
        files_list = load_foler_files(path)
        return load_testcases_by_path(files_list)

    elif os.path.isfile(path):
        testset = {
            "name": "",
            "config": {},
            "testcases": []
        }
        try:
            testcases_list = load_testcases(path)
        except ParamsError:
            return []

        for item in testcases_list:
            for key in item:
                if key == "config":
                    testset["config"] = item["config"]
                    testset["name"] = item["config"].get("name", "")
                elif key == "test":
                    testset["testcases"].append(item["test"])

        return [testset]

    else:
        return []

def is_variable(content):
    """ check if content is a variable, which is in format $variable
    @param (str) content
    @return (bool) True or False

    e.g. $variable => True
         abc => False
    """
    matched = variable_regexp.match(content)
    return True if matched else False

def parse_variable(content):
    """ parse variable name from string content.
    @param (str) content
    @return (str) variable name

    e.g. $variable => variable
    """
    matched = variable_regexp.match(content)
    return matched.group(1)

def is_functon(content):
    """ check if content is a function, which is in format ${func()}
    @param (str) content
    @return (bool) True or False

    e.g. ${func()} => True
         ${func(5)} => True
         ${func(1, 2)} => True
         ${func(a=1, b=2)} => True
         $abc => False
         abc => False
    """
    matched = function_regexp.match(content)
    return True if matched else False

def parse_string_value(str_value):
    """ parse string to number if possible
    e.g. "123" => 123
         "12.2" => 12.3
         "abc" => "abc"
         "$var" => "$var"
    """
    try:
        return ast.literal_eval(str_value)
    except ValueError:
        return str_value
    except SyntaxError:
        # e.g. $var, ${func}
        return str_value

def parse_function(content):
    """ parse function name and args from string content.
    @param (str) content
    @return (dict) function name and args

    e.g. ${func()} => {'func_name': 'func', 'args': [], 'kwargs': {}}
         ${func(5)} => {'func_name': 'func', 'args': [5], 'kwargs': {}}
         ${func(1, 2)} => {'func_name': 'func', 'args': [1, 2], 'kwargs': {}}
         ${func(a=1, b=2)} => {'func_name': 'func', 'args': [], 'kwargs': {'a': 1, 'b': 2}}
         ${func(1, 2, a=3, b=4)} => {'func_name': 'func', 'args': [1, 2], 'kwargs': {'a':3, 'b':4}}
    """
    function_meta = {
        "args": [],
        "kwargs": {}
    }
    matched = function_regexp.match(content)
    function_meta["func_name"] = matched.group(1)

    args_str = matched.group(2).replace(" ", "")
    if args_str == "":
        return function_meta

    args_list = args_str.split(',')
    for arg in args_list:
        if '=' in arg:
            key, value = arg.split('=')
            function_meta["kwargs"][key] = parse_string_value(value)
        else:
            function_meta["args"].append(parse_string_value(arg))

    return function_meta

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
    stripped_query = query.strip(delimiter)
    if not stripped_query:
        return None

    try:
        for key in stripped_query.split(delimiter):
            if isinstance(json_content, list):
                key = int(key)
            json_content = json_content[key]
    except (KeyError, ValueError, IndexError):
        raise ParamsError("invalid query string in extract_binds!")

    return json_content

def match_expected(value, expected, comparator="eq"):
    """ check if value matches expected value.
    @param value: value that get from response.
    @param expected: expected result described in testcase
    @param comparator: compare method
    """
    try:
        if comparator in ["eq", "equals", "=="]:
            assert value == expected
        elif comparator in ["str_eq", "string_equals"]:
            assert str(value) == str(expected)
        elif comparator in ["ne", "not_equals"]:
            assert value != expected
        elif comparator in ["len_eq", "length_equal", "count_eq"]:
            assert len(value) == expected
        elif comparator in ["len_gt", "count_gt", "length_greater_than", "count_greater_than"]:
            assert len(value) > expected
        elif comparator in ["len_ge", "count_ge", "length_greater_than_or_equals", \
            "count_greater_than_or_equals"]:
            assert len(value) >= expected
        elif comparator in ["len_lt", "count_lt", "length_less_than", "count_less_than"]:
            assert len(value) < expected
        elif comparator in ["len_le", "count_le", "length_less_than_or_equals", \
            "count_less_than_or_equals"]:
            assert len(value) <= expected
        elif comparator in ["lt", "less_than"]:
            assert value < expected
        elif comparator in ["le", "less_than_or_equals"]:
            assert value <= expected
        elif comparator in ["gt", "greater_than"]:
            assert value > expected
        elif comparator in ["ge", "greater_than_or_equals"]:
            assert value >= expected
        elif comparator in ["contains"]:
            assert expected in value
        elif comparator in ["contained_by"]:
            assert value in expected
        elif comparator in ["regex"]:
            assert re.match(expected, value)
        elif comparator in ["str_len", "string_length"]:
            assert len(value) == int(expected)
        elif comparator in ["startswith"]:
            assert str(value).startswith(str(expected))
        else:
            raise ParamsError("comparator not supported!")

        return True
    except AssertionError:
        return False

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
