import os
import unittest
from httprunner import parser, exceptions


class TestParser(unittest.TestCase):

    def test_parse_string_value(self):
        self.assertEqual(parser.parse_string_value("123"), 123)
        self.assertEqual(parser.parse_string_value("12.3"), 12.3)
        self.assertEqual(parser.parse_string_value("a123"), "a123")
        self.assertEqual(parser.parse_string_value("$var"), "$var")
        self.assertEqual(parser.parse_string_value("${func}"), "${func}")

    def test_extract_variables(self):
        self.assertEqual(
            parser.extract_variables("$var"),
            ["var"]
        )
        self.assertEqual(
            parser.extract_variables("$var123"),
            ["var123"]
        )
        self.assertEqual(
            parser.extract_variables("$var_name"),
            ["var_name"]
        )
        self.assertEqual(
            parser.extract_variables("var"),
            []
        )
        self.assertEqual(
            parser.extract_variables("a$var"),
            ["var"]
        )
        self.assertEqual(
            parser.extract_variables("$v ar"),
            ["v"]
        )
        self.assertEqual(
            parser.extract_variables(" "),
            []
        )
        self.assertEqual(
            parser.extract_variables("$abc*"),
            ["abc"]
        )
        self.assertEqual(
            parser.extract_variables("${func()}"),
            []
        )
        self.assertEqual(
            parser.extract_variables("${func(1,2)}"),
            []
        )
        self.assertEqual(
            parser.extract_variables("${gen_md5($TOKEN, $data, $random)}"),
            ["TOKEN", "data", "random"]
        )

    def test_parse_function(self):
        self.assertEqual(
            parser.parse_function("func()"),
            {'func_name': 'func', 'args': [], 'kwargs': {}}
        )
        self.assertEqual(
            parser.parse_function("func(5)"),
            {'func_name': 'func', 'args': [5], 'kwargs': {}}
        )
        self.assertEqual(
            parser.parse_function("func(1, 2)"),
            {'func_name': 'func', 'args': [1, 2], 'kwargs': {}}
        )
        self.assertEqual(
            parser.parse_function("func(a=1, b=2)"),
            {'func_name': 'func', 'args': [], 'kwargs': {'a': 1, 'b': 2}}
        )
        self.assertEqual(
            parser.parse_function("func(a= 1, b =2)"),
            {'func_name': 'func', 'args': [], 'kwargs': {'a': 1, 'b': 2}}
        )
        self.assertEqual(
            parser.parse_function("func(1, 2, a=3, b=4)"),
            {'func_name': 'func', 'args': [1, 2], 'kwargs': {'a': 3, 'b': 4}}
        )
        self.assertEqual(
            parser.parse_function("func($request, 123)"),
            {'func_name': 'func', 'args': ["$request", 123], 'kwargs': {}}
        )
        self.assertEqual(
            parser.parse_function("func( )"),
            {'func_name': 'func', 'args': [], 'kwargs': {}}
        )
        self.assertEqual(
            parser.parse_function("func(hello world, a=3, b=4)"),
            {'func_name': 'func', 'args': ["hello world"], 'kwargs': {'a': 3, 'b': 4}}
        )
        self.assertEqual(
            parser.parse_function("func($request, 12 3)"),
            {'func_name': 'func', 'args': ["$request", '12 3'], 'kwargs': {}}
        )

    def test_parse_validator(self):
        validator = {"check": "status_code", "comparator": "eq", "expect": 201}
        self.assertEqual(
            parser.parse_validator(validator),
            {"check": "status_code", "comparator": "eq", "expect": 201}
        )

        validator = {'eq': ['status_code', 201]}
        self.assertEqual(
            parser.parse_validator(validator),
            {"check": "status_code", "comparator": "eq", "expect": 201}
        )
