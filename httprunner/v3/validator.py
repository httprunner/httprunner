from typing import Text

import jmespath
from loguru import logger

from httprunner.v3.exceptions import ParamsError, ValidationFailure
from httprunner.v3.response import ResponseObject


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
    elif comparator in ["len_gt", "count_gt", "length_greater_than", "count_greater_than"]:
        return "length_greater_than"
    elif comparator in ["len_ge", "count_ge", "length_greater_than_or_equals",
                        "count_greater_than_or_equals"]:
        return "length_greater_than_or_equals"
    elif comparator in ["len_lt", "count_lt", "length_less_than", "count_less_than"]:
        return "length_less_than"
    elif comparator in ["len_le", "count_le", "length_less_than_or_equals",
                        "count_less_than_or_equals"]:
        return "length_less_than_or_equals"
    else:
        return comparator


def uniform_validator(validator):
    """ unify validator

    Args:
        validator (dict): validator maybe in two formats:

            format1: this is kept for compatiblity with the previous versions.
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

    return {
        "check": check_item,
        "expect": expect_value,
        "assert": assert_method
    }


class AssertMethods(object):

    @staticmethod
    def equals(actual_value, expect_value):
        assert actual_value == expect_value

    @staticmethod
    def less_than(actual_value, expect_value):
        assert actual_value < expect_value

    @staticmethod
    def greater_than(actual_value, expect_value):
        assert actual_value > expect_value


class Validator(object):

    def __init__(self, resp_obj: ResponseObject):
        self.resp_meta = {
            "status_code": resp_obj.obj.status_code,
            "headers": resp_obj.obj.headers,
            "body": resp_obj.obj.json()
        }

    def validate(self, validators):

        for v in validators:
            u_validator = uniform_validator(v)
            field = u_validator["check"]
            assert_method = u_validator["assert"]
            expect_value = u_validator["expect"]
            actual_value = jmespath.search(field, self.resp_meta)

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
