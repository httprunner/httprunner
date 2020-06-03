"""
Built-in validate comparators.
"""

import re


def equal(check_value, expect_value):
    assert check_value == expect_value


def greater_than(check_value, expect_value):
    assert check_value > expect_value


def less_than(check_value, expect_value):
    assert check_value < expect_value


def greater_or_equals(check_value, expect_value):
    assert check_value >= expect_value


def less_or_equals(check_value, expect_value):
    assert check_value <= expect_value


def not_equal(check_value, expect_value):
    assert check_value != expect_value


def string_equals(check_value, expect_value):
    assert str(check_value) == str(expect_value)


def length_equal(check_value, expect_value):
    assert isinstance(expect_value, int)
    assert len(check_value) == expect_value


def length_greater_than(check_value, expect_value):
    assert isinstance(expect_value, (int, float))
    assert len(check_value) > expect_value


def length_greater_or_equals(check_value, expect_value):
    assert isinstance(expect_value, (int, float))
    assert len(check_value) >= expect_value


def length_less_than(check_value, expect_value):
    assert isinstance(expect_value, (int, float))
    assert len(check_value) < expect_value


def length_less_or_equals(check_value, expect_value):
    assert isinstance(expect_value, (int, float))
    assert len(check_value) <= expect_value


def contains(check_value, expect_value):
    assert isinstance(check_value, (list, tuple, dict, str, bytes))
    assert expect_value in check_value


def contained_by(check_value, expect_value):
    assert isinstance(expect_value, (list, tuple, dict, str, bytes))
    assert check_value in expect_value


def type_match(check_value, expect_value):
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

    assert isinstance(check_value, get_type(expect_value))


def regex_match(check_value, expect_value):
    assert isinstance(expect_value, str)
    assert isinstance(check_value, str)
    assert re.match(expect_value, check_value)


def startswith(check_value, expect_value):
    assert str(check_value).startswith(str(expect_value))


def endswith(check_value, expect_value):
    assert str(check_value).endswith(str(expect_value))
