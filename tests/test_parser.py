import os
import time
import unittest

from httprunner import exceptions, parser


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

    def test_parse_parameters_raw_list(self):
        parameters = [
            {"user_agent": ["iOS/10.1", "iOS/10.2", "iOS/10.3"]},
            {"username-password": [("user1", "111111"), ["test2", "222222"]]}
        ]
        cartesian_product_parameters = parser.parse_parameters(parameters)
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
        cartesian_product_parameters = parser.parse_parameters(
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
        cartesian_product_parameters = parser.parse_parameters(
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
        cartesian_product_parameters = parser.parse_parameters(
            parameters,
            testset_path
        )
        self.assertEqual(
            len(cartesian_product_parameters),
            3 * 2 * 3
        )

    def test_parse_data(self):
        content = {
            'request': {
                'url': '/api/users/$uid',
                'method': "$method",
                'headers': {'token': '$token'},
                'data': {
                    "null": None,
                    "true": True,
                    "false": False,
                    "empty_str": ""
                }
            }
        }
        mapping = {
            "$uid": 1000,
            "$method": "POST"
        }
        result = parser.parse_data(content, mapping)
        self.assertEqual("/api/users/1000", result["request"]["url"])
        self.assertEqual("$token", result["request"]["headers"]["token"])
        self.assertEqual("POST", result["request"]["method"])
        self.assertIsNone(result["request"]["data"]["null"])
        self.assertTrue(result["request"]["data"]["true"])
        self.assertFalse(result["request"]["data"]["false"])
        self.assertEqual("", result["request"]["data"]["empty_str"])


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
        testcase_parser = parser.TestcaseParser(variables=variables)
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
        testcase_parser = parser.TestcaseParser()

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
        testcase_parser = parser.TestcaseParser(variables=variables)
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
        testcase_parser = parser.TestcaseParser(variables=variables)
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
        testcase_parser = parser.TestcaseParser(variables=variables)
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
        testcase_parser = parser.TestcaseParser(functions=functions)

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
        testcase_parser = parser.TestcaseParser(functions=functions)
        self.assertEqual(
            testcase_parser._eval_content_functions("${add_two_nums(1, 2)}"),
            3
        )
        self.assertEqual(
            testcase_parser._eval_content_functions("/api/${add_two_nums(1, 2)}"),
            "/api/3"
        )

    def test_eval_content_functions_search_upward(self):
        testcase_parser = parser.TestcaseParser()

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
        parsed_testcase = parser.TestcaseParser(variables, functions)\
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
