# encoding: utf-8

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
