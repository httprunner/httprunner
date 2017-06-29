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

def parse_content_with_variables(content, variables_binds):
    """ replace variables with bind value
    """
    # check if content includes ${variable}
    matched = re.match(r"(.*)\$\{(.*)\}(.*)", content)
    if matched:
        # this is a variable, and will replace with its bind value
        variable_name = matched.group(2)
        value = variables_binds.get(variable_name)
        if value is None:
            raise ParamsError(
                "%s is not defined in bind variables!" % variable_name)
        if matched.group(1) or matched.group(3):
            # e.g. /api/users/${uid}
            return re.sub(r"\$\{.*\}", value, content)

        return value

    return content

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
            assert len(value) == len(expected)
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

        return True
    except AssertionError:
        return False
