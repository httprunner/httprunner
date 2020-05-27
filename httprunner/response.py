from typing import Dict, Text, Any, NoReturn

import jmespath
import requests
from loguru import logger

from httprunner.exceptions import ValidationFailure, ParamsError
from httprunner.parser import parse_data, parse_string_value, get_mapping_function
from httprunner.schema import VariablesMapping, Validators, FunctionsMapping


def get_uniform_comparator(comparator: Text):
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
    elif comparator in [
        "len_gt",
        "count_gt",
        "length_greater_than",
        "count_greater_than",
    ]:
        return "length_greater_than"
    elif comparator in [
        "len_ge",
        "count_ge",
        "length_greater_than_or_equals",
        "count_greater_than_or_equals",
    ]:
        return "length_greater_than_or_equals"
    elif comparator in ["len_lt", "count_lt", "length_less_than", "count_less_than"]:
        return "length_less_than"
    elif comparator in [
        "len_le",
        "count_le",
        "length_less_than_or_equals",
        "count_less_than_or_equals",
    ]:
        return "length_less_than_or_equals"
    else:
        return comparator


def uniform_validator(validator):
    """ unify validator

    Args:
        validator (dict): validator maybe in two formats:

            format1: this is kept for compatibility with the previous versions.
                {"check": "status_code", "assert": "eq", "expect": 201}
                {"check": "$resp_body_success", "assert": "eq", "expect": True}
            format2: recommended new version, {assert: [check_item, expected_value]}
                {'eq': ['status_code', 201]}
                {'eq': ['$resp_body_success', True]}

    Returns
        dict: validator info

            {
                "check": "status_code",
                "expect": 201,
                "assert": "equals"
            }

    """
    if not isinstance(validator, dict):
        raise ParamsError(f"invalid validator: {validator}")

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
            raise ParamsError(f"invalid validator: {validator}")

        check_item, expect_value = compare_values

    else:
        raise ParamsError(f"invalid validator: {validator}")

    # uniform comparator, e.g. lt => less_than, eq => equals
    assert_method = get_uniform_comparator(comparator)

    return {"check": check_item, "expect": expect_value, "assert": assert_method}


class ResponseObject(object):
    def __init__(self, resp_obj: requests.Response):
        """ initialize with a requests.Response object

        Args:
            resp_obj (instance): requests.Response instance

        """
        self.resp_obj = resp_obj

        try:
            body = resp_obj.json()
        except ValueError:
            body = resp_obj.content

        self.resp_obj_meta = {
            "status_code": resp_obj.status_code,
            "headers": resp_obj.headers,
            "cookies": dict(resp_obj.cookies),
            "body": body,
        }
        self.validation_results: Dict = {}

    def extract(self, extractors: Dict[Text, Text]) -> Dict[Text, Any]:
        if not extractors:
            return {}

        extract_mapping = {}
        for key, field in extractors.items():
            field_value = jmespath.search(field, self.resp_obj_meta)
            extract_mapping[key] = field_value

        logger.info(f"extract mapping: {extract_mapping}")
        return extract_mapping

    def validate(
        self,
        validators: Validators,
        variables_mapping: VariablesMapping = None,
        functions_mapping: FunctionsMapping = None,
    ) -> NoReturn:

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
            if "$" in check_item:
                # check_item is variable or function
                check_value = parse_data(
                    check_item, variables_mapping, functions_mapping
                )
            else:
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
                "expect_value": expect_value,
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
                validate_msg += (
                    f"\n"
                    f"check_item: {check_item}\n"
                    f"check_value: {check_value}({type(check_value).__name__})\n"
                    f"assert_method: {assert_method}\n"
                    f"expect_value: {expect_value}({type(expect_value).__name__})"
                )
                logger.error(validate_msg)
                failures.append(validate_msg)

            self.validation_results["validate_extractor"].append(validator_dict)

        if not validate_pass:
            failures_string = "\n".join([failure for failure in failures])
            raise ValidationFailure(failures_string)
