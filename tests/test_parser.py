import os
import time
import unittest

from httprunner import exceptions, loader, parser
from httprunner.loader import load
from tests.debugtalk import gen_random_string, sum_two


class TestParserBasic(unittest.TestCase):

    def test_parse_string_value(self):
        self.assertEqual(parser.parse_string_value("123"), 123)
        self.assertEqual(parser.parse_string_value("12.3"), 12.3)
        self.assertEqual(parser.parse_string_value("a123"), "a123")
        self.assertEqual(parser.parse_string_value("$var"), "$var")
        self.assertEqual(parser.parse_string_value("${func}"), "${func}")

    def test_regex_findall_variables(self):
        self.assertEqual(
            parser.regex_findall_variables("$var"),
            ["var"]
        )
        self.assertEqual(
            parser.regex_findall_variables("$var123"),
            ["var123"]
        )
        self.assertEqual(
            parser.regex_findall_variables("$var_name"),
            ["var_name"]
        )
        self.assertEqual(
            parser.regex_findall_variables("var"),
            []
        )
        self.assertEqual(
            parser.regex_findall_variables("a$var"),
            ["var"]
        )
        self.assertEqual(
            parser.regex_findall_variables("a$var${var2}$var3${var4}"),
            ["var", "var2", "var3", "var4"]
        )
        self.assertEqual(
            parser.regex_findall_variables("$v ar"),
            ["v"]
        )
        self.assertEqual(
            parser.regex_findall_variables(" "),
            []
        )
        self.assertEqual(
            parser.regex_findall_variables("$abc*"),
            ["abc"]
        )
        self.assertEqual(
            parser.regex_findall_variables("${func()}"),
            []
        )
        self.assertEqual(
            parser.regex_findall_variables("${func(1,2)}"),
            []
        )
        self.assertEqual(
            parser.regex_findall_variables("${gen_md5($TOKEN, $data, $random)}"),
            ["TOKEN", "data", "random"]
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

    def test_extract_variables(self):
        prepared_content = parser.prepare_lazy_data("123$a", {}, {"a"})
        self.assertEqual(
            parser.extract_variables(prepared_content),
            {"a"}
        )
        prepared_content = parser.prepare_lazy_data("$a$b", {}, {"a", "b"})
        self.assertEqual(
            parser.extract_variables(prepared_content),
            {"a", "b"}
        )
        prepared_content = parser.prepare_lazy_data(["$a$b", "$c", "d"], {}, {"a", "b", "c", "d"})
        self.assertEqual(
            parser.extract_variables(prepared_content),
            {"a", "b", "c"}
        )
        prepared_content = parser.prepare_lazy_data(
            {"a": 1, "b": {"c": "$d", "e": 3}},
            {},
            {"d"}
        )
        self.assertEqual(
            parser.extract_variables(prepared_content),
            {"d"}
        )
        prepared_content = parser.prepare_lazy_data(
            {"a": ["$b"], "b": {"c": "$d", "e": 3}},
            {},
            {"b", "d"}
        )
        self.assertEqual(
            parser.extract_variables(prepared_content),
            {"b", "d"}
        )
        prepared_content = parser.prepare_lazy_data(
            ["$a$b", "$c", {"c": "$d"}],
            {},
            {"a", "b", "c", "d"}
        )
        self.assertEqual(
            parser.extract_variables(prepared_content),
            {"a", "b", "c", "d"}
        )

    def test_extract_functions(self):
        self.assertEqual(
            parser.regex_findall_functions("${func()}"),
            [('func', '')]
        )
        self.assertEqual(
            parser.regex_findall_functions("${func(5)}"),
            [('func', '5')]
        )
        self.assertEqual(
            parser.regex_findall_functions("${func(a=1, b=2)}"),
            [('func', 'a=1, b=2')]
        )
        self.assertEqual(
            parser.regex_findall_functions("${func(1, $b, c=$x, d=4)}"),
            [('func', '1, $b, c=$x, d=4')]
        )
        self.assertEqual(
            parser.regex_findall_functions("/api/1000?_t=${get_timestamp()}"),
            [('get_timestamp', '')]
        )
        self.assertEqual(
            parser.regex_findall_functions("/api/${add(1, 2)}"),
            [('add', '1, 2')]
        )
        self.assertEqual(
            parser.regex_findall_functions("/api/${add(1, 2)}?_t=${get_timestamp()}"),
            [('add', '1, 2'), ('get_timestamp', '')]
        )
        self.assertEqual(
            parser.regex_findall_functions("abc${func(1, 2, a=3, b=4)}def"),
            [('func', '1, 2, a=3, b=4')]
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
        result = parser.eval_lazy_data(content, variables_mapping, functions_mapping)
        self.assertEqual("/api/users/1000", result["request"]["url"])
        self.assertEqual("abc123", result["request"]["headers"]["token"])
        self.assertEqual("POST", result["request"]["method"])
        self.assertIsNone(result["request"]["data"]["null"])
        self.assertTrue(result["request"]["data"]["true"])
        self.assertFalse(result["request"]["data"]["false"])
        self.assertEqual("", result["request"]["data"]["empty_str"])
        self.assertEqual("abc4def", result["request"]["data"]["value"])

    def test_eval_lazy_data(self):
        variables_mapping = {
            "var_1": "abc",
            "var_2": "def",
            "var_3": 123,
            "var_4": {"a": 1},
            "var_5": True,
            "var_6": None
        }
        self.assertEqual(
            parser.eval_lazy_data("$var_1", variables_mapping=variables_mapping),
            "abc"
        )
        self.assertEqual(
            parser.eval_lazy_data("var_1", variables_mapping=variables_mapping),
            "var_1"
        )
        self.assertEqual(
            parser.eval_lazy_data("$var_1#XYZ", variables_mapping=variables_mapping),
            "abc#XYZ"
        )
        self.assertEqual(
            parser.eval_lazy_data("/$var_1/$var_2/var3", variables_mapping=variables_mapping),
            "/abc/def/var3"
        )
        self.assertEqual(
            parser.eval_lazy_data("/$var_1/$var_2/$var_1", variables_mapping=variables_mapping),
            "/abc/def/abc"
        )
        self.assertEqual(
            parser.eval_lazy_data("$var_3", variables_mapping=variables_mapping),
            123
        )
        self.assertEqual(
            parser.eval_lazy_data("$var_4", variables_mapping=variables_mapping),
            {"a": 1}
        )
        self.assertEqual(
            parser.eval_lazy_data("$var_5", variables_mapping=variables_mapping),
            True
        )
        self.assertEqual(
            parser.eval_lazy_data("abc$var_5", variables_mapping=variables_mapping),
            "abcTrue"
        )
        self.assertEqual(
            parser.eval_lazy_data("abc$var_4", variables_mapping=variables_mapping),
            "abc{'a': 1}"
        )
        self.assertEqual(
            parser.eval_lazy_data("$var_6", variables_mapping=variables_mapping),
            None
        )

        with self.assertRaises(exceptions.VariableNotFound):
            parser.eval_lazy_data("/api/$SECRET_KEY", variables_mapping=variables_mapping)

        self.assertEqual(
            parser.eval_lazy_data(["$var_1", "$var_2"], variables_mapping=variables_mapping),
            ["abc", "def"]
        )
        self.assertEqual(
            parser.eval_lazy_data({"$var_1": "$var_2"}, variables_mapping=variables_mapping),
            {"abc": "def"}
        )

    def test_parse_func_var_abnormal(self):
        variables_mapping = {
            "var_1": "abc",
            "var_2": "def",
            "var_3": 123,
            "var_4": {"a": 1},
            "var_5": True,
            "var_6": None
        }
        check_variables_set = variables_mapping.keys()
        functions_mapping = {
            "func1": lambda x,y: str(x) + str(y)
        }

        # {
        var = parser.LazyString("ABC$var_1{", functions_mapping, check_variables_set)
        self.assertEqual(var._string, "ABC{}{{")
        self.assertEqual(var._args, ["var_1"])
        self.assertEqual(var.to_value(variables_mapping), "ABCabc{")

        var = parser.LazyString("{ABC$var_1{}a}", functions_mapping, check_variables_set)
        self.assertEqual(var._string, "{{ABC{}{{}}a}}")
        self.assertEqual(var._args, ["var_1"])
        self.assertEqual(var.to_value(variables_mapping), "{ABCabc{}a}")

        var = parser.LazyString("AB{C$var_1{}a}", functions_mapping, check_variables_set)
        self.assertEqual(var._string, "AB{{C{}{{}}a}}")
        self.assertEqual(var._args, ["var_1"])
        self.assertEqual(var.to_value(variables_mapping), "AB{Cabc{}a}")

        # }
        var = parser.LazyString("ABC$var_1}", functions_mapping, check_variables_set)
        self.assertEqual(var._string, "ABC{}}}")
        self.assertEqual(var._args, ["var_1"])
        self.assertEqual(var.to_value(variables_mapping), "ABCabc}")

        # $$
        var = parser.LazyString("ABC$$var_1{", functions_mapping, check_variables_set)
        self.assertEqual(var._string, "ABC$var_1{{")
        self.assertEqual(var._args, [])
        self.assertEqual(var.to_value(variables_mapping), "ABC$var_1{")

        # $$$
        var = parser.LazyString("ABC$$$var_1{", functions_mapping, check_variables_set)
        self.assertEqual(var._string, "ABC${}{{")
        self.assertEqual(var._args, ["var_1"])
        self.assertEqual(var.to_value(variables_mapping), "ABC$abc{")

        # $$$$
        var = parser.LazyString("ABC$$$$var_1{", functions_mapping, check_variables_set)
        self.assertEqual(var._string, "ABC$$var_1{{")
        self.assertEqual(var._args, [])
        self.assertEqual(var.to_value(variables_mapping), "ABC$$var_1{")

        # ${
        var = parser.LazyString("ABC$var_1${", functions_mapping, check_variables_set)
        self.assertEqual(var._string, "ABC{}${{")
        self.assertEqual(var._args, ["var_1"])
        self.assertEqual(var.to_value(variables_mapping), "ABCabc${")

        var = parser.LazyString("ABC$var_1${a", functions_mapping, check_variables_set)
        self.assertEqual(var._string, "ABC{}${{a")
        self.assertEqual(var._args, ["var_1"])
        self.assertEqual(var.to_value(variables_mapping), "ABCabc${a")

        # $}
        var = parser.LazyString("ABC$var_1$}a", functions_mapping, check_variables_set)
        self.assertEqual(var._string, "ABC{}$}}a")
        self.assertEqual(var._args, ["var_1"])
        self.assertEqual(var.to_value(variables_mapping), "ABCabc$}a")

        # }{
        var = parser.LazyString("ABC$var_1}{a", functions_mapping, check_variables_set)
        self.assertEqual(var._string, "ABC{}}}{{a")
        self.assertEqual(var._args, ["var_1"])
        self.assertEqual(var.to_value(variables_mapping), "ABCabc}{a")

        # {}
        var = parser.LazyString("ABC$var_1{}a", functions_mapping, check_variables_set)
        self.assertEqual(var._string, "ABC{}{{}}a")
        self.assertEqual(var._args, ["var_1"])
        self.assertEqual(var.to_value(variables_mapping), "ABCabc{}a")

    def test_parse_func_var_duplicate(self):
        variables_mapping = {
            "var_1": "abc",
            "var_2": "def",
            "var_3": 123,
            "var_4": {"a": 1},
            "var_5": True,
            "var_6": None
        }
        check_variables_set = variables_mapping.keys()
        functions_mapping = {
            "func1": lambda x,y: str(x) + str(y)
        }
        var = parser.LazyString(
            "ABC${func1($var_1, $var_3)}--${func1($var_1, $var_3)}",
            functions_mapping,
            check_variables_set
        )
        self.assertEqual(var._string, "ABC{}--{}")
        self.assertEqual(var.to_value(variables_mapping), "ABCabc123--abc123")

        var = parser.LazyString("ABC${func1($var_1, $var_3)}$var_1", functions_mapping, check_variables_set)
        self.assertEqual(var._string, "ABC{}{}")
        self.assertEqual(var.to_value(variables_mapping), "ABCabc123abc")

        var = parser.LazyString(
            "ABC${func1($var_1, $var_3)}$var_1--${func1($var_1, $var_3)}$var_1",
            functions_mapping,
            check_variables_set
        )
        self.assertEqual(var._string, "ABC{}{}--{}{}")
        self.assertEqual(var.to_value(variables_mapping), "ABCabc123abc--abc123abc")

    def test_parse_function(self):
        variables_mapping = {
            "var_1": "abc",
            "var_2": "def",
            "var_3": 123,
            "var_4": {"a": 1},
            "var_5": True,
            "var_6": None
        }
        check_variables_set = variables_mapping.keys()
        functions_mapping = {
            "func1": lambda x,y: str(x) + str(y)
        }

        var = parser.LazyString("${func1($var_1, $var_3)}", functions_mapping, check_variables_set)
        self.assertEqual(var._string, "{}")
        self.assertIsInstance(var._args[0], parser.LazyFunction)
        self.assertEqual(var.to_value(variables_mapping), "abc123")

        var = parser.LazyString("ABC${func1($var_1, $var_3)}DE", functions_mapping, check_variables_set)
        self.assertEqual(var._string, "ABC{}DE")
        self.assertIsInstance(var._args[0], parser.LazyFunction)
        self.assertEqual(var.to_value(variables_mapping), "ABCabc123DE")

        var = parser.LazyString("ABC${func1($var_1, $var_3)}$var_5", functions_mapping, check_variables_set)
        self.assertEqual(var._string, "ABC{}{}")
        self.assertEqual(var.to_value(variables_mapping), "ABCabc123True")

        var = parser.LazyString("ABC${func1($var_1, $var_3)}DE$var_4", functions_mapping, check_variables_set)
        self.assertEqual(var._string, "ABC{}DE{}")
        self.assertEqual(var.to_value(variables_mapping), "ABCabc123DE{'a': 1}")

        var = parser.LazyString("ABC$var_5${func1($var_1, $var_3)}", functions_mapping, check_variables_set)
        self.assertEqual(var._string, "ABC{}{}")
        self.assertEqual(var.to_value(variables_mapping), "ABCTrueabc123")

        # Python builtin functions
        var = parser.LazyString("ABC${ord(a)}DEF${len(abcd)}", functions_mapping, check_variables_set)
        self.assertEqual(var._string, "ABC{}DEF{}")
        self.assertEqual(var.to_value(variables_mapping), "ABC97DEF4")

    def test_parse_variable(self):
        """ variable format ${var} and $var
        """
        variables_mapping = {
            "var_1": "abc",
            "var_2": "def",
            "var_3": 123,
            "var_4": {"a": 1},
            "var_5": True,
            "var_6": None
        }
        check_variables_set = variables_mapping.keys()
        functions_mapping = {}

        # format: $var
        var = parser.LazyString("ABC$var_1", functions_mapping, check_variables_set)
        self.assertEqual(var._string, "ABC{}")
        self.assertEqual(var._args, ["var_1"])
        self.assertEqual(var.to_value(variables_mapping), "ABCabc")

        var = parser.LazyString("ABC$var_1$var_3", functions_mapping, check_variables_set)
        self.assertEqual(var._string, "ABC{}{}")
        self.assertEqual(var._args, ["var_1", "var_3"])
        self.assertEqual(var.to_value(variables_mapping), "ABCabc123")

        var = parser.LazyString("ABC$var_1/$var_3", functions_mapping, check_variables_set)
        self.assertEqual(var._string, "ABC{}/{}")
        self.assertEqual(var._args, ["var_1", "var_3"])
        self.assertEqual(var.to_value(variables_mapping), "ABCabc/123")

        var = parser.LazyString("ABC$var_1/", functions_mapping, check_variables_set)
        self.assertEqual(var._string, "ABC{}/")
        self.assertEqual(var._args, ["var_1"])
        self.assertEqual(var.to_value(variables_mapping), "ABCabc/")

        var = parser.LazyString("ABC$var_1$", functions_mapping, check_variables_set)
        self.assertEqual(var._string, "ABC{}$")
        self.assertEqual(var._args, ["var_1"])
        self.assertEqual(var.to_value(variables_mapping), "ABCabc$")

        var = parser.LazyString("ABC$var_1/123$var_1/456", functions_mapping, check_variables_set)
        self.assertEqual(var._string, "ABC{}/123{}/456")
        self.assertEqual(var._args, ["var_1", "var_1"])
        self.assertEqual(var.to_value(variables_mapping), "ABCabc/123abc/456")

        var = parser.LazyString("ABC$var_1/$var_2/$var_1", functions_mapping, check_variables_set)
        self.assertEqual(var._string, "ABC{}/{}/{}")
        self.assertEqual(var._args, ["var_1", "var_2", "var_1"])
        self.assertEqual(var.to_value(variables_mapping), "ABCabc/def/abc")

        var = parser.LazyString("func1($var_1, $var_3)", functions_mapping, check_variables_set)
        self.assertEqual(var._string, "func1({}, {})")
        self.assertEqual(var._args, ["var_1", "var_3"])
        self.assertEqual(var.to_value(variables_mapping), "func1(abc, 123)")

        # format: ${var}
        var = parser.LazyString("ABC${var_1}", functions_mapping, check_variables_set)
        self.assertEqual(var._string, "ABC{}")
        self.assertEqual(var._args, ["var_1"])
        self.assertEqual(var.to_value(variables_mapping), "ABCabc")

        var = parser.LazyString("ABC${var_1}${var_3}", functions_mapping, check_variables_set)
        self.assertEqual(var._string, "ABC{}{}")
        self.assertEqual(var._args, ["var_1", "var_3"])
        self.assertEqual(var.to_value(variables_mapping), "ABCabc123")

        var = parser.LazyString("ABC${var_1}/${var_3}", functions_mapping, check_variables_set)
        self.assertEqual(var._string, "ABC{}/{}")
        self.assertEqual(var._args, ["var_1", "var_3"])
        self.assertEqual(var.to_value(variables_mapping), "ABCabc/123")

        var = parser.LazyString("ABC${var_1}/", functions_mapping, check_variables_set)
        self.assertEqual(var._string, "ABC{}/")
        self.assertEqual(var._args, ["var_1"])
        self.assertEqual(var.to_value(variables_mapping), "ABCabc/")

        var = parser.LazyString("ABC${var_1}123", functions_mapping, check_variables_set)
        self.assertEqual(var._string, "ABC{}123")
        self.assertEqual(var._args, ["var_1"])
        self.assertEqual(var.to_value(variables_mapping), "ABCabc123")

        var = parser.LazyString("ABC${var_1}/123${var_1}/456", functions_mapping, check_variables_set)
        self.assertEqual(var._string, "ABC{}/123{}/456")
        self.assertEqual(var._args, ["var_1", "var_1"])
        self.assertEqual(var.to_value(variables_mapping), "ABCabc/123abc/456")

        var = parser.LazyString("ABC${var_1}/${var_2}/${var_1}", functions_mapping, check_variables_set)
        self.assertEqual(var._string, "ABC{}/{}/{}")
        self.assertEqual(var._args, ["var_1", "var_2", "var_1"])
        self.assertEqual(var.to_value(variables_mapping), "ABCabc/def/abc")

        var = parser.LazyString("func1(${var_1}, ${var_3})", functions_mapping, check_variables_set)
        self.assertEqual(var._string, "func1({}, {})")
        self.assertEqual(var._args, ["var_1", "var_3"])
        self.assertEqual(var.to_value(variables_mapping), "func1(abc, 123)")

    def test_parse_data_multiple_identical_variables(self):
        variables_mapping = {
            "userid": 100,
            "data": 1498
        }
        content = "/users/$userid/training/$data?userId=$userid&data=$data"
        self.assertEqual(
            parser.eval_lazy_data(content, variables_mapping=variables_mapping),
            "/users/100/training/1498?userId=100&data=1498"
        )

        variables_mapping = {
            "user": 100,
            "userid": 1000,
            "data": 1498
        }
        content = "/users/$user/$userid/$data?userId=$userid&data=$data"
        self.assertEqual(
            parser.eval_lazy_data(content, variables_mapping=variables_mapping),
            "/users/100/1000/1498?userId=1000&data=1498"
        )

    def test_parse_data_functions(self):
        functions_mapping = {
            "gen_random_string": gen_random_string
        }
        result = parser.eval_lazy_data("${gen_random_string(5)}", functions_mapping=functions_mapping)
        self.assertEqual(len(result), 5)

        add_two_nums = lambda a, b=1: a + b
        functions_mapping["add_two_nums"] = add_two_nums
        self.assertEqual(
            parser.eval_lazy_data("${add_two_nums(1)}", functions_mapping=functions_mapping),
            2
        )
        self.assertEqual(
            parser.eval_lazy_data("${add_two_nums(1, 2)}", functions_mapping=functions_mapping),
            3
        )
        self.assertEqual(
            parser.eval_lazy_data("/api/${add_two_nums(1, 2)}", functions_mapping=functions_mapping),
            "/api/3"
        )

        with self.assertRaises(exceptions.FunctionNotFound):
            parser.eval_lazy_data("/api/${gen_md5(abc)}", functions_mapping=functions_mapping)

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
        parsed_testcase = parser.eval_lazy_data(
            testcase_template,
            variables_mapping=variables,
            functions_mapping=functions
        )
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

    def test_parse_variables_mapping(self):
        variables = {
            "varA": "123$varB",
            "varB": "456$varC",
            "varC": "${sum_two($a, $b)}",
            "a": 1,
            "b": 2
        }
        functions = {
            "sum_two": sum_two
        }
        prepared_variables = parser.prepare_lazy_data(variables, functions, variables.keys())
        parsed_variables = parser.parse_variables_mapping(prepared_variables)
        self.assertEqual(parsed_variables["varA"], "1234563")
        self.assertEqual(parsed_variables["varB"], "4563")
        self.assertEqual(parsed_variables["varC"], 3)

    def test_parse_variables_mapping_fix_duplicate_function_call(self):
        # fix duplicate function calling
        variables = {
            "varA": "$varB",
            "varB": "${gen_random_string(5)}"
        }
        functions = {
            "gen_random_string": gen_random_string
        }
        prepared_variables = parser.prepare_lazy_data(variables, functions, variables.keys())
        parsed_variables = parser.parse_variables_mapping(prepared_variables)
        self.assertEqual(parsed_variables["varA"], parsed_variables["varB"])

    def test_parse_variables_mapping_dead_circle(self):
        variables = {
            "varA": "$varB",
            "varB": "123$varC"
        }
        check_variables_set = {"varA", "varB", "varC"}
        prepared_variables = parser.prepare_lazy_data(variables, {}, check_variables_set)
        with self.assertRaises(exceptions.VariableNotFound):
            parser.parse_variables_mapping(prepared_variables)

    def test_parse_variables_mapping_not_found(self):
        variables = {
            "varA": "123$varB",
            "varB": "456$varC",
            "varC": "${sum_two($a, $b)}",
            "b": 2
        }
        functions = {
            "sum_two": sum_two
        }
        with self.assertRaises(exceptions.VariableNotFound):
            parser.prepare_lazy_data(variables, functions, variables.keys())

    def test_parse_variables_mapping_ref_self(self):
        variables = {
            "varC": "${sum_two($a, $b)}",
            "a": 1,
            "b": 2,
            "token": "$token"
        }
        functions = {
            "sum_two": sum_two
        }
        prepared_variables = parser.prepare_lazy_data(variables, functions, variables.keys())
        with self.assertRaises(exceptions.VariableNotFound):
            parser.parse_variables_mapping(prepared_variables)

    def test_parse_variables_mapping_2(self):
        variables = {
            "host2": "https://httprunner.org",
            "num3": "${sum_two($num2, 4)}",
            "num2": "${sum_two($num1, 3)}",
            "num1": "${sum_two(1, 2)}"
        }
        functions = {
            "sum_two": sum_two
        }
        prepared_variables = parser.prepare_lazy_data(variables, functions, variables.keys())
        parsed_testcase = parser.parse_variables_mapping(prepared_variables)
        self.assertEqual(parsed_testcase["num3"], 10)
        self.assertEqual(parsed_testcase["num2"], 6)
        self.assertEqual(parsed_testcase["num1"], 3)

    def test_is_var_or_func_exist(self):
        self.assertTrue(parser.is_var_or_func_exist("$var"))
        self.assertTrue(parser.is_var_or_func_exist("${var}"))
        self.assertTrue(parser.is_var_or_func_exist("$var${var}"))
        self.assertFalse(parser.is_var_or_func_exist("${var"))
        self.assertFalse(parser.is_var_or_func_exist("$$var"))
        self.assertFalse(parser.is_var_or_func_exist("var$$0"))
        self.assertTrue(parser.is_var_or_func_exist("var$$$0"))
        self.assertFalse(parser.is_var_or_func_exist("var$$$$0"))
        self.assertTrue(parser.is_var_or_func_exist("${func()}"))
        self.assertTrue(parser.is_var_or_func_exist("${func($a)}"))
        self.assertTrue(parser.is_var_or_func_exist("${func($a)}$b"))

    def test_parse_variables_mapping_dollar_notation(self):
        variables = {
            "varA": "123$varB",
            "varB": "456$$0",
            "varC": "${sum_two($a, $b)}",
            "a": 1,
            "b": 2,
            "c": "abc"
        }
        functions = {
            "sum_two": sum_two
        }
        prepared_variables = parser.prepare_lazy_data(variables, functions, variables.keys())
        parsed_testcase = parser.parse_variables_mapping(prepared_variables)
        self.assertEqual(parsed_testcase["varA"], "123456$0")
        self.assertEqual(parsed_testcase["varB"], "456$0")
        self.assertEqual(parsed_testcase["varC"], 3)

    def test_prepare_lazy_data(self):
        variables = {
            "host": "https://httprunner.org",
            "num4": "${sum_two($num0, 5)}",
            "num3": "${sum_two($num2, 4)}",
            "num2": "${sum_two($num1, 3)}",
            "num1": "${sum_two(1, 2)}",
            "num0": 0
        }
        functions = {
            "sum_two": sum_two
        }
        parser.prepare_lazy_data(
            variables,
            functions,
            variables.keys()
        )

    def test_prepare_lazy_data_not_found(self):
        variables = {
            "host": "https://httprunner.org",
            "num4": "${sum_two($num0, 5)}",
            "num3": "${sum_two($num2, 4)}",
            "num2": "${sum_two($num1, 3)}",
            "num1": "${sum_two(1, 2)}"
        }
        functions = {
            "sum_two": sum_two
        }
        with self.assertRaises(exceptions.VariableNotFound):
            parser.prepare_lazy_data(
                variables,
                functions,
                variables.keys()
            )

    def test_prepare_lazy_data_dual_dollar(self):
        variables = {
            "num0": 123,
            "var1": "abc$$num0",
            "var2": "abc$$$num0",
            "var3": "abc$$$$num0",
        }
        functions = {
            "sum_two": sum_two
        }
        prepared_variables = parser.prepare_lazy_data(
            variables,
            functions,
            variables.keys()
        )
        self.assertEqual(prepared_variables["var1"], "abc$num0")
        self.assertIsInstance(prepared_variables["var2"], parser.LazyString)
        self.assertEqual(prepared_variables["var3"], "abc$$num0")

        parsed_variables = parser.parse_variables_mapping(prepared_variables)
        self.assertEqual(parsed_variables["var1"], "abc$num0")
        self.assertEqual(parsed_variables["var2"], "abc$123")
        self.assertEqual(parsed_variables["var3"], "abc$$num0")

    def test_get_uniform_comparator(self):
        self.assertEqual(parser.get_uniform_comparator("eq"), "equals")
        self.assertEqual(parser.get_uniform_comparator("=="), "equals")
        self.assertEqual(parser.get_uniform_comparator("lt"), "less_than")
        self.assertEqual(parser.get_uniform_comparator("le"), "less_than_or_equals")
        self.assertEqual(parser.get_uniform_comparator("gt"), "greater_than")
        self.assertEqual(parser.get_uniform_comparator("ge"), "greater_than_or_equals")
        self.assertEqual(parser.get_uniform_comparator("ne"), "not_equals")

        self.assertEqual(parser.get_uniform_comparator("str_eq"), "string_equals")
        self.assertEqual(parser.get_uniform_comparator("len_eq"), "length_equals")
        self.assertEqual(parser.get_uniform_comparator("count_eq"), "length_equals")

        self.assertEqual(parser.get_uniform_comparator("len_gt"), "length_greater_than")
        self.assertEqual(parser.get_uniform_comparator("count_gt"), "length_greater_than")
        self.assertEqual(parser.get_uniform_comparator("count_greater_than"), "length_greater_than")

        self.assertEqual(parser.get_uniform_comparator("len_ge"), "length_greater_than_or_equals")
        self.assertEqual(parser.get_uniform_comparator("count_ge"), "length_greater_than_or_equals")
        self.assertEqual(parser.get_uniform_comparator("count_greater_than_or_equals"), "length_greater_than_or_equals")

        self.assertEqual(parser.get_uniform_comparator("len_lt"), "length_less_than")
        self.assertEqual(parser.get_uniform_comparator("count_lt"), "length_less_than")
        self.assertEqual(parser.get_uniform_comparator("count_less_than"), "length_less_than")

        self.assertEqual(parser.get_uniform_comparator("len_le"), "length_less_than_or_equals")
        self.assertEqual(parser.get_uniform_comparator("count_le"), "length_less_than_or_equals")
        self.assertEqual(parser.get_uniform_comparator("count_less_than_or_equals"), "length_less_than_or_equals")

    def test_parse_validator(self):
        _validator = {"check": "status_code", "comparator": "eq", "expect": 201}
        self.assertEqual(
            parser.uniform_validator(_validator),
            {"check": "status_code", "comparator": "equals", "expect": 201}
        )

        _validator = {'eq': ['status_code', 201]}
        self.assertEqual(
            parser.uniform_validator(_validator),
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
            parser.uniform_validator(_validator)
            for _validator in def_validators
        ]
        ref_validators = [
            parser.uniform_validator(_validator)
            for _validator in current_validators
        ]

        extended_validators = parser.extend_validators(def_validators, ref_validators)
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
            parser.uniform_validator(_validator)
            for _validator in def_validators
        ]
        ref_validators = [
            parser.uniform_validator(_validator)
            for _validator in current_validators
        ]

        extended_validators = parser.extend_validators(def_validators, ref_validators)
        self.assertEqual(len(extended_validators), 3)
        self.assertIn({'check': {'b': 1}, 'expect': 201, 'comparator': 'equals'}, extended_validators)
        self.assertNotIn({'check': {'b': 1}, 'expect': 200, 'comparator': 'equals'}, extended_validators)


class TestParser(unittest.TestCase):

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

    def test_parse_parameters_custom_function(self):
        parameters = [
            {"user_agent": "${get_user_agent()}"},
            {"app_version": "${gen_app_version()}"},
            {"username-password": "${get_account()}"},
            {"username2-password2": "${get_account_in_tuple()}"}
        ]
        dot_env_path = os.path.join(
            os.getcwd(), "tests", ".env"
        )
        load.load_dot_env_file(dot_env_path)
        from tests import debugtalk
        cartesian_product_parameters = parser.parse_parameters(
            parameters,
            functions_mapping=load.load_module_functions(debugtalk)
        )
        self.assertIn(
            {
                'user_agent': 'iOS/10.1',
                'app_version': '2.8.5',
                'username': 'user1',
                'password': '111111',
                'username2': 'user1',
                'password2': '111111'
            },
            cartesian_product_parameters
        )
        self.assertEqual(
            len(cartesian_product_parameters),
            2 * 2 * 2 * 2
        )

    def test_parse_parameters_parameterize(self):
        loader.load_project_data(os.path.join(os.getcwd(), "tests"))
        parameters = [
            {"app_version": "${parameterize(data/app_version.csv)}"},
            {"username-password": "${parameterize(data/account.csv)}"}
        ]
        cartesian_product_parameters = parser.parse_parameters(parameters)
        self.assertEqual(
            len(cartesian_product_parameters),
            2 * 3
        )

    def test_parse_parameters_mix(self):
        project_mapping = loader.load_project_data(os.path.join(os.getcwd(), "tests"))

        parameters = [
            {"user_agent": ["iOS/10.1", "iOS/10.2", "iOS/10.3"]},
            {"app_version": "${gen_app_version()}"},
            {"username-password": "${parameterize(data/account.csv)}"}
        ]
        cartesian_product_parameters = parser.parse_parameters(
            parameters, functions_mapping=project_mapping["functions"])
        self.assertEqual(
            len(cartesian_product_parameters),
            3 * 2 * 3
        )

    def test_parse_tests_testcase(self):
        testcase_file_path = os.path.join(
            os.getcwd(), 'tests/data/demo_testcase.yml')
        tests_mapping = loader.load_cases(testcase_file_path)
        testcases = tests_mapping["testcases"]
        self.assertEqual(
            testcases[0]["config"]["variables"]["var_c"],
            "${sum_two($var_a, $var_b)}"
        )
        self.assertEqual(
            testcases[0]["config"]["variables"]["PROJECT_KEY"],
            "${ENV(PROJECT_KEY)}"
        )
        parsed_testcases = parser.parse_tests(tests_mapping)
        self.assertIsInstance(parsed_testcases, list)
        test_dict1 = parsed_testcases[0]["teststeps"][0]
        self.assertEqual(test_dict1["variables"]["var_c"].raw_string, "${sum_two($var_a, $var_b)}")
        self.assertEqual(test_dict1["variables"]["PROJECT_KEY"].raw_string, "${ENV(PROJECT_KEY)}")
        self.assertIsInstance(parsed_testcases[0]["config"]["name"], parser.LazyString)

    def test_parse_tests_override_variables(self):
        tests_mapping = {
            'testcases': [
                {
                    "config": {
                        'name': '',
                        'variables': [
                            {"password": "123456"},
                            {"creator": "user_test_001"}
                        ]
                    },
                    "teststeps": [
                        {
                            'name': 'testcase1',
                            "variables": [
                                {"creator": "user_test_002"},
                                {"username": "$creator"}
                            ],
                            'request': {'url': '/api1', 'method': 'GET'}
                        }
                    ]
                }
            ]
        }
        parsed_testcases = parser.parse_tests(tests_mapping)
        test_dict1_variables = parsed_testcases[0]["teststeps"][0]["variables"]
        self.assertEqual(test_dict1_variables["creator"], "user_test_001")
        self.assertEqual(test_dict1_variables["username"].raw_string, "$creator")

    def test_parse_tests_base_url_priority(self):
        """ base_url & verify: priority test_dict > config
        """
        tests_mapping = {
            'testcases': [
                {
                    "config": {
                        'name': '',
                        "base_url": "$host",
                        'variables': {
                            "host": "https://github.com"
                        },
                        "verify": False
                    },
                    "teststeps": [
                        {
                            'name': 'testcase1',
                            "base_url": "https://httprunner.org",
                            'request': {'url': '/api1', 'method': 'GET', "verify": True}
                        }
                    ]
                }
            ]
        }
        parsed_testcases = parser.parse_tests(tests_mapping)
        test_dict = parsed_testcases[0]["teststeps"][0]
        self.assertEqual(test_dict["request"]["url"], "/api1")
        self.assertEqual(test_dict["request"]["verify"], True)

    def test_parse_tests_base_url_path_with_variable(self):
        tests_mapping = {
            'testcases': [
                {
                    "config": {
                        'name': '',
                        "base_url": "$host1",
                        'variables': {
                            "host1": "https://github.com"
                        }
                    },
                    "teststeps": [
                        {
                            'name': 'testcase1',
                            "variables": {
                                "host2": "https://httprunner.org"
                            },
                            'request': {'url': '$host2/api1', 'method': 'GET'}
                        }
                    ]
                }
            ]
        }
        parsed_testcases = parser.parse_tests(tests_mapping)
        test_dict = parsed_testcases[0]["teststeps"][0]
        self.assertEqual(test_dict["variables"]["host2"], "https://httprunner.org")
        parsed_test_dict = parser.parse_lazy_data(test_dict, test_dict["variables"])
        self.assertEqual(parsed_test_dict["request"]["url"], "https://httprunner.org/api1")

    def test_parse_tests_base_url_test_dict(self):
        tests_mapping = {
            'testcases': [
                {
                    "config": {
                        'name': '',
                        "base_url": "$host1",
                        'variables': {
                            "host1": "https://github.com"
                        }
                    },
                    "teststeps": [
                        {
                            'name': 'testcase1',
                            "base_url": "$host2",
                            "variables": {
                                "host2": "https://httprunner.org"
                            },
                            'request': {'url': '/api1', 'method': 'GET'}
                        }
                    ]
                }
            ]
        }
        parsed_testcases = parser.parse_tests(tests_mapping)
        test_dict = parsed_testcases[0]["teststeps"][0]
        parsed_test_dict = parser.parse_lazy_data(test_dict, test_dict["variables"])
        self.assertEqual(parsed_test_dict["base_url"], "https://httprunner.org")

    def test_parse_tests_variable_with_function(self):
        tests_mapping = {
            "project_mapping": {
                "functions": {
                    "sum_two": sum_two,
                    "gen_random_string": gen_random_string
                }
            },
            'testcases': [
                {
                    "config": {
                        'name': '',
                        "base_url": "$host1",
                        'variables': {
                            "host1": "https://github.com",
                            "var_a": "${gen_random_string(5)}",
                            "var_b": "$var_a"
                        }
                    },
                    "teststeps": [
                        {
                            'name': 'testcase1',
                            "base_url": "$host2",
                            "variables": {
                                "host2": "https://httprunner.org",
                                "num3": "${sum_two($num2, 4)}",
                                "num2": "${sum_two($num1, 3)}",
                                "num1": "${sum_two(1, 2)}",
                                "str1": "${gen_random_string(5)}",
                                "str2": "$str1"
                            },
                            'request': {
                                'url': '/api1/?num1=$num1&num2=$num2&num3=$num3',
                                'method': 'GET'
                            }
                        }
                    ]
                }
            ]
        }
        parsed_testcases = parser.parse_tests(tests_mapping)
        test_dict = parsed_testcases[0]["teststeps"][0]
        variables = parser.parse_variables_mapping(test_dict["variables"])
        self.assertEqual(variables["num3"], 10)
        self.assertEqual(variables["num2"], 6)
        parsed_test_dict = parser.parse_lazy_data(test_dict, variables)
        self.assertEqual(parsed_test_dict["base_url"], "https://httprunner.org")
        self.assertEqual(
            parsed_test_dict["request"]["url"],
            "/api1/?num1=3&num2=6&num3=10"
        )
        self.assertEqual(variables["str1"], variables["str2"])

    def test_parse_tests_variable_not_found(self):
        tests_mapping = {
            "project_mapping": {
                "functions": {
                    "sum_two": sum_two
                }
            },
            'testcases': [
                {
                    "config": {
                        'name': '',
                        "base_url": "$host1",
                        'variables': {
                            "host1": "https://github.com"
                        }
                    },
                    "teststeps": [
                        {
                            'name': 'testcase1',
                            "base_url": "$host2",
                            "variables": {
                                "host2": "https://httprunner.org",
                                "num4": "${sum_two($num0, 5)}",
                                "num3": "${sum_two($num2, 4)}",
                                "num2": "${sum_two($num1, 3)}",
                                "num1": "${sum_two(1, 2)}"
                            },
                            'request': {
                                'url': '/api1/?num1=$num1&num2=$num2&num3=$num3&num4=$num4',
                                'method': 'GET'
                            }
                        }
                    ]
                }
            ]
        }
        parser.parse_tests(tests_mapping)
        parse_failed_testfiles = parser.get_parse_failed_testfiles()
        self.assertIn("testcase", parse_failed_testfiles)

    def test_parse_tests_base_url_teststep_empty(self):
        """ base_url & verify: priority test_dict > config
        """
        tests_mapping = {
            'testcases': [
                {
                    "config": {
                        'name': '',
                        "base_url": "$host",
                        'variables': {
                            "host": "https://github.com"
                        },
                        "verify": False
                    },
                    "teststeps": [
                        {
                            'name': 'testcase1',
                            "base_url": "",
                            'request': {'url': '/api1', 'method': 'GET', "verify": True}
                        }
                    ]
                }
            ]
        }
        parsed_testcases = parser.parse_tests(tests_mapping)
        test_dict = parsed_testcases[0]["teststeps"][0]
        self.assertEqual(str(test_dict["base_url"]), 'LazyString($host)')
        self.assertEqual(test_dict["request"]["verify"], True)

    def test_parse_tests_verify_config_set(self):
        """ verify priority: test_dict > config
        """
        tests_mapping = {
            'testcases': [
                {
                    "config": {
                        'name': 'bugfix verify',
                        "base_url": "https://httpbin.org/",
                        "verify": False
                    },
                    "teststeps": [
                        {
                            'name': 'testcase1',
                            'request': {'url': '/headers', 'method': 'GET'}
                        }
                    ]
                }
            ]
        }
        parsed_testcases = parser.parse_tests(tests_mapping)
        test_dict = parsed_testcases[0]["teststeps"][0]
        self.assertEqual(test_dict["request"]["verify"], False)

    def test_parse_tests_verify_config_unset(self):
        """ verify priority: test_dict > config
        """
        tests_mapping = {
            'testcases': [
                {
                    "config": {
                        'name': 'bugfix verify',
                        "base_url": "https://httpbin.org/",
                    },
                    "teststeps": [
                        {
                            'name': 'testcase1',
                            'request': {'url': '/headers', 'method': 'GET'}
                        }
                    ]
                }
            ]
        }
        parsed_testcases = parser.parse_tests(tests_mapping)
        test_dict = parsed_testcases[0]["teststeps"][0]
        self.assertEqual(test_dict["request"]["verify"], True)

    def test_parse_tests_verify_step_set_false(self):
        """ verify priority: test_dict > config
        """
        tests_mapping = {
            'testcases': [
                {
                    "config": {
                        'name': 'bugfix verify',
                        "base_url": "https://httpbin.org/",
                        "verify": True
                    },
                    "teststeps": [
                        {
                            'name': 'testcase1',
                            'request': {'url': '/headers', 'method': 'GET', "verify": False}
                        }
                    ]
                }
            ]
        }
        parsed_testcases = parser.parse_tests(tests_mapping)
        test_dict = parsed_testcases[0]["teststeps"][0]
        self.assertEqual(test_dict["request"]["verify"], False)

    def test_parse_tests_verify_nested_testcase_unset(self):
        tests_mapping = {
            'testcases': [
                {
                    'config': {
                        'name': 'inquiry price',
                        'verify': False
                    },
                    'teststeps': [
                        {
                            'name': 'login system',
                            'testcase': 'testcases/deps/login.yml',
                            'testcase_def': {
                                'config': {
                                    'name': 'login system'
                                },
                                'teststeps': [
                                    {
                                        'name': '/',
                                        'request': {
                                            'method': 'GET',
                                            'url': 'https://httpbin.org/'
                                        }
                                    }
                                ]
                            }
                        }
                    ]
                }
            ]
        }
        parsed_testcases = parser.parse_tests(tests_mapping)
        test_dict = parsed_testcases[0]["teststeps"][0]
        self.assertEqual(test_dict["teststeps"][0]["request"]["verify"], False)

    def test_parse_environ(self):
        os.environ["PROJECT_KEY"] = "ABCDEFGH"
        content = {
            "variables": [
                {"PROJECT_KEY": "${ENV(PROJECT_KEY)}"}
            ]
        }
        result = parser.eval_lazy_data(content)

        content = {
            "variables": [
                {"PROJECT_KEY": "${ENV(PROJECT_KEY, abc)}"}
            ]
        }
        with self.assertRaises(exceptions.ParamsError):
            parser.eval_lazy_data(content)

        content = {
            "variables": [
                {"PROJECT_KEY": "${ENV(abc=123)}"}
            ]
        }
        with self.assertRaises(exceptions.ParamsError):
            parser.eval_lazy_data(content)

    def test_extend_with_api(self):
        loader.load_project_data(os.path.join(os.getcwd(), "tests"))
        raw_testinfo = {
            "name": "get token",
            "base_url": "https://github.com",
            "api": "api/get_token.yml",
        }
        api_def_dict = loader.buildup.load_teststep(raw_testinfo)
        test_block = {
            "name": "override block",
            "times": 3,
            "variables": [
                {"var": 123}
            ],
            "base_url": "https://httprunner.org",
            'request': {
                'url': '/api/get-token',
                'method': 'POST',
                'headers': {'user_agent': '$user_agent', 'device_sn': '$device_sn', 'os_platform': '$os_platform', 'app_version': '$app_version'},
                'json': {'sign': '${get_sign($device_sn, $os_platform, $app_version)}'}
            },
            'validate': [
                {"check": "status_code", "comparator": "equals", "expect": 201},
                {"check": "content.token", "comparator": "length_equals", "expect": 32}
            ]
        }

        parser._extend_with_api(test_block, api_def_dict)
        self.assertEqual(test_block["base_url"], "https://github.com")
        self.assertEqual(test_block["name"], "override block")
        self.assertEqual({'var': 123}, test_block["variables"])
        self.assertIn({'check': 'status_code', 'expect': 201, 'comparator': 'equals'}, test_block["validate"])
        self.assertIn({'check': 'content.token', 'comparator': 'length_equals', 'expect': 32}, test_block["validate"])
        self.assertEqual(test_block["times"], 3)
