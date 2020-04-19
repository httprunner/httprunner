from typing import Dict, Text, Any

import jmespath
import requests
from loguru import logger

from httprunner.v3.exceptions import ParamsError, ValidationFailure
from httprunner.v3.validator import uniform_validator, AssertMethods


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

    def validate(self, validators):

        for v in validators:
            u_validator = uniform_validator(v)
            field = u_validator["check"]
            assert_method = u_validator["assert"]
            expect_value = u_validator["expect"]
            actual_value = jmespath.search(field, self.resp_obj_meta)

            msg = f"assert {field} {assert_method} {expect_value}"

            try:
                assert_func = getattr(AssertMethods, assert_method)
            except AttributeError:
                raise ParamsError(f"Assert Method not supported: {assert_method}")

            try:
                assert_func(actual_value, expect_value)
                msg += " - success"
                logger.info(msg)
            except AssertionError:
                msg += " - fail"
                logger.error(msg)
                raise ValidationFailure(f"assert {field}: {actual_value} {assert_method} {expect_value}")

    def extract(self, extractors: Dict[Text, Text]) -> Dict[Text, Any]:
        if not extractors:
            return {}

        extract_mapping = {}
        for key, field in extractors.items():
            field_value = jmespath.search(field, self.resp_obj_meta)
            extract_mapping[key] = field_value

        logger.info(f"extract mapping: {extract_mapping}")
        return extract_mapping
