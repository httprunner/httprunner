import json
import os
import types

import jsonschema

from httprunner import exceptions, logger

schemas_root_dir = os.path.join(os.path.dirname(__file__), "schemas")
common_schema_path = os.path.join(schemas_root_dir, "common.schema.json")
api_schema_path = os.path.join(schemas_root_dir, "api.schema.json")
testcase_schema_v2_path = os.path.join(schemas_root_dir, "testcase.schema.v2.json")
testsuite_schema_path = os.path.join(schemas_root_dir, "testsuite.schema.json")

with open(api_schema_path) as f:
    api_schema = json.load(f)

with open(common_schema_path) as f:
    common_schema = json.load(f)
    resolver = jsonschema.RefResolver("file://{}/".format(os.path.abspath(schemas_root_dir)), common_schema)

with open(testcase_schema_v2_path) as f:
    testcase_schema_v2 = json.load(f)

with open(testsuite_schema_path) as f:
    testsuite_schema = json.load(f)


class JsonSchemaChecker(object):

    @staticmethod
    def validate_api_format(content):

        try:
            jsonschema.validate(content, api_schema, resolver=resolver)
        except jsonschema.exceptions.ValidationError as ex:
            logger.log_error(str(ex))
            raise exceptions.FileFormatError

        return True

    @staticmethod
    def validate_testsuite_format(content):
        try:
            jsonschema.validate(content, testsuite_schema, resolver=resolver)
        except jsonschema.exceptions.ValidationError as ex:
            logger.log_error(str(ex))
            raise exceptions.FileFormatError

        return True


class JsonSchemaV1Checker(JsonSchemaChecker):
    pass


class JsonSchemaV2Checker(JsonSchemaChecker):

    @staticmethod
    def validate_testcase_format(content):
        """ check testcase format v2 if valid
        """
        try:
            jsonschema.validate(content, testcase_schema_v2, resolver=resolver)
        except jsonschema.exceptions.ValidationError as ex:
            logger.log_error(str(ex))
            raise exceptions.FileFormatError

        return True


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
