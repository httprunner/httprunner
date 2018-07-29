# encoding: utf-8

import collections
import copy
import hashlib
import hmac
import imp
import importlib
import io
import json
import os.path
import random
import string
import types
from datetime import datetime

from httprunner import exceptions, logger, parser
from httprunner.compat import (OrderedDict, basestring, builtin_str, is_py2,
                               is_py3, numeric_types, str)
from requests.structures import CaseInsensitiveDict

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


def query_json(json_content, query, delimiter='.'):
    """ Do an xpath-like query with json_content.
    @param (dict/list/string) json_content
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
        "person.name.first_name.0"  =>  "L"
        "person.cities.0"         =>  "Guangzhou"
    @return queried result
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


def substitute_variables_with_mapping(content, mapping):
    """ substitute variables in content with mapping
    e.g.
    @params
        content = {
            'request': {
                'url': '/api/users/$uid',
                'headers': {'token': '$token'}
            }
        }
        mapping = {"$uid": 1000}
    @return
        {
            'request': {
                'url': '/api/users/1000',
                'headers': {'token': '$token'}
            }
        }
    """
    # TODO: refactor type check
    if isinstance(content, bool):
        return content

    if isinstance(content, (numeric_types, type)):
        return content

    if not content:
        return content

    if isinstance(content, (list, set, tuple)):
        return [
            substitute_variables_with_mapping(item, mapping)
            for item in content
        ]

    if isinstance(content, dict):
        substituted_data = {}
        for key, value in content.items():
            eval_key = substitute_variables_with_mapping(key, mapping)
            eval_value = substitute_variables_with_mapping(value, mapping)
            substituted_data[eval_key] = eval_value

        return substituted_data

    # content is in string format here
    for var, value in mapping.items():
        if content == var:
            # content is a variable
            content = value
        else:
            if not isinstance(value, str):
                value = builtin_str(value)
            content = content.replace(var, value)

    return content


def _get_validators_mapping(validators):
    """ get validators mapping from api or test validators
    @param (list) validators:
        [
            {"check": "v1", "expect": 201, "comparator": "eq"},
            {"check": {"b": 1}, "expect": 200, "comparator": "eq"}
        ]
    @return
        {
            ("v1", "eq"): {"check": "v1", "expect": 201, "comparator": "eq"},
            ('{"b": 1}', "eq"): {"check": {"b": 1}, "expect": 200, "comparator": "eq"}
        }
    """
    validators_mapping = {}

    for validator in validators:
        validator = parser.parse_validator(validator)

        if not isinstance(validator["check"], collections.Hashable):
            check = json.dumps(validator["check"])
        else:
            check = validator["check"]

        key = (check, validator["comparator"])
        validators_mapping[key] = validator

    return validators_mapping


def _merge_validator(def_validators, current_validators):
    """ merge def_validators with current_validators
    @params:
        def_validators: [{'eq': ['v1', 200]}, {"check": "s2", "expect": 16, "comparator": "len_eq"}]
        current_validators: [{"check": "v1", "expect": 201}, {'len_eq': ['s3', 12]}]
    @return:
        [
            {"check": "v1", "expect": 201, "comparator": "eq"},
            {"check": "s2", "expect": 16, "comparator": "len_eq"},
            {"check": "s3", "expect": 12, "comparator": "len_eq"}
        ]
    """
    if not def_validators:
        return current_validators

    elif not current_validators:
        return def_validators

    else:
        api_validators_mapping = _get_validators_mapping(def_validators)
        test_validators_mapping = _get_validators_mapping(current_validators)

        api_validators_mapping.update(test_validators_mapping)
        return list(api_validators_mapping.values())


def _merge_extractor(def_extrators, current_extractors):
    """ merge def_extrators with current_extractors
    @params:
        def_extrators: [{"var1": "val1"}, {"var2": "val2"}]
        current_extractors: [{"var1": "val111"}, {"var3": "val3"}]
    @return:
        [
            {"var1": "val111"},
            {"var2": "val2"},
            {"var3": "val3"}
        ]
    """
    if not def_extrators:
        return current_extractors

    elif not current_extractors:
        return def_extrators

    else:
        extractor_dict = OrderedDict()
        for api_extrator in def_extrators:
            if len(api_extrator) != 1:
                logger.log_warning("incorrect extractor: {}".format(api_extrator))
                continue

            var_name = list(api_extrator.keys())[0]
            extractor_dict[var_name] = api_extrator[var_name]

        for test_extrator in current_extractors:
            if len(test_extrator) != 1:
                logger.log_warning("incorrect extractor: {}".format(test_extrator))
                continue

            var_name = list(test_extrator.keys())[0]
            extractor_dict[var_name] = test_extrator[var_name]

        extractor_list = []
        for key, value in extractor_dict.items():
            extractor_list.append({key: value})

        return extractor_list


def _override_block(def_block, current_block):
    """ override def_block with current_block
    @param def_block:
        {
            "name": "get token",
            "request": {...},
            "validate": [{'eq': ['status_code', 200]}]
        }
    @param current_block:
        {
            "name": "get token",
            "extract": [{"token": "content.token"}],
            "validate": [{'eq': ['status_code', 201]}, {'len_eq': ['content.token', 16]}]
        }
    @return
        {
            "name": "get token",
            "request": {...},
            "extract": [{"token": "content.token"}],
            "validate": [{'eq': ['status_code', 201]}, {'len_eq': ['content.token', 16]}]
        }
    """
    def_validators = def_block.get("validate") or def_block.get("validators", [])
    current_validators = current_block.get("validate") or current_block.get("validators", [])

    def_extrators = def_block.get("extract") \
        or def_block.get("extractors") \
        or def_block.get("extract_binds", [])
    current_extractors = current_block.get("extract") \
        or current_block.get("extractors") \
        or current_block.get("extract_binds", [])

    current_block.update(def_block)
    current_block["validate"] = _merge_validator(
        def_validators,
        current_validators
    )
    current_block["extract"] = _merge_extractor(
        def_extrators,
        current_extractors
    )


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
    if is_py3:
        imported_module = importlib.machinery.SourceFileLoader(
            'module_name', file_path).load_module()
    elif is_py2:
        imported_module = imp.load_source('module_name', file_path)
    else:
        raise RuntimeError("Neither Python 3 nor Python 2.")

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
            raise exceptions.FunctionNotFound(err_msg)
        else:
            raise exceptions.VariableNotFound(err_msg)

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
    new_ordered_dict = copy.copy(ordered_dict)
    for var, value in override_mapping.items():
        new_ordered_dict.update({var: value})

    return new_ordered_dict

def override_variables_binds(variables, new_mapping):
    """ convert variables in testcase to ordered mapping, with new_mapping overrided
    """
    if isinstance(variables, list):
        variables_ordered_dict = convert_to_order_dict(variables)
    elif isinstance(variables, (OrderedDict, dict)):
        variables_ordered_dict = variables
    else:
        raise exceptions.ParamsError("variables error!")

    return update_ordered_dict(
        variables_ordered_dict,
        new_mapping
    )

def print_output(outputs):

    if not outputs:
        return

    content = "\n================== Variables & Output ==================\n"
    content += '{:<6} | {:<16} :  {:<}\n'.format("Type", "Variable", "Value")
    content += '{:<6} | {:<16} :  {:<}\n'.format("-" * 6, "-" * 16, "-" * 27)

    def prepare_content(var_type, in_out):
        content = ""
        for variable, value in in_out.items():

            if is_py2:
                if isinstance(variable, unicode):
                    variable = variable.encode("utf-8")
                if isinstance(value, unicode):
                    value = value.encode("utf-8")

            content += '{:<6} | {:<16} :  {:<}\n'.format(var_type, variable, value)

        return content

    for output in outputs:
        _in = output["in"]
        _out = output["out"]

        if not _out:
            continue

        content += prepare_content("Var", _in)
        content += "\n"
        content += prepare_content("Out", _out)
        content += "-" * 56 + "\n"

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


def validate_json_file(file_list):
    """ validate JSON testset format
    """
    for json_file in set(file_list):
        if not json_file.endswith(".json"):
            logger.log_warning("Only JSON file format can be validated, skip: {}".format(json_file))
            continue

        logger.color_print("Start to validate JSON file: {}".format(json_file), "GREEN")

        with io.open(json_file) as stream:
            try:
                json.load(stream)
            except ValueError as e:
                raise SystemExit(e)

        print("OK")

def prettify_json_file(file_list):
    """ prettify JSON testset format
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

def get_python2_retire_msg():
    retire_day = datetime(2020, 1, 1)
    today = datetime.now()
    left_days = (retire_day - today).days

    if left_days > 0:
        retire_msg = "Python 2 will retire in {} days, why not move to Python 3?".format(left_days)
    else:
        retire_msg = "Python 2 has been retired, you should move to Python 3."

    return retire_msg
