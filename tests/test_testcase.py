import time
import unittest

from ate import testcase
from ate.exception import ParamsError


class TestcaseParserUnittest(unittest.TestCase):

    def test_extract_variables(self):
        self.assertEqual(
            testcase.extract_variables("$var"),
            ["var"]
        )
        self.assertEqual(
            testcase.extract_variables("$var123"),
            ["var123"]
        )
        self.assertEqual(
            testcase.extract_variables("$var_name"),
            ["var_name"]
        )
        self.assertEqual(
            testcase.extract_variables("var"),
            []
        )
        self.assertEqual(
            testcase.extract_variables("a$var"),
            ["var"]
        )
        self.assertEqual(
            testcase.extract_variables("$v ar"),
            ["v"]
        )
        self.assertEqual(
            testcase.extract_variables(" "),
            []
        )
        self.assertEqual(
            testcase.extract_variables("$abc*"),
            ["abc"]
        )
        self.assertEqual(
            testcase.extract_variables("${func()}"),
            []
        )
        self.assertEqual(
            testcase.extract_variables("${func(1,2)}"),
            []
        )
        self.assertEqual(
            testcase.extract_variables("${gen_md5($TOKEN, $data, $random)}"),
            ["TOKEN", "data", "random"]
        )

    def test_eval_content_variables(self):
        variable_mapping = {
            "var_1": "abc",
            "var_2": "def",
            "var_3": 123,
            "var_4": {"a": 1},
            "var_5": True,
            "var_6": None
        }
        self.assertEqual(
            testcase.eval_content_variables("$var_1", variable_mapping),
            "abc"
        )
        self.assertEqual(
            testcase.eval_content_variables("var_1", variable_mapping),
            "var_1"
        )
        self.assertEqual(
            testcase.eval_content_variables("$var_1#XYZ", variable_mapping),
            "abc#XYZ"
        )
        self.assertEqual(
            testcase.eval_content_variables("/$var_1/$var_2/var3", variable_mapping),
            "/abc/def/var3"
        )
        self.assertEqual(
            testcase.eval_content_variables("/$var_1/$var_2/$var_1", variable_mapping),
            "/abc/def/abc"
        )
        self.assertEqual(
            testcase.eval_content_variables("${func($var_1, $var_2, xyz)}", variable_mapping),
            "${func(abc, def, xyz)}"
        )
        self.assertEqual(
            testcase.eval_content_variables("$var_3", variable_mapping),
            123
        )
        self.assertEqual(
            testcase.eval_content_variables("$var_4", variable_mapping),
            {"a": 1}
        )
        self.assertEqual(
            testcase.eval_content_variables("$var_5", variable_mapping),
            True
        )
        self.assertEqual(
            testcase.eval_content_variables("abc$var_5", variable_mapping),
            "abcTrue"
        )
        self.assertEqual(
            testcase.eval_content_variables("abc$var_4", variable_mapping),
            "abc{'a': 1}"
        )
        self.assertEqual(
            testcase.eval_content_variables("$var_6", variable_mapping),
            None
        )

    def test_parse_string_value(self):
        self.assertEqual(testcase.parse_string_value("123"), 123)
        self.assertEqual(testcase.parse_string_value("12.3"), 12.3)
        self.assertEqual(testcase.parse_string_value("a123"), "a123")
        self.assertEqual(testcase.parse_string_value("$var"), "$var")
        self.assertEqual(testcase.parse_string_value("${func}"), "${func}")

    def test_parse_functon(self):
        self.assertEqual(
            testcase.parse_function("${func()}"),
            {'func_name': 'func', 'args': [], 'kwargs': {}}
        )
        self.assertEqual(
            testcase.parse_function("${func(5)}"),
            {'func_name': 'func', 'args': [5], 'kwargs': {}}
        )
        self.assertEqual(
            testcase.parse_function("${func(1, 2)}"),
            {'func_name': 'func', 'args': [1, 2], 'kwargs': {}}
        )
        self.assertEqual(
            testcase.parse_function("${func(a=1, b=2)}"),
            {'func_name': 'func', 'args': [], 'kwargs': {'a': 1, 'b': 2}}
        )
        self.assertEqual(
            testcase.parse_function("${func(a= 1, b =2)}"),
            {'func_name': 'func', 'args': [], 'kwargs': {'a': 1, 'b': 2}}
        )
        self.assertEqual(
            testcase.parse_function("${func(1, 2, a=3, b=4)}"),
            {'func_name': 'func', 'args': [1, 2], 'kwargs': {'a': 3, 'b': 4}}
        )

    def test_parse_content_with_bindings_variables(self):
        variables_binds = {
            "str_1": "str_value1",
            "str_2": "str_value2"
        }
        testcase_parser = testcase.TestcaseParser(variables_binds=variables_binds)
        self.assertEqual(
            testcase_parser.parse_content_with_bindings("$str_1"),
            "str_value1"
        )
        self.assertEqual(
            testcase_parser.parse_content_with_bindings("123$str_1/456"),
            "123str_value1/456"
        )

        with self.assertRaises(ParamsError):
            testcase_parser.parse_content_with_bindings("$str_3")

        self.assertEqual(
            testcase_parser.parse_content_with_bindings(["$str_1", "str3"]),
            ["str_value1", "str3"]
        )
        self.assertEqual(
            testcase_parser.parse_content_with_bindings({"key": "$str_1"}),
            {"key": "str_value1"}
        )

    def test_parse_content_with_bindings_multiple_identical_variables(self):
        variables_binds = {
            "userid": 100,
            "data": 1498
        }
        testcase_parser = testcase.TestcaseParser(variables_binds=variables_binds)
        content = "/users/$userid/training/$data?userId=$userid&data=$data"
        self.assertEqual(
            testcase_parser.parse_content_with_bindings(content),
            "/users/100/training/1498?userId=100&data=1498"
        )

    def test_parse_variables_multiple_identical_variables(self):
        variables_binds = {
            "user": 100,
            "userid": 1000,
            "data": 1498
        }
        testcase_parser = testcase.TestcaseParser(variables_binds=variables_binds)
        content = "/users/$user/$userid/$data?userId=$userid&data=$data"
        self.assertEqual(
            testcase_parser.parse_content_with_bindings(content),
            "/users/100/1000/1498?userId=1000&data=1498"
        )

    def test_parse_content_with_bindings_functions(self):
        import random, string
        functions_binds = {
            "gen_random_string": lambda str_len: ''.join(random.choice(string.ascii_letters + string.digits) \
                for _ in range(str_len))
        }
        testcase_parser = testcase.TestcaseParser(functions_binds=functions_binds)

        result = testcase_parser.parse_content_with_bindings("${gen_random_string(5)}")
        self.assertEqual(len(result), 5)

        add_two_nums = lambda a, b=1: a + b
        functions_binds["add_two_nums"] = add_two_nums
        self.assertEqual(
            testcase_parser.parse_content_with_bindings("${add_two_nums(1)}"),
            2
        )
        self.assertEqual(
            testcase_parser.parse_content_with_bindings("${add_two_nums(1, 2)}"),
            3
        )

    def test_extract_functions(self):
        self.assertEqual(
            testcase.extract_functions("${func()}"),
            ["${func()}"]
        )
        self.assertEqual(
            testcase.extract_functions("${func(5)}"),
            ["${func(5)}"]
        )
        self.assertEqual(
            testcase.extract_functions("${func(a=1, b=2)}"),
            ["${func(a=1, b=2)}"]
        )
        self.assertEqual(
            testcase.extract_functions("${func(1, $b, c=$x, d=4)}"),
            ["${func(1, $b, c=$x, d=4)}"]
        )
        self.assertEqual(
            testcase.extract_functions("/api/1000?_t=${get_timestamp()}"),
            ["${get_timestamp()}"]
        )
        self.assertEqual(
            testcase.extract_functions("/api/${add(1, 2)}"),
            ["${add(1, 2)}"]
        )
        self.assertEqual(
            testcase.extract_functions("/api/${add(1, 2)}?_t=${get_timestamp()}"),
            ["${add(1, 2)}", "${get_timestamp()}"]
        )
        self.assertEqual(
            testcase.extract_functions("abc${func(1, 2, a=3, b=4)}def"),
            ["${func(1, 2, a=3, b=4)}"]
        )

    def test_eval_content_functions(self):
        functions_binds = {
            "add_two_nums": lambda a, b=1: a + b
        }
        testcase_parser = testcase.TestcaseParser(functions_binds=functions_binds)
        self.assertEqual(
            testcase_parser.eval_content_functions("${add_two_nums(1, 2)}"),
            3
        )
        self.assertEqual(
            testcase_parser.eval_content_functions("/api/${add_two_nums(1, 2)}"),
            "/api/3"
        )

    def test_eval_content_functions_search_upward(self):
        testcase_parser = testcase.TestcaseParser()

        with self.assertRaises(ParamsError):
            testcase_parser.eval_content_functions("/api/${gen_md5(abc)}")

        testcase_parser.file_path = "tests/data/demo_testset_hardcode.yml"
        content = testcase_parser.eval_content_functions("/api/${gen_md5(abc)}")
        self.assertEqual(content, "/api/900150983cd24fb0d6963f7d28e17f72")

    def test_parse_content_with_bindings_testcase(self):
        variables_binds = {
            "uid": "1000",
            "random": "A2dEx",
            "authorization": "a83de0ff8d2e896dbd8efb81ba14e17d",
            "data": {"name": "user", "password": "123456"}
        }
        functions_binds = {
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
        parsed_testcase = testcase.TestcaseParser(variables_binds, functions_binds)\
            .parse_content_with_bindings(testcase_template)

        self.assertEqual(
            parsed_testcase["url"],
            "http://127.0.0.1:5000/api/users/1000/3"
        )
        self.assertEqual(
            parsed_testcase["headers"]["authorization"],
            variables_binds["authorization"]
        )
        self.assertEqual(
            parsed_testcase["headers"]["random"],
            variables_binds["random"]
        )
        self.assertEqual(
            parsed_testcase["body"],
            variables_binds["data"]
        )
        self.assertEqual(
            parsed_testcase["headers"]["sum"],
            3
        )
