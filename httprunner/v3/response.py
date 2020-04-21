from typing import Dict, Text, Any, NoReturn

import jmespath
import requests
from loguru import logger

from httprunner.v3.exceptions import ValidationFailure, ParamsError
from httprunner.v3.parser import parse_data, parse_string_value, get_mapping_function
from httprunner.v3.schema import VariablesMapping, Validators, FunctionsMapping
from httprunner.v3.validator import uniform_validator


class ResponseObject(object):

    def __init__(self, resp_obj: requests.Response):
        """ initialize with a requests.Response object

        Args:
            resp_obj (instance): requests.Response instance

        """
        self.resp_obj = resp_obj
        self.resp_obj_meta = {
            "status_code": resp_obj.status_code,
            "headers": resp_obj.headers,
            "body": resp_obj.json()
        }
        self.validation_results = {}

    def __getattr__(self, key):
        try:
            if key == "json":
                value = self.resp_obj.json()
            elif key == "cookies":
                value = self.resp_obj.cookies.get_dict()
            else:
                value = getattr(self.resp_obj, key)

            self.__dict__[key] = value
            return value
        except AttributeError:
            err_msg = f"ResponseObject does not have attribute: {key}"
            logger.error(err_msg)
            raise ParamsError(err_msg)

    def extract(self, extractors: Dict[Text, Text]) -> Dict[Text, Any]:
        if not extractors:
            return {}

        extract_mapping = {}
        for key, field in extractors.items():
            field_value = jmespath.search(field, self.resp_obj_meta)
            extract_mapping[key] = field_value

        logger.info(f"extract mapping: {extract_mapping}")
        return extract_mapping

    def validate(self,
                 validators: Validators,
                 variables_mapping: VariablesMapping = None,
                 functions_mapping: FunctionsMapping = None) -> NoReturn:

        self.validation_results = {}
        if not validators:
            return

        validate_pass = True
        failures = []

        for v in validators:

            if "validate_extractor" not in self.validation_results:
                self.validation_results["validate_extractor"] = []

            u_validator = uniform_validator(v)

            # check item
            check_item = u_validator["check"]
            check_value = jmespath.search(check_item, self.resp_obj_meta)
            check_value = parse_string_value(check_value)

            # comparator
            assert_method = u_validator["assert"]
            assert_func = get_mapping_function(assert_method, functions_mapping)

            # expect item
            expect_item = u_validator["expect"]
            # parse expected value with config/teststep/extracted variables
            expect_value = parse_data(expect_item, variables_mapping, functions_mapping)

            validate_msg = f"assert {check_item} {assert_method} {expect_value}({type(expect_value).__name__})"

            validator_dict = {
                "comparator": assert_method,
                "check": check_item,
                "check_value": check_value,
                "expect": expect_item,
                "expect_value": expect_value
            }

            try:
                assert_func(check_value, expect_value)
                validate_msg += "\t==> pass"
                logger.info(validate_msg)
                validator_dict["check_result"] = "pass"
            except AssertionError:
                validate_pass = False
                validator_dict["check_result"] = "fail"
                validate_msg += "\t==> fail"
                validate_msg += f"\n" \
                                f"check_item: {check_item}\n" \
                                f"check_value: {check_value}({type(check_value).__name__})\n" \
                                f"assert_method: {assert_method}\n" \
                                f"expect_value: {expect_value}({type(expect_value).__name__})"
                logger.error(validate_msg)
                failures.append(validate_msg)

            self.validation_results["validate_extractor"].append(validator_dict)

        if not validate_pass:
            failures_string = "\n".join([failure for failure in failures])
            raise ValidationFailure(failures_string)
