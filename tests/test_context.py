import os
import time
import unittest

import requests
from httprunner import context, exceptions, loader, parser, response, runner
from tests.base import ApiServerUnittest


class TestContext(ApiServerUnittest):

    def setUp(self):
        self.context = context.Context()
        testcase_file_path = os.path.join(os.getcwd(), 'tests/data/demo_binds.yml')
        self.testcases = loader.load_file(testcase_file_path)

    def test_context_init_functions(self):
        self.assertIn("get_timestamp", self.context.testset_functions_config)
        self.assertIn("gen_random_string", self.context.testset_functions_config)

        variables = [
            {"random": "${gen_random_string(5)}"},
            {"timestamp10": "${get_timestamp(10)}"}
        ]
        self.context.bind_variables(variables)
        context_variables = self.context.testcase_variables_mapping

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
            testcase_variables = self.context.testcase_variables_mapping
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
            testcase_variables = self.context.testcase_variables_mapping
            self.assertNotIn("GLOBAL_TOKEN", testset_variables)
            self.assertIn("GLOBAL_TOKEN", testcase_variables)
            self.assertEqual(testcase_variables["GLOBAL_TOKEN"], "debugtalk")
            self.assertNotIn("token", testset_variables)
            self.assertIn("token", testcase_variables)
            self.assertEqual(testcase_variables["token"], "debugtalk")

    def test_context_bind_lambda_functions(self):
        function_binds = {
            "add_one": lambda x: x + 1,
            "add_two_nums": lambda x, y: x + y
        }
        variables = [
            {"add1": "${add_one(2)}"},
            {"sum2nums": "${add_two_nums(2,3)}"}
        ]
        self.context.bind_functions(function_binds)
        self.context.bind_variables(variables)

        context_variables = self.context.testcase_variables_mapping
        self.assertIn("add1", context_variables)
        self.assertEqual(context_variables["add1"], 3)
        self.assertIn("sum2nums", context_variables)
        self.assertEqual(context_variables["sum2nums"], 5)

    def test_call_builtin_functions(self):
        testcase1 = {
            "variables": [
                {"length": "${len(debugtalk)}"},
                {"smallest": "${min(2, 3, 8)}"},
                {"largest": "${max(2, 3, 8)}"}
            ]
        }
        testcase2 = self.testcases["builtin_functions"]

        for testcase in [testcase1, testcase2]:
            variables = testcase['variables']
            self.context.bind_variables(variables)

            context_variables = self.context.testcase_variables_mapping
            self.assertEqual(context_variables["length"], 9)
            self.assertEqual(context_variables["smallest"], 2)
            self.assertEqual(context_variables["largest"], 8)

    def test_import_module_items(self):
        variables = [
            {"TOKEN": "debugtalk"},
            {"random": "${gen_random_string(5)}"},
            {"data": '{"name": "user", "password": "123456"}'},
            {"authorization": "${gen_md5($TOKEN, $data, $random)}"}
        ]
        from tests import debugtalk
        from tests.debugtalk import gen_md5
        self.context.import_module_items(debugtalk)
        self.context.bind_variables(variables)
        context_variables = self.context.testcase_variables_mapping

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
        self.assertEqual(gen_md5(TOKEN, data, random), authorization)
        self.assertIn("SECRET_KEY", context_variables)
        SECRET_KEY = context_variables["SECRET_KEY"]
        self.assertEqual(SECRET_KEY, "DebugTalk")

    def test_get_parsed_request(self):
        test_runner = runner.Runner()
        testcase = {
            "variables": [
                {"TOKEN": "debugtalk"},
                {"random": "${gen_random_string(5)}"},
                {"data": '{"name": "user", "password": "123456"}'},
                {"authorization": "${gen_md5($TOKEN, $data, $random)}"}
            ],
            "request": {
                "url": "http://127.0.0.1:5000/api/users/1000",
                "method": "POST",
                "headers": {
                    "Content-Type": "application/json",
                    "authorization": "$authorization",
                    "random": "$random",
                    "secret_key": "$SECRET_KEY"
                },
                "data": "$data"
            }
        }
        from tests import debugtalk
        self.context.import_module_items(debugtalk)
        self.context.bind_variables(testcase["variables"])
        parsed_request = self.context.get_parsed_request(testcase["request"])
        self.assertIn("authorization", parsed_request["headers"])
        self.assertEqual(len(parsed_request["headers"]["authorization"]), 32)
        self.assertIn("random", parsed_request["headers"])
        self.assertEqual(len(parsed_request["headers"]["random"]), 5)
        self.assertIn("data", parsed_request)
        self.assertEqual(parsed_request["data"], testcase["variables"][2]["data"])
        self.assertEqual(parsed_request["headers"]["secret_key"], "DebugTalk")

    def test_exec_content_functions(self):
        test_runner = runner.Runner()
        content = "${sleep_N_secs(1)}"
        start_time = time.time()
        test_runner.context.eval_content(content)
        end_time = time.time()
        elapsed_time = end_time - start_time
        self.assertGreater(elapsed_time, 1)

    def test_do_validation(self):
        self.context.do_validation(
            {"check": "check", "check_value": 1, "expect": 1, "comparator": "eq"}
        )
        self.context.do_validation(
            {"check": "check", "check_value": "abc", "expect": "abc", "comparator": "=="}
        )

        config_dict = {
            "path": 'tests/data/demo_testset_hardcode.yml'
        }
        self.context.config_context(config_dict, "testset")
        self.context.do_validation(
            {"check": "status_code", "check_value": "201", "expect": 3, "comparator": "sum_status_code"}
        )

    def test_validate(self):
        url = "http://127.0.0.1:5000/"
        resp = requests.get(url)
        resp_obj = response.ResponseObject(resp)

        validators = [
            {"eq": ["$resp_status_code", 201]},
            {"check": "$resp_status_code", "comparator": "eq", "expect": 201},
            {"check": "$resp_body_success", "comparator": "eq", "expect": True}
        ]
        variables = [
            {"resp_status_code": 200},
            {"resp_body_success": True}
        ]
        self.context.bind_variables(variables)

        with self.assertRaises(exceptions.ValidationFailure):
            self.context.validate(validators, resp_obj)

        validators = [
            {"eq": ["$resp_status_code", 201]},
            {"check": "$resp_status_code", "comparator": "eq", "expect": 201},
            {"check": "$resp_body_success", "comparator": "eq", "expect": True},
            {"check": "${is_status_code_200($resp_status_code)}", "comparator": "eq", "expect": False}
        ]
        variables = [
            {"resp_status_code": 201},
            {"resp_body_success": True}
        ]
        self.context.bind_variables(variables)
        from tests.debugtalk import is_status_code_200
        functions = {
            "is_status_code_200": is_status_code_200
        }
        self.context.bind_functions(functions)

        self.context.validate(validators, resp_obj)

    def test_validate_exception(self):
        url = "http://127.0.0.1:5000/"
        resp = requests.get(url)
        resp_obj = response.ResponseObject(resp)

        # expected value missed in validators
        validators = [
            {"eq": ["$resp_status_code", 201]},
            {"check": "$resp_status_code", "comparator": "eq", "expect": 201}
        ]
        variables = []
        self.context.bind_variables(variables)

        with self.assertRaises(exceptions.VariableNotFound):
            self.context.validate(validators, resp_obj)

        # expected value missed in variables mapping
        variables = [
            {"resp_status_code": 200}
        ]
        self.context.bind_variables(variables)

        with self.assertRaises(exceptions.ValidationFailure):
            self.context.validate(validators, resp_obj)


