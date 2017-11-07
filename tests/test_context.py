import os
import time
import unittest

from httprunner import runner, testcase, utils
from httprunner.context import Context
from httprunner.exception import ParamsError


class VariableBindsUnittest(unittest.TestCase):

    def setUp(self):
        self.context = Context()
        testcase_file_path = os.path.join(os.getcwd(), 'tests/data/demo_binds.yml')
        self.testcases = testcase._load_file(testcase_file_path)

    def test_context_init_functions(self):
        self.assertIn("get_timestamp", self.context.testset_functions_config)
        self.assertIn("gen_random_string", self.context.testset_functions_config)

        variables = [
            {"random": "${gen_random_string(5)}"},
            {"timestamp10": "${get_timestamp(10)}"}
        ]
        self.context.bind_variables(variables)
        context_variables = self.context.get_testcase_variables_mapping()

        self.assertEqual(len(context_variables["random"]), 5)
        self.assertEqual(len(context_variables["timestamp10"]), 10)

    def test_context_bind_testset_variables(self):
        # testcase in JSON format
        testcase1 = {
            "variables": [
                {"GLOBAL_TOKEN": "debugtalk"},
                {"token": "$GLOBAL_TOKEN"}
            ]
        }
        # testcase in YAML format
        testcase2 = self.testcases["bind_variables"]

        for testcase in [testcase1, testcase2]:
            variables = testcase['variables']
            self.context.bind_variables(variables, level="testset")

            testset_variables = self.context.testset_shared_variables_mapping
            testcase_variables = self.context.get_testcase_variables_mapping()
            self.assertIn("GLOBAL_TOKEN", testset_variables)
            self.assertIn("GLOBAL_TOKEN", testcase_variables)
            self.assertEqual(testset_variables["GLOBAL_TOKEN"], "debugtalk")
            self.assertIn("token", testset_variables)
            self.assertIn("token", testcase_variables)
            self.assertEqual(testset_variables["token"], "debugtalk")

    def test_context_bind_testcase_variables(self):
        testcase1 = {
            "variables": [
                {"GLOBAL_TOKEN": "debugtalk"},
                {"token": "$GLOBAL_TOKEN"}
            ]
        }
        testcase2 = self.testcases["bind_variables"]

        for testcase in [testcase1, testcase2]:
            variables = testcase['variables']
            self.context.bind_variables(variables)

            testset_variables = self.context.testset_shared_variables_mapping
            testcase_variables = self.context.get_testcase_variables_mapping()
            self.assertNotIn("GLOBAL_TOKEN", testset_variables)
            self.assertIn("GLOBAL_TOKEN", testcase_variables)
            self.assertEqual(testcase_variables["GLOBAL_TOKEN"], "debugtalk")
            self.assertNotIn("token", testset_variables)
            self.assertIn("token", testcase_variables)
            self.assertEqual(testcase_variables["token"], "debugtalk")

    def test_context_bind_lambda_functions(self):
        testcase1 = {
            "function_binds": {
                "add_one": lambda x: x + 1,
                "add_two_nums": lambda x, y: x + y
            },
            "variables": [
                {"add1": "${add_one(2)}"},
                {"sum2nums": "${add_two_nums(2,3)}"}
            ]
        }
        testcase2 = self.testcases["bind_lambda_functions"]

        for testcase in [testcase1, testcase2]:
            function_binds = testcase.get('function_binds', {})
            self.context.bind_functions(function_binds)

            variables = testcase['variables']
            self.context.bind_variables(variables)

            context_variables = self.context.get_testcase_variables_mapping()
            self.assertIn("add1", context_variables)
            self.assertEqual(context_variables["add1"], 3)
            self.assertIn("sum2nums", context_variables)
            self.assertEqual(context_variables["sum2nums"], 5)

    def test_context_bind_lambda_functions_with_import(self):
        testcase1 = {
            "requires": ["random", "string", "hashlib"],
            "function_binds": {
                "gen_random_string": "lambda str_len: ''.join(random.choice(string.ascii_letters + string.digits) for _ in range(str_len))",
                "gen_md5": "lambda *str_args: hashlib.md5(''.join(str_args).encode('utf-8')).hexdigest()"
            },
            "variables": [
                {"TOKEN": "debugtalk"},
                {"random": "${gen_random_string(5)}"},
                {"data": '{"name": "user", "password": "123456"}'},
                {"authorization": "${gen_md5($TOKEN, $data, $random)}"}
            ]
        }
        testcase2 = self.testcases["bind_lambda_functions_with_import"]

        for testcase in [testcase1, testcase2]:
            requires = testcase.get('requires', [])
            self.context.import_requires(requires)

            function_binds = testcase.get('function_binds', {})
            self.context.bind_functions(function_binds)

            variables = testcase['variables']
            self.context.bind_variables(variables)
            context_variables = self.context.get_testcase_variables_mapping()

            self.assertIn("TOKEN", context_variables)
            TOKEN = context_variables["TOKEN"]
            self.assertEqual(TOKEN, "debugtalk")
            self.assertIn("random", context_variables)
            self.assertIsInstance(context_variables["random"], str)
            self.assertEqual(len(context_variables["random"]), 5)
            random = context_variables["random"]
            self.assertIn("data", context_variables)
            data = context_variables["data"]
            self.assertIn("authorization", context_variables)
            self.assertEqual(len(context_variables["authorization"]), 32)
            authorization = context_variables["authorization"]
            self.assertEqual(utils.gen_md5(TOKEN, data, random), authorization)

    def test_import_module_items(self):
        testcase1 = {
            "import_module_items": ["tests.data.debugtalk"],
            "variables": [
                {"TOKEN": "debugtalk"},
                {"random": "${gen_random_string(5)}"},
                {"data": '{"name": "user", "password": "123456"}'},
                {"authorization": "${gen_md5($TOKEN, $data, $random)}"}
            ]
        }
        testcase2 = self.testcases["bind_module_functions"]

        for testcase in [testcase1, testcase2]:
            module_items = testcase.get('import_module_items', [])
            self.context.import_module_items(module_items)

            variables = testcase['variables']
            self.context.bind_variables(variables)
            context_variables = self.context.get_testcase_variables_mapping()

            self.assertIn("TOKEN", context_variables)
            TOKEN = context_variables["TOKEN"]
            self.assertEqual(TOKEN, "debugtalk")
            self.assertIn("random", context_variables)
            self.assertIsInstance(context_variables["random"], str)
            self.assertEqual(len(context_variables["random"]), 5)
            random = context_variables["random"]
            self.assertIn("data", context_variables)
            data = context_variables["data"]
            self.assertIn("authorization", context_variables)
            self.assertEqual(len(context_variables["authorization"]), 32)
            authorization = context_variables["authorization"]
            self.assertEqual(utils.gen_md5(TOKEN, data, random), authorization)
            self.assertIn("SECRET_KEY", context_variables)
            SECRET_KEY = context_variables["SECRET_KEY"]
            self.assertEqual(SECRET_KEY, "DebugTalk")

    def test_get_parsed_request(self):
        test_runner = runner.Runner()
        testcase = {
            "import_module_items": ["tests.data.debugtalk"],
            "variables": [
                {"TOKEN": "debugtalk"},
                {"random": "${gen_random_string(5)}"},
                {"data": '{"name": "user", "password": "123456"}'},
                {"authorization": "${gen_md5($TOKEN, $data, $random)}"}
            ],
            "request": {
                "url": "http://127.0.0.1:5000/api/users/1000",
                "METHOD": "POST",
                "Headers": {
                    "Content-Type": "application/json",
                    "authorization": "$authorization",
                    "random": "$random",
                    "SECRET_KEY": "$SECRET_KEY"
                },
                "Data": "$data"
            }
        }
        parsed_request = test_runner.init_config(testcase, level="testcase")
        self.assertIn("authorization", parsed_request["headers"])
        self.assertEqual(len(parsed_request["headers"]["authorization"]), 32)
        self.assertIn("random", parsed_request["headers"])
        self.assertEqual(len(parsed_request["headers"]["random"]), 5)
        self.assertIn("data", parsed_request)
        self.assertEqual(parsed_request["data"], testcase["variables"][2]["data"])
        self.assertEqual(parsed_request["headers"]["SECRET_KEY"], "DebugTalk")

    def test_exec_content_functions(self):
        test_runner = runner.Runner()
        content = "${sleep(1)}"
        start_time = time.time()
        test_runner.context.exec_content_functions(content)
        end_time = time.time()
        elapsed_time = end_time - start_time
        self.assertGreater(elapsed_time, 1)
