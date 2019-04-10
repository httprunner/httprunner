import unittest

from httprunner import validator


class TestValidator(unittest.TestCase):

    def test_is_testcases(self):
        data_structure = "path/to/file"
        self.assertFalse(validator.is_testcases(data_structure))
        data_structure = ["path/to/file1", "path/to/file2"]
        self.assertFalse(validator.is_testcases(data_structure))

        data_structure = {
            "project_mapping": {
                "PWD": "XXXXX",
                "functions": {},
                "env": {}
            },
            "testcases": [
                {   # testcase data structure
                    "config": {
                        "name": "desc1",
                        "path": "testcase1_path",
                        "variables": [],                    # optional
                    },
                    "teststeps": [
                        # test data structure
                        {
                            'name': 'test step desc1',
                            'variables': [],    # optional
                            'extract': [],      # optional
                            'validate': [],
                            'request': {}
                        },
                        # test_dict2   # another test dict
                    ]
                },
                # testcase_dict_2     # another testcase dict
            ]
        }
        self.assertTrue(validator.is_testcases(data_structure))
        data_structure = [
            {
                "name": "desc1",
                "config": {},
                "api": {},
                "testcases": ["testcase11", "testcase12"]
            },
            {
                "name": "desc2",
                "config": {},
                "api": {},
                "testcases": ["testcase21", "testcase22"]
            }
        ]
        self.assertTrue(data_structure)

    def test_is_variable(self):
        var1 = 123
        var2 = "abc"
        self.assertTrue(validator.is_variable(("var1", var1)))
        self.assertTrue(validator.is_variable(("var2", var2)))

        __var = 123
        self.assertFalse(validator.is_variable(("__var", __var)))

        func = lambda x: x + 1
        self.assertFalse(validator.is_variable(("func", func)))

        self.assertFalse(validator.is_variable(("unittest", unittest)))

    def test_is_function(self):
        func = lambda x: x + 1
        self.assertTrue(validator.is_function(func))
        self.assertTrue(validator.is_function(validator.is_testcase))

    def test_get_uniform_comparator(self):
        self.assertEqual(validator.get_uniform_comparator("eq"), "equals")
        self.assertEqual(validator.get_uniform_comparator("=="), "equals")
        self.assertEqual(validator.get_uniform_comparator("lt"), "less_than")
        self.assertEqual(validator.get_uniform_comparator("le"), "less_than_or_equals")
        self.assertEqual(validator.get_uniform_comparator("gt"), "greater_than")
        self.assertEqual(validator.get_uniform_comparator("ge"), "greater_than_or_equals")
        self.assertEqual(validator.get_uniform_comparator("ne"), "not_equals")

        self.assertEqual(validator.get_uniform_comparator("str_eq"), "string_equals")
        self.assertEqual(validator.get_uniform_comparator("len_eq"), "length_equals")
        self.assertEqual(validator.get_uniform_comparator("count_eq"), "length_equals")

        self.assertEqual(validator.get_uniform_comparator("len_gt"), "length_greater_than")
        self.assertEqual(validator.get_uniform_comparator("count_gt"), "length_greater_than")
        self.assertEqual(validator.get_uniform_comparator("count_greater_than"), "length_greater_than")

        self.assertEqual(validator.get_uniform_comparator("len_ge"), "length_greater_than_or_equals")
        self.assertEqual(validator.get_uniform_comparator("count_ge"), "length_greater_than_or_equals")
        self.assertEqual(validator.get_uniform_comparator("count_greater_than_or_equals"), "length_greater_than_or_equals")

        self.assertEqual(validator.get_uniform_comparator("len_lt"), "length_less_than")
        self.assertEqual(validator.get_uniform_comparator("count_lt"), "length_less_than")
        self.assertEqual(validator.get_uniform_comparator("count_less_than"), "length_less_than")

        self.assertEqual(validator.get_uniform_comparator("len_le"), "length_less_than_or_equals")
        self.assertEqual(validator.get_uniform_comparator("count_le"), "length_less_than_or_equals")
        self.assertEqual(validator.get_uniform_comparator("count_less_than_or_equals"), "length_less_than_or_equals")

    def test_parse_validator(self):
        _validator = {"check": "status_code", "comparator": "eq", "expect": 201}
        self.assertEqual(
            validator.uniform_validator(_validator),
            {"check": "status_code", "comparator": "equals", "expect": 201}
        )

        _validator = {'eq': ['status_code', 201]}
        self.assertEqual(
            validator.uniform_validator(_validator),
            {"check": "status_code", "comparator": "equals", "expect": 201}
        )


    def test_extend_validators(self):
        def_validators = [
            {'eq': ['v1', 200]},
            {"check": "s2", "expect": 16, "comparator": "len_eq"}
        ]
        current_validators = [
            {"check": "v1", "expect": 201},
            {'len_eq': ['s3', 12]}
        ]
        def_validators = [
            validator.uniform_validator(_validator)
            for _validator in def_validators
        ]
        ref_validators = [
            validator.uniform_validator(_validator)
            for _validator in current_validators
        ]

        extended_validators = validator.extend_validators(def_validators, ref_validators)
        self.assertIn(
            {"check": "v1", "expect": 201, "comparator": "equals"},
            extended_validators
        )
        self.assertIn(
            {"check": "s2", "expect": 16, "comparator": "length_equals"},
            extended_validators
        )
        self.assertIn(
            {"check": "s3", "expect": 12, "comparator": "length_equals"},
            extended_validators
        )

    def test_extend_validators_with_dict(self):
        def_validators = [
            {'eq': ["a", {"v": 1}]},
            {'eq': [{"b": 1}, 200]}
        ]
        current_validators = [
            {'len_eq': ['s3', 12]},
            {'eq': [{"b": 1}, 201]}
        ]
        def_validators = [
            validator.uniform_validator(_validator)
            for _validator in def_validators
        ]
        ref_validators = [
            validator.uniform_validator(_validator)
            for _validator in current_validators
        ]

        extended_validators = validator.extend_validators(def_validators, ref_validators)
        self.assertEqual(len(extended_validators), 3)
        self.assertIn({'check': {'b': 1}, 'expect': 201, 'comparator': 'equals'}, extended_validators)
        self.assertNotIn({'check': {'b': 1}, 'expect': 200, 'comparator': 'equals'}, extended_validators)
