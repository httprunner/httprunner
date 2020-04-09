import os

from loguru import logger
from pydantic import ValidationError

from httprunner import exceptions
from httprunner.schema import Api, TestCase, TestSuite


class JsonSchemaChecker(object):

    @staticmethod
    def validate_api_format(content):
        """ check api format if valid
        """
        try:
            Api.parse_obj(content)
        except ValidationError as ex:
            logger.error(ex)
            raise exceptions.FileFormatError(ex)

    @staticmethod
    def validate_testcase_format(content):
        """ check testcase format if valid
        """
        try:
            TestCase.parse_obj(content)
        except ValidationError as ex:
            logger.error(ex)
            raise exceptions.FileFormatError(ex)

    @staticmethod
    def validate_testsuite_format(content):
        """ check testsuite format if valid
        """
        try:
            TestSuite.parse_obj(content)
        except ValidationError as ex:
            logger.error(ex)
            raise exceptions.FileFormatError(ex)


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
                JsonSchemaChecker.validate_testcase_format(item)
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
                JsonSchemaChecker.validate_testsuite_format(item)
                is_testcase = True
            except exceptions.FileFormatError:
                pass

            if not is_testcase:
                return False

        return True

    else:
        return False
