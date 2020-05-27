import os
import unittest

from httprunner import compat


class TestCompat(unittest.TestCase):
    def test_convert_jmespath(self):

        self.assertEqual(compat.convert_jmespath("content.abc"), "body.abc")
        self.assertEqual(compat.convert_jmespath("json.abc"), "body.abc")
        self.assertEqual(
            compat.convert_jmespath("headers.Content-Type"), 'headers."Content-Type"'
        )
        self.assertEqual(
            compat.convert_jmespath("body.data.buildings.0.building_id"),
            "body.data.buildings[0].building_id",
        )

    def test_convert_extractors(self):
        self.assertEqual(
            compat.convert_extractors(
                [{"varA": "content.varA"}, {"varB": "json.varB"}]
            ),
            {"varA": "body.varA", "varB": "body.varB"},
        )
        self.assertEqual(
            compat.convert_extractors([{"varA": "content.0.varA"}]),
            {"varA": "body[0].varA"},
        )
        self.assertEqual(
            compat.convert_extractors({"varA": "content.0.varA"}),
            {"varA": "body[0].varA"},
        )

    def test_convert_validators(self):
        self.assertEqual(
            compat.convert_validators(
                [{"check": "content.abc", "assert": "eq", "expect": 201}]
            ),
            [{"check": "body.abc", "assert": "eq", "expect": 201}],
        )
        self.assertEqual(
            compat.convert_validators([{"eq": ["content.abc", 201]}]),
            [{"eq": ["body.abc", 201]}],
        )
        self.assertEqual(
            compat.convert_validators([{"eq": ["content.0.name", 201]}]),
            [{"eq": ["body[0].name", 201]}],
        )
