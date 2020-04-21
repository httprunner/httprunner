import time
import unittest

from httprunner.v3 import parser
from httprunner.v3.exceptions import VariableNotFound, FunctionNotFound


class TestParserBasic(unittest.TestCase):

    def test_parse_variables_mapping(self):
        variables = {
            "varA": "$varB",
            "varB": "$varC",
            "varC": "123",
            "a": 1,
            "b": 2
        }
        parsed_variables = parser.parse_variables_mapping(variables)
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
            parser.parse_variables_mapping(variables)

    def test_parse_string_value(self):
        self.assertEqual(parser.parse_string_value("123"), 123)
        self.assertEqual(parser.parse_string_value("12.3"), 12.3)
        self.assertEqual(parser.parse_string_value("a123"), "a123")
        self.assertEqual(parser.parse_string_value("$var"), "$var")
        self.assertEqual(parser.parse_string_value("${func}"), "${func}")

    def test_extract_variables(self):
        self.assertEqual(
            parser.extract_variables("$var"),
            {"var"}
        )
        self.assertEqual(
            parser.extract_variables("$var123"),
            {"var123"}
        )
        self.assertEqual(
            parser.extract_variables("$var_name"),
            {"var_name"}
        )
        self.assertEqual(
            parser.extract_variables("var"),
            set()
        )
        self.assertEqual(
            parser.extract_variables("a$var"),
            {"var"}
        )
        self.assertEqual(
            parser.extract_variables("$v ar"),
            {"v"}
        )
        self.assertEqual(
            parser.extract_variables(" "),
            set()
        )
        self.assertEqual(
            parser.extract_variables("$abc*"),
            {"abc"}
        )
        self.assertEqual(
            parser.extract_variables("${func()}"),
            set()
        )
        self.assertEqual(
            parser.extract_variables("${func(1,2)}"),
            set()
        )
        self.assertEqual(
            parser.extract_variables("${gen_md5($TOKEN, $data, $random)}"),
            {"TOKEN", "data", "random"}
        )

    def test_parse_function_params(self):
        self.assertEqual(
            parser.parse_function_params(""),
            {'args': [], 'kwargs': {}}
        )
        self.assertEqual(
            parser.parse_function_params("5"),
            {'args': [5], 'kwargs': {}}
        )
        self.assertEqual(
            parser.parse_function_params("1, 2"),
            {'args': [1, 2], 'kwargs': {}}
        )
        self.assertEqual(
            parser.parse_function_params("a=1, b=2"),
            {'args': [], 'kwargs': {'a': 1, 'b': 2}}
        )
        self.assertEqual(
            parser.parse_function_params("a= 1, b =2"),
            {'args': [], 'kwargs': {'a': 1, 'b': 2}}
        )
        self.assertEqual(
            parser.parse_function_params("1, 2, a=3, b=4"),
            {'args': [1, 2], 'kwargs': {'a': 3, 'b': 4}}
        )
        self.assertEqual(
            parser.parse_function_params("$request, 123"),
            {'args': ["$request", 123], 'kwargs': {}}
        )
        self.assertEqual(
            parser.parse_function_params("  "),
            {'args': [], 'kwargs': {}}
        )
        self.assertEqual(
            parser.parse_function_params("hello world, a=3, b=4"),
            {'args': ["hello world"], 'kwargs': {'a': 3, 'b': 4}}
        )
        self.assertEqual(
            parser.parse_function_params("$request, 12 3"),
            {'args': ["$request", '12 3'], 'kwargs': {}}
        )

    def test_extract_functions(self):
        self.assertEqual(
            parser.regex_findall_functions("${func()}"),
            [("func", "")]
        )
        self.assertEqual(
            parser.regex_findall_functions("${func(5)}"),
            [("func", "5")]
        )
        self.assertEqual(
            parser.regex_findall_functions("${func(a=1, b=2)}"),
            [("func", "a=1, b=2")]
        )
        self.assertEqual(
            parser.regex_findall_functions("${func(1, $b, c=$x, d=4)}"),
            [("func", "1, $b, c=$x, d=4")]
        )
        self.assertEqual(
            parser.regex_findall_functions("/api/1000?_t=${get_timestamp()}"),
            [("get_timestamp", "")]
        )
        self.assertEqual(
            parser.regex_findall_functions("/api/${add(1, 2)}"),
            [("add", "1, 2")]
        )
        self.assertEqual(
            parser.regex_findall_functions("/api/${add(1, 2)}?_t=${get_timestamp()}"),
            [('add', '1, 2'), ('get_timestamp', '')]
        )
        self.assertEqual(
            parser.regex_findall_functions("abc${func(1, 2, a=3, b=4)}def"),
            [('func', '1, 2, a=3, b=4')]
        )

    def test_parse_content(self):
        content = {
            'request': {
                'url': '/api/users/$uid',
                'method': "$method",
                'headers': {'token': '$token'},
                'data': {
                    "null": None,
                    "true": True,
                    "false": False,
                    "empty_str": "",
                    "value": "abc${add_one(3)}def"
                }
            }
        }
        variables_mapping = {
            "uid": 1000,
            "method": "POST",
            "token": "abc123"
        }
        functions_mapping = {
            "add_one": lambda x: x + 1
        }
        result = parser.parse_content(content, variables_mapping, functions_mapping)
        self.assertEqual("/api/users/1000", result["request"]["url"])
        self.assertEqual("abc123", result["request"]["headers"]["token"])
        self.assertEqual("POST", result["request"]["method"])
        self.assertIsNone(result["request"]["data"]["null"])
        self.assertTrue(result["request"]["data"]["true"])
        self.assertFalse(result["request"]["data"]["false"])
        self.assertEqual("", result["request"]["data"]["empty_str"])
        self.assertEqual("abc4def", result["request"]["data"]["value"])

    def test_parse_content_with_variables(self):
        variables_mapping = {
            "var_1": "abc",
            "var_2": "def",
            "var_3": 123,
            "var_4": {"a": 1},
            "var_5": True,
            "var_6": None
        }
        self.assertEqual(
            parser.parse_content("$var_1", variables_mapping),
            "abc"
        )
        self.assertEqual(
            parser.parse_content("${var_1}", variables_mapping),
            "abc"
        )
        self.assertEqual(
            parser.parse_content("var_1", variables_mapping),
            "var_1"
        )
        self.assertEqual(
            parser.parse_content("$var_1#XYZ", variables_mapping),
            "abc#XYZ"
        )
        self.assertEqual(
            parser.parse_content("${var_1}#XYZ", variables_mapping),
            "abc#XYZ"
        )
        self.assertEqual(
            parser.parse_content("/$var_1/$var_2/var3", variables_mapping),
            "/abc/def/var3"
        )
        self.assertEqual(
            parser.parse_content("$var_3", variables_mapping),
            123
        )
        self.assertEqual(
            parser.parse_content("$var_4", variables_mapping),
            {"a": 1}
        )
        self.assertEqual(
            parser.parse_content("$var_5", variables_mapping),
            True
        )
        self.assertEqual(
            parser.parse_content("abc$var_5", variables_mapping),
            "abcTrue"
        )
        self.assertEqual(
            parser.parse_content("abc$var_4", variables_mapping),
            "abc{'a': 1}"
        )
        self.assertEqual(
            parser.parse_content("$var_6", variables_mapping),
            None
        )

        with self.assertRaises(VariableNotFound):
            parser.parse_content("/api/$SECRET_KEY", variables_mapping)

        self.assertEqual(
            parser.parse_content(["$var_1", "$var_2"], variables_mapping),
            ["abc", "def"]
        )
        self.assertEqual(
            parser.parse_content({"$var_1": "$var_2"}, variables_mapping),
            {"abc": "def"}
        )

    def test_parse_data_multiple_identical_variables(self):
        variables_mapping = {
            "var_1": "abc",
            "var_2": "def",
        }
        self.assertEqual(
            parser.parse_content("/$var_1/$var_2/$var_1", variables_mapping),
            "/abc/def/abc"
        )

        variables_mapping = {
            "userid": 100,
            "data": 1498
        }
        content = "/users/$userid/training/$data?userId=$userid&data=$data"
        self.assertEqual(
            parser.parse_content(content, variables_mapping),
            "/users/100/training/1498?userId=100&data=1498"
        )

        variables_mapping = {
            "user": 100,
            "userid": 1000,
            "data": 1498
        }
        content = "/users/$user/$userid/$data?userId=$userid&data=$data"
        self.assertEqual(
            parser.parse_content(content, variables_mapping),
            "/users/100/1000/1498?userId=1000&data=1498"
        )

    def test_parse_data_functions(self):
        import random, string
        functions_mapping = {
            "gen_random_string": lambda str_len: ''.join(random.choice(string.ascii_letters + string.digits) \
                for _ in range(str_len))
        }
        result = parser.parse_content("${gen_random_string(5)}", functions_mapping=functions_mapping)
        self.assertEqual(len(result), 5)

        add_two_nums = lambda a, b=1: a + b
        functions_mapping["add_two_nums"] = add_two_nums
        self.assertEqual(
            parser.parse_content("${add_two_nums(1)}", functions_mapping=functions_mapping),
            2
        )
        self.assertEqual(
            parser.parse_content("${add_two_nums(1, 2)}", functions_mapping=functions_mapping),
            3
        )
        self.assertEqual(
            parser.parse_content("/api/${add_two_nums(1, 2)}", functions_mapping=functions_mapping),
            "/api/3"
        )

        with self.assertRaises(FunctionNotFound):
            parser.parse_content("/api/${gen_md5(abc)}")

    def test_parse_data_testcase(self):
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
        parsed_testcase = parser.parse_content(testcase_template, variables, functions)
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
