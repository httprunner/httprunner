import io
import json
import os
import platform

import jsonschema

from httprunner import exceptions, logger

schemas_root_dir = os.path.join(os.path.dirname(__file__), "schemas")
common_schema_path = os.path.join(schemas_root_dir, "common.schema.json")
api_schema_path = os.path.join(schemas_root_dir, "api.schema.json")
testcase_schema_v1_path = os.path.join(schemas_root_dir, "testcase.schema.v1.json")
testcase_schema_v2_path = os.path.join(schemas_root_dir, "testcase.schema.v2.json")
testsuite_schema_v1_path = os.path.join(schemas_root_dir, "testsuite.schema.v1.json")
testsuite_schema_v2_path = os.path.join(schemas_root_dir, "testsuite.schema.v2.json")

with io.open(api_schema_path, encoding='utf-8') as f:
    api_schema = json.load(f)

with io.open(common_schema_path, encoding='utf-8') as f:
    if platform.system() == "Windows":
        absolute_base_path = 'file:///' + os.path.abspath(schemas_root_dir).replace("\\", "/") + '/'
    else:
        # Linux, Darwin
        absolute_base_path = "file://" + os.path.abspath(schemas_root_dir) + "/"

    common_schema = json.load(f)
    resolver = jsonschema.RefResolver(absolute_base_path, common_schema)

with io.open(testcase_schema_v1_path, encoding='utf-8') as f:
    testcase_schema_v1 = json.load(f)

with io.open(testcase_schema_v2_path, encoding='utf-8') as f:
    testcase_schema_v2 = json.load(f)

with io.open(testsuite_schema_v1_path, encoding='utf-8') as f:
    testsuite_schema_v1 = json.load(f)

with io.open(testsuite_schema_v2_path, encoding='utf-8') as f:
    testsuite_schema_v2 = json.load(f)


class JsonSchemaChecker(object):

    @staticmethod
    def validate_format(content, scheme):
        """ check api/testcase/testsuite format if valid
        """
        try:
            jsonschema.validate(content, scheme, resolver=resolver)
        except jsonschema.exceptions.ValidationError as ex:
            logger.log_error(str(ex))
            raise exceptions.FileFormatError

        return True

    @staticmethod
    def validate_api_format(content):
        """ check api format if valid
        """
        return JsonSchemaChecker.validate_format(content, api_schema)

    @staticmethod
    def validate_testcase_v1_format(content):
        """ check testcase format v1 if valid
        """
        return JsonSchemaChecker.validate_format(content, testcase_schema_v1)

    @staticmethod
    def validate_testcase_v2_format(content):
        """ check testcase format v2 if valid
        """
        return JsonSchemaChecker.validate_format(content, testcase_schema_v2)

    @staticmethod
    def validate_testsuite_v1_format(content):
        """ check testsuite format v1 if valid
        """
        return JsonSchemaChecker.validate_format(content, testsuite_schema_v1)

    @staticmethod
    def validate_testsuite_v2_format(content):
        """ check testsuite format v2 if valid
        """
        return JsonSchemaChecker.validate_format(content, testsuite_schema_v2)


def is_test_path(path):
    """ check if path is valid json/yaml file path or a existed directory.

    Args:
        path (str/list/tuple): file path/directory or file path list.

    Returns:
        bool: True if path is valid file path or path list, otherwise False.

    """
    if not isinstance(path, (str, list, tuple)):
        return False

    elif isinstance(path, (list, tuple)):
        for p in path:
            if not is_test_path(p):
                return False

        return True

    else:
        # path is string
        if not os.path.exists(path):
            return False

        # path exists
        if os.path.isfile(path):
            # path is a file
            file_suffix = os.path.splitext(path)[1].lower()
            if file_suffix not in ['.json', '.yaml', '.yml']:
                # path is not json/yaml file
                return False
            else:
                return True
        elif os.path.isdir(path):
            # path is a directory
            return True
        else:
            # path is neither a folder nor a file, maybe a symbol link or something else
            return False


def is_test_content(data_structure):
    """ check if data_structure is apis/testcases/testsuites.

    Args:
        data_structure (dict): should include keys, apis or testcases or testsuites

    Returns:
        bool: True if data_structure is valid apis/testcases/testsuites, otherwise False.

    """
    if not isinstance(data_structure, dict):
        return False

    if "apis" in data_structure:
        # maybe a group of api content
        apis = data_structure["apis"]
        if not isinstance(apis, list):
            return False

        for item in apis:
            is_testcase = False
            try:
                JsonSchemaChecker.validate_api_format(item)
                is_testcase = True
            except exceptions.FileFormatError:
                pass

            if not is_testcase:
                return False

        return True

    elif "testcases" in data_structure:
        # maybe a testsuite, containing a group of testcases
        testcases = data_structure["testcases"]
        if not isinstance(testcases, list):
            return False

        for item in testcases:
            is_testcase = False
            try:
                JsonSchemaChecker.validate_testcase_v2_format(item)
                is_testcase = True
            except exceptions.FileFormatError:
                pass

            try:
                JsonSchemaChecker.validate_testcase_v2_format(item)
                is_testcase = True
            except exceptions.FileFormatError:
                pass

            if not is_testcase:
                return False

        return True

    elif "testsuites" in data_structure:
        # maybe a group of testsuites
        testsuites = data_structure["testsuites"]
        if not isinstance(testsuites, list):
            return False

        for item in testsuites:
            is_testcase = False
            try:
                JsonSchemaChecker.validate_testsuite_v1_format(item)
                is_testcase = True
            except exceptions.FileFormatError:
                pass

            try:
                JsonSchemaChecker.validate_testsuite_v2_format(item)
                is_testcase = True
            except exceptions.FileFormatError:
                pass

            if not is_testcase:
                return False

        return True

    else:
        return False
