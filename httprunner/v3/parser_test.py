import unittest

from httprunner.v3.parser import parse_variables_mapping
from httprunner.v3.exceptions import VariableNotFound


class TestParserBasic(unittest.TestCase):

    def test_parse_variables_mapping(self):
        variables = {
            "varA": "$varB",
            "varB": "$varC",
            "varC": "123",
            "a": 1,
            "b": 2
        }
        parsed_variables = parse_variables_mapping(variables)
        print(parsed_variables)
        self.assertEqual(parsed_variables["varA"], "123")
        self.assertEqual(parsed_variables["varB"], "123")

    def test_parse_variables_mapping_exception(self):
        variables = {
            "varA": "$varB",
            "varB": "$varC",
            "a": 1,
            "b": 2
        }
        with self.assertRaises(VariableNotFound):
            parse_variables_mapping(variables)
