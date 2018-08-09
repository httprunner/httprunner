# encoding: utf-8
import types

""" validate data format
TODO: refactor with JSON schema validate
"""

def is_testcase(data_structure):
    """ check if data_structure is a testcase
    testcase should always be in the following data structure:
        {
            "name": "desc1",
            "config": {},
            "api": {},
            "testcases": [testcase11, testcase12]
        }
    """
    if not isinstance(data_structure, dict):
        return False

    if "name" not in data_structure or "testcases" not in data_structure:
        return False

    if not isinstance(data_structure["testcases"], list):
        return False

    return True

def is_testcases(data_structure):
    """ check if data_structure is testcase or testcases list
    testsets should always be in the following data structure:
        testset_dict
        or
        [
            testset_dict_1,
            testset_dict_2
        ]
    """
    if not isinstance(data_structure, list):
        return is_testcase(data_structure)

    for item in data_structure:
        if not is_testcase(item):
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

