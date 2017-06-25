import os
import unittest

from ate import utils
from ate.context import Context


class VariableBindsUnittest(unittest.TestCase):

    def setUp(self):
        self.context = Context()
        testcase_file_path = os.path.join(os.getcwd(), 'test/data/demo_binds.yml')
        self.testcases = utils.load_testcases(testcase_file_path)

    def test_context_variable_string(self):
        # testcase in JSON format
        testcase1 = {
            "variable_binds": [
                {"TOKEN": "debugtalk"}
            ]
        }
        # testcase in YAML format
        testcase2 = self.testcases[0]

        for testcase in [testcase1, testcase2]:
            variable_binds = testcase['variable_binds']
            self.context.bind_variables(variable_binds)

            context_variables = self.context.variables
            self.assertIn("TOKEN", context_variables)
            self.assertEqual(context_variables["TOKEN"], "debugtalk")

    def test_context_variable_list(self):
        testcase1 = {
            "variable_binds": [
                {"var": [1, 2, 3]}
            ]
        }
        testcase2 = self.testcases[1]

        for testcase in [testcase1, testcase2]:
            variable_binds = testcase['variable_binds']
            self.context.bind_variables(variable_binds)

            context_variables = self.context.variables
            self.assertIn("var", context_variables)
            self.assertEqual(context_variables["var"], [1, 2, 3])

    def test_context_variable_json(self):
        testcase1 = {
            "variable_binds": [
                {"data": {'name': 'user', 'password': '123456'}}
            ]
        }
        testcase2 = self.testcases[2]

        for testcase in [testcase1, testcase2]:
            variable_binds = testcase['variable_binds']
            self.context.bind_variables(variable_binds)

            context_variables = self.context.variables
            self.assertIn("data", context_variables)
            self.assertEqual(
                context_variables["data"],
                {'name': 'user', 'password': '123456'}
            )

    def test_context_variable_variable(self):
        testcase1 = {
            "variable_binds": [
                {"GLOBAL_TOKEN": "debugtalk"},
                {"token": "$GLOBAL_TOKEN"}
            ]
        }
        testcase2 = self.testcases[3]

        for testcase in [testcase1, testcase2]:
            variable_binds = testcase['variable_binds']
            self.context.bind_variables(variable_binds)

            context_variables = self.context.variables
            self.assertIn("GLOBAL_TOKEN", context_variables)
            self.assertEqual(context_variables["GLOBAL_TOKEN"], "debugtalk")
            self.assertIn("token", context_variables)
            self.assertEqual(context_variables["token"], "debugtalk")

    def test_context_variable_function_lambda(self):
        testcase1 = {
            "function_binds": {
                "add_one": lambda x: x + 1,
                "add_two_nums": lambda x, y: x + y
            },
            "variable_binds": [
                {"add1": {"func": "add_one", "args": [2]}},
                {"sum2nums": {"func": "add_two_nums", "args": [2, 3]}}
            ]
        }
        testcase2 = self.testcases[4]

        for testcase in [testcase1, testcase2]:
            function_binds = testcase.get('function_binds', {})
            self.context.bind_functions(function_binds)

            variable_binds = testcase['variable_binds']
            self.context.bind_variables(variable_binds)

            context_variables = self.context.variables
            self.assertIn("add1", context_variables)
            self.assertEqual(context_variables["add1"], 3)
            self.assertIn("sum2nums", context_variables)
            self.assertEqual(context_variables["sum2nums"], 5)

    def test_context_variable_function_lambda_with_import(self):
        testcase1 = {
            "requires": ["random", "string", "hashlib"],
            "function_binds": {
                "gen_random_string": "lambda str_len: ''.join(random.choice(string.ascii_letters + string.digits) for _ in range(str_len))",
                "gen_md5": "lambda *str_args: hashlib.md5(''.join(str_args).encode('utf-8')).hexdigest()"
            },
            "variable_binds": [
                {"TOKEN": "debugtalk"},
                {"random": {"func": "gen_random_string", "args": [5]}},
                {"data": "{'name': 'user', 'password': '123456'}"},
                {"md5": {"func": "gen_md5", "args": ["$TOKEN", "$data", "$random"]}}
            ]
        }
        testcase2 = self.testcases[5]

        for testcase in [testcase1, testcase2]:
            requires = testcase.get('requires', [])
            self.context.import_requires(requires)

            function_binds = testcase.get('function_binds', {})
            self.context.bind_functions(function_binds)

            variable_binds = testcase['variable_binds']
            self.context.bind_variables(variable_binds)

            context_variables = self.context.variables
            self.assertIn("random", context_variables)
            self.assertIsInstance(context_variables["random"], str)
            self.assertEqual(len(context_variables["random"]), 5)
            self.assertIn("md5", context_variables)
            self.assertEqual(len(context_variables["md5"]), 32)
