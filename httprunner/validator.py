# encoding: utf-8
import collections
import io
import json
import os
import types

from httprunner import exceptions, logger

""" validate data format
TODO: refactor with JSON schema validate
"""


def is_testcase(data_structure):
    """ check if data_structure is a testcase.

    Args:
        data_structure (dict): testcase should always be in the following data structure:

            {
                "config": {
                    "name": "desc1",
                    "variables": [],    # optional
                    "request": {}       # optional
                },
                "teststeps": [
                    test_dict1,
                    {   # test_dict2
                        'name': 'test step desc2',
                        'variables': [],    # optional
                        'extract': [],      # optional
                        'validate': [],
                        'request': {},
                        'function_meta': {}
                    }
                ]
            }

    Returns:
        bool: True if data_structure is valid testcase, otherwise False.

    """
    # TODO: replace with JSON schema validation
    if not isinstance(data_structure, dict):
        return False

    if "teststeps" not in data_structure:
        return False

    if not isinstance(data_structure["teststeps"], list):
        return False

    return True


def is_testcases(data_structure):
    """ check if data_structure is testcase or testcases list.

    Args:
        data_structure (dict): testcase(s) should always be in the following data structure:
            {
                "project_mapping": {
                    "PWD": "XXXXX",
                    "functions": {},
                    "env": {}
                },
                "testcases": [
                    {   # testcase data structure
                        "config": {
                            "name": "desc1",
                            "path": "testcase1_path",
                            "variables": [],                    # optional
                        },
                        "teststeps": [
                            # test data structure
                            {
                                'name': 'test step desc1',
                                'variables': [],    # optional
                                'extract': [],      # optional
                                'validate': [],
                                'request': {}
                            },
                            test_dict_2   # another test dict
                        ]
                    },
                    testcase_dict_2     # another testcase dict
                ]
            }

    Returns:
        bool: True if data_structure is valid testcase(s), otherwise False.

    """
    if not isinstance(data_structure, dict):
        return False

    if "testcases" not in data_structure:
        return False

    testcases = data_structure["testcases"]
    if not isinstance(testcases, list):
        return False

    for item in testcases:
        if not is_testcase(item):
            return False

    return True


def is_testcase_path(path):
    """ check if path is testcase path or path list.

    Args:
        path (str/list): file path or file path list.

    Returns:
        bool: True if path is valid file path or path list, otherwise False.

    """
    if not isinstance(path, (str, list)):
        return False

    if isinstance(path, list):
        for p in path:
            if not is_testcase_path(p):
                return False

    if isinstance(path, str):
        if not os.path.exists(path):
            return False

    return True


###############################################################################
##   testcase validator utils
###############################################################################

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


def uniform_validator(validator):
    """ unify validator

    Args:
        validator (dict): validator maybe in two formats:

            format1: this is kept for compatiblity with the previous versions.
                {"check": "status_code", "comparator": "eq", "expect": 201}
                {"check": "$resp_body_success", "comparator": "eq", "expect": True}
            format2: recommended new version, {comparator: [check_item, expected_value]}
                {'eq': ['status_code', 201]}
                {'eq': ['$resp_body_success', True]}

    Returns
        dict: validator info

            {
                "check": "status_code",
                "expect": 201,
                "comparator": "equals"
            }

    """
    if not isinstance(validator, dict):
        raise exceptions.ParamsError("invalid validator: {}".format(validator))

    if "check" in validator and "expect" in validator:
        # format1
        check_item = validator["check"]
        expect_value = validator["expect"]
        comparator = validator.get("comparator", "eq")

    elif len(validator) == 1:
        # format2
        comparator = list(validator.keys())[0]
        compare_values = validator[comparator]

        if not isinstance(compare_values, list) or len(compare_values) != 2:
            raise exceptions.ParamsError("invalid validator: {}".format(validator))

        check_item, expect_value = compare_values

    else:
        raise exceptions.ParamsError("invalid validator: {}".format(validator))

    # uniform comparator, e.g. lt => less_than, eq => equals
    comparator = get_uniform_comparator(comparator)

    return {
        "check": check_item,
        "expect": expect_value,
        "comparator": comparator
    }


def _convert_validators_to_mapping(validators):
    """ convert validators list to mapping.

    Args:
        validators (list): validators in list

    Returns:
        dict: validators mapping, use (check, comparator) as key.

    Examples:
        >>> validators = [
                {"check": "v1", "expect": 201, "comparator": "eq"},
                {"check": {"b": 1}, "expect": 200, "comparator": "eq"}
            ]
        >>> _convert_validators_to_mapping(validators)
            {
                ("v1", "eq"): {"check": "v1", "expect": 201, "comparator": "eq"},
                ('{"b": 1}', "eq"): {"check": {"b": 1}, "expect": 200, "comparator": "eq"}
            }

    """
    validators_mapping = {}

    for validator in validators:
        if not isinstance(validator["check"], collections.Hashable):
            check = json.dumps(validator["check"])
        else:
            check = validator["check"]

        key = (check, validator["comparator"])
        validators_mapping[key] = validator

    return validators_mapping


def extend_validators(raw_validators, override_validators):
    """ extend raw_validators with override_validators.
        override_validators will merge and override raw_validators.

    Args:
        raw_validators (dict):
        override_validators (dict):

    Returns:
        list: extended validators

    Examples:
        >>> raw_validators = [{'eq': ['v1', 200]}, {"check": "s2", "expect": 16, "comparator": "len_eq"}]
        >>> override_validators = [{"check": "v1", "expect": 201}, {'len_eq': ['s3', 12]}]
        >>> extend_validators(raw_validators, override_validators)
            [
                {"check": "v1", "expect": 201, "comparator": "eq"},
                {"check": "s2", "expect": 16, "comparator": "len_eq"},
                {"check": "s3", "expect": 12, "comparator": "len_eq"}
            ]

    """

    if not raw_validators:
        return override_validators

    elif not override_validators:
        return raw_validators

    else:
        def_validators_mapping = _convert_validators_to_mapping(raw_validators)
        ref_validators_mapping = _convert_validators_to_mapping(override_validators)

        def_validators_mapping.update(ref_validators_mapping)
        return list(def_validators_mapping.values())


###############################################################################
##   validate varibles and functions
###############################################################################


def is_function(item):
    """ Takes item object, returns True if it is a function.
    """
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


def validate_json_file(file_list):
    """ validate JSON testcase format
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