class TestTestcaseParser(unittest.TestCase):

    def test_eval_content_variables(self):
        variables = {
            "var_1": "abc",
            "var_2": "def",
            "var_3": 123,
            "var_4": {"a": 1},
            "var_5": True,
            "var_6": None
        }
        testcase_parser = context.TestcaseParser(variables=variables)
        self.assertEqual(
            testcase_parser._eval_content_variables("$var_1"),
            "abc"
        )
        self.assertEqual(
            testcase_parser._eval_content_variables("var_1"),
            "var_1"
        )
        self.assertEqual(
            testcase_parser._eval_content_variables("$var_1#XYZ"),
            "abc#XYZ"
        )
        self.assertEqual(
            testcase_parser._eval_content_variables("/$var_1/$var_2/var3"),
            "/abc/def/var3"
        )
        self.assertEqual(
            testcase_parser._eval_content_variables("/$var_1/$var_2/$var_1"),
            "/abc/def/abc"
        )
        self.assertEqual(
            testcase_parser._eval_content_variables("${func($var_1, $var_2, xyz)}"),
            "${func(abc, def, xyz)}"
        )
        self.assertEqual(
            testcase_parser._eval_content_variables("$var_3"),
            123
        )
        self.assertEqual(
            testcase_parser._eval_content_variables("$var_4"),
            {"a": 1}
        )
        self.assertEqual(
            testcase_parser._eval_content_variables("$var_5"),
            True
        )
        self.assertEqual(
            testcase_parser._eval_content_variables("abc$var_5"),
            "abcTrue"
        )
        self.assertEqual(
            testcase_parser._eval_content_variables("abc$var_4"),
            "abc{'a': 1}"
        )
        self.assertEqual(
            testcase_parser._eval_content_variables("$var_6"),
            None
        )

    def test_eval_content_variables_search_upward(self):
        testcase_parser = context.TestcaseParser()

        with self.assertRaises(exceptions.VariableNotFound):
            testcase_parser._eval_content_variables("/api/$SECRET_KEY")

        testcase_parser.file_path = "tests/data/demo_testset_hardcode.yml"
        content = testcase_parser._eval_content_variables("/api/$SECRET_KEY")
        self.assertEqual(content, "/api/DebugTalk")


    def test_parse_content_with_bindings_variables(self):
        variables = {
            "str_1": "str_value1",
            "str_2": "str_value2"
        }
        testcase_parser = context.TestcaseParser(variables=variables)
        self.assertEqual(
            testcase_parser.eval_content_with_bindings("$str_1"),
            "str_value1"
        )
        self.assertEqual(
            testcase_parser.eval_content_with_bindings("123$str_1/456"),
            "123str_value1/456"
        )

        with self.assertRaises(exceptions.VariableNotFound):
            testcase_parser.eval_content_with_bindings("$str_3")

        self.assertEqual(
            testcase_parser.eval_content_with_bindings(["$str_1", "str3"]),
            ["str_value1", "str3"]
        )
        self.assertEqual(
            testcase_parser.eval_content_with_bindings({"key": "$str_1"}),
            {"key": "str_value1"}
        )

    def test_parse_content_with_bindings_multiple_identical_variables(self):
        variables = {
            "userid": 100,
            "data": 1498
        }
        testcase_parser = context.TestcaseParser(variables=variables)
        content = "/users/$userid/training/$data?userId=$userid&data=$data"
        self.assertEqual(
            testcase_parser.eval_content_with_bindings(content),
            "/users/100/training/1498?userId=100&data=1498"
        )

    def test_parse_variables_multiple_identical_variables(self):
        variables = {
            "user": 100,
            "userid": 1000,
            "data": 1498
        }
        testcase_parser = context.TestcaseParser(variables=variables)
        content = "/users/$user/$userid/$data?userId=$userid&data=$data"
        self.assertEqual(
            testcase_parser.eval_content_with_bindings(content),
            "/users/100/1000/1498?userId=1000&data=1498"
        )

    def test_parse_content_with_bindings_functions(self):
        import random, string
        functions = {
            "gen_random_string": lambda str_len: ''.join(random.choice(string.ascii_letters + string.digits) \
                for _ in range(str_len))
        }
        testcase_parser = context.TestcaseParser(functions=functions)

        result = testcase_parser.eval_content_with_bindings("${gen_random_string(5)}")
        self.assertEqual(len(result), 5)

        add_two_nums = lambda a, b=1: a + b
        functions["add_two_nums"] = add_two_nums
        self.assertEqual(
            testcase_parser.eval_content_with_bindings("${add_two_nums(1)}"),
            2
        )
        self.assertEqual(
            testcase_parser.eval_content_with_bindings("${add_two_nums(1, 2)}"),
            3
        )

    def test_extract_functions(self):
        self.assertEqual(
            parser.extract_functions("${func()}"),
            ["func()"]
        )
        self.assertEqual(
            parser.extract_functions("${func(5)}"),
            ["func(5)"]
        )
        self.assertEqual(
            parser.extract_functions("${func(a=1, b=2)}"),
            ["func(a=1, b=2)"]
        )
        self.assertEqual(
            parser.extract_functions("${func(1, $b, c=$x, d=4)}"),
            ["func(1, $b, c=$x, d=4)"]
        )
        self.assertEqual(
            parser.extract_functions("/api/1000?_t=${get_timestamp()}"),
            ["get_timestamp()"]
        )
        self.assertEqual(
            parser.extract_functions("/api/${add(1, 2)}"),
            ["add(1, 2)"]
        )
        self.assertEqual(
            parser.extract_functions("/api/${add(1, 2)}?_t=${get_timestamp()}"),
            ["add(1, 2)", "get_timestamp()"]
        )
        self.assertEqual(
            parser.extract_functions("abc${func(1, 2, a=3, b=4)}def"),
            ["func(1, 2, a=3, b=4)"]
        )

    def test_eval_content_functions(self):
        functions = {
            "add_two_nums": lambda a, b=1: a + b
        }
        testcase_parser = context.TestcaseParser(functions=functions)
        self.assertEqual(
            testcase_parser._eval_content_functions("${add_two_nums(1, 2)}"),
            3
        )
        self.assertEqual(
            testcase_parser._eval_content_functions("/api/${add_two_nums(1, 2)}"),
            "/api/3"
        )

    def test_eval_content_functions_search_upward(self):
        testcase_parser = context.TestcaseParser()

        with self.assertRaises(exceptions.FunctionNotFound):
            testcase_parser._eval_content_functions("/api/${gen_md5(abc)}")

        testcase_parser.file_path = "tests/data/demo_testset_hardcode.yml"
        content = testcase_parser._eval_content_functions("/api/${gen_md5(abc)}")
        self.assertEqual(content, "/api/900150983cd24fb0d6963f7d28e17f72")

    def test_parse_content_with_bindings_testcase(self):
        variables = {
            "uid": "1000",
            "random": "A2dEx",
            "authorization": "a83de0ff8d2e896dbd8efb81ba14e17d",
            "data": {"name": "user", "password": "123456"}
        }
        functions = {
            "add_two_nums": lambda a, b=1: a + b,
            "get_timestamp": lambda: int(time.time() * 1000)
        }
        testcase_template = {
            "url": "http://127.0.0.1:5000/api/users/$uid/${add_two_nums(1,2)}",
            "method": "POST",
            "headers": {
                "Content-Type": "application/json",
                "authorization": "$authorization",
                "random": "$random",
                "sum": "${add_two_nums(1, 2)}"
            },
            "body": "$data"
        }
        parsed_testcase = context.TestcaseParser(variables, functions)\
            .eval_content_with_bindings(testcase_template)

        self.assertEqual(
            parsed_testcase["url"],
            "http://127.0.0.1:5000/api/users/1000/3"
        )
        self.assertEqual(
            parsed_testcase["headers"]["authorization"],
            variables["authorization"]
        )
        self.assertEqual(
            parsed_testcase["headers"]["random"],
            variables["random"]
        )
        self.assertEqual(
            parsed_testcase["body"],
            variables["data"]
        )
        self.assertEqual(
            parsed_testcase["headers"]["sum"],
            3
        )

    def test_parse_parameters_raw_list(self):
        parameters = [
            {"user_agent": ["iOS/10.1", "iOS/10.2", "iOS/10.3"]},
            {"username-password": [("user1", "111111"), ["test2", "222222"]]}
        ]
        cartesian_product_parameters = context.parse_parameters(parameters)
        self.assertEqual(
            len(cartesian_product_parameters),
            3 * 2
        )
        self.assertEqual(
            cartesian_product_parameters[0],
            {'user_agent': 'iOS/10.1', 'username': 'user1', 'password': '111111'}
        )

    def test_parse_parameters_parameterize(self):
        parameters = [
            {"app_version": "${parameterize(app_version.csv)}"},
            {"username-password": "${parameterize(account.csv)}"}
        ]
        testset_path = os.path.join(
            os.getcwd(),
            "tests/data/demo_parameters.yml"
        )
        cartesian_product_parameters = context.parse_parameters(
            parameters,
            testset_path
        )
        self.assertEqual(
            len(cartesian_product_parameters),
            2 * 3
        )

    def test_parse_parameters_custom_function(self):
        parameters = [
            {"app_version": "${gen_app_version()}"},
            {"username-password": "${get_account()}"}
        ]
        testset_path = os.path.join(
            os.getcwd(),
            "tests/data/demo_parameters.yml"
        )
        cartesian_product_parameters = context.parse_parameters(
            parameters,
            testset_path
        )
        self.assertEqual(
            len(cartesian_product_parameters),
            2 * 2
        )

    def test_parse_parameters_mix(self):
        parameters = [
            {"user_agent": ["iOS/10.1", "iOS/10.2", "iOS/10.3"]},
            {"app_version": "${gen_app_version()}"},
            {"username-password": "${parameterize(account.csv)}"}
        ]
        testset_path = os.path.join(
            os.getcwd(),
            "tests/data/demo_parameters.yml"
        )
        cartesian_product_parameters = context.parse_parameters(
            parameters,
            testset_path
        )
        self.assertEqual(
            len(cartesian_product_parameters),
            3 * 2 * 3
        )
