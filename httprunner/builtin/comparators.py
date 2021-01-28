"""
Built-in validate comparators.
"""

import re
from typing import Text, Any, Union


def equal(check_value: Any, expect_value: Any, message: Text = ""):
    assert check_value == expect_value, message


def greater_than(
    check_value: Union[int, float], expect_value: Union[int, float], message: Text = ""
):
    assert check_value > expect_value, message


def less_than(
    check_value: Union[int, float], expect_value: Union[int, float], message: Text = ""
):
    assert check_value < expect_value, message


def greater_or_equals(
    check_value: Union[int, float], expect_value: Union[int, float], message: Text = ""
):
    assert check_value >= expect_value, message


def less_or_equals(
    check_value: Union[int, float], expect_value: Union[int, float], message: Text = ""
):
    assert check_value <= expect_value, message


def not_equal(check_value: Any, expect_value: Any, message: Text = ""):
    assert check_value != expect_value, message


def string_equals(check_value: Text, expect_value: Any, message: Text = ""):
    assert str(check_value) == str(expect_value), message


def length_equal(check_value: Text, expect_value: int, message: Text = ""):
    assert isinstance(expect_value, int), "expect_value should be int type"
    assert len(check_value) == expect_value, message


def length_greater_than(
    check_value: Text, expect_value: Union[int, float], message: Text = ""
):
    assert isinstance(
        expect_value, (int, float)
    ), "expect_value should be int/float type"
    assert len(check_value) > expect_value, message


def length_greater_or_equals(
    check_value: Text, expect_value: Union[int, float], message: Text = ""
):
    assert isinstance(
        expect_value, (int, float)
    ), "expect_value should be int/float type"
    assert len(check_value) >= expect_value, message


def length_less_than(
    check_value: Text, expect_value: Union[int, float], message: Text = ""
):
    assert isinstance(
        expect_value, (int, float)
    ), "expect_value should be int/float type"
    assert len(check_value) < expect_value, message


def length_less_or_equals(
    check_value: Text, expect_value: Union[int, float], message: Text = ""
):
    assert isinstance(
        expect_value, (int, float)
    ), "expect_value should be int/float type"
    assert len(check_value) <= expect_value, message


def contains(check_value: Any, expect_value: Any, message: Text = ""):
    assert isinstance(
        check_value, (list, tuple, dict, str, bytes)
    ), "expect_value should be list/tuple/dict/str/bytes type"
    assert expect_value in check_value, message


def contained_by(check_value: Any, expect_value: Any, message: Text = ""):
    assert isinstance(
        expect_value, (list, tuple, dict, str, bytes)
    ), "expect_value should be list/tuple/dict/str/bytes type"
    assert check_value in expect_value, message


def type_match(check_value: Any, expect_value: Any, message: Text = ""):
    def get_type(name):
        if isinstance(name, type):
            return name
        elif isinstance(name, str):
            try:
                return __builtins__[name]
            except KeyError:
                raise ValueError(name)
        else:
            raise ValueError(name)

    if expect_value in ["None", "NoneType", None]:
        assert check_value is None, message
    else:
        assert type(check_value) == get_type(expect_value), message


def regex_match(check_value: Text, expect_value: Any, message: Text = ""):
    assert isinstance(expect_value, str), "expect_value should be Text type"
    assert isinstance(check_value, str), "check_value should be Text type"
    assert re.match(expect_value, check_value), message


def startswith(check_value: Any, expect_value: Any, message: Text = ""):
    assert str(check_value).startswith(str(expect_value)), message


def endswith(check_value: Text, expect_value: Any, message: Text = ""):
    assert str(check_value).endswith(str(expect_value)), message
