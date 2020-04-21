from typing import Dict, Text, Any, NoReturn

import jmespath
import requests
from loguru import logger

from httprunner.v3.exceptions import ValidationFailure
from httprunner.v3.parser import parse_data, parse_string_value, get_mapping_function
from httprunner.v3.schema import VariablesMapping, Validators, FunctionsMapping
from httprunner.v3.validator import uniform_validator


class ResponseObject(object):

    def __init__(self, resp_obj: requests.Response):
        """ initialize with a requests.Response object

        Args:
            resp_obj (instance): requests.Response instance

        """
        self.resp_obj_meta = {
            "status_code": resp_obj.status_code,
            "headers": resp_obj.headers,
            "body": resp_obj.json()
        }

    def validate(self,
                 validators: Validators,
                 variables_mapping: VariablesMapping = None,
                 functions_mapping: FunctionsMapping = None) -> NoReturn:

        for v in validators:
            u_validator = uniform_validator(v)
            field = u_validator["check"]
            assert_method = u_validator["assert"]
            expect_value = u_validator["expect"]
            actual_value = jmespath.search(field, self.resp_obj_meta)

            msg = f"assert {field} {assert_method} {expect_value}"

            assert_func = get_mapping_function(assert_method, functions_mapping)
            actual_value = parse_string_value(actual_value)
            # parse expected value with config/teststep/extracted variables
            expect_value = parse_data(expect_value, variables_mapping, functions_mapping)

            try:
                assert_func(actual_value, expect_value)
                msg += " - success"
                logger.info(msg)
            except AssertionError:
                msg += " - fail"
                logger.error(msg)
                actual_type = type(actual_value).__name__
                expect_type = type(expect_value).__name__
                raise ValidationFailure(f"assert {field}: {actual_value}({actual_type}) {assert_method} {expect_value}({expect_type})")

    def extract(self, extractors: Dict[Text, Text]) -> Dict[Text, Any]:
        if not extractors:
            return {}

        extract_mapping = {}
        for key, field in extractors.items():
            field_value = jmespath.search(field, self.resp_obj_meta)
            extract_mapping[key] = field_value

        logger.info(f"extract mapping: {extract_mapping}")
        return extract_mapping
