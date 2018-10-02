# encoding: utf-8
import os
import types


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
                    teststep1,
                    {   # teststep2
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

            testcase_dict
            or
            [
                testcase_dict_1,
                testcase_dict_2
            ]
    Returns:
        bool: True if data_structure is valid testcase(s), otherwise False.

    """
    if not isinstance(data_structure, list):
        return is_testcase(data_structure)

    for item in data_structure:
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
##   validate varibles and functions
###############################################################################


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
