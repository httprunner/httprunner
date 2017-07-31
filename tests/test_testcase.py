import unittest

from ate.exception import ParamsError
from ate import testcase


class TestcaseParserUnittest(unittest.TestCase):

    def test_get_contain_variables(self):
        self.assertEqual(
            testcase.get_contain_variables("$var"),
            ["var"]
        )
        self.assertEqual(
            testcase.get_contain_variables("$var123"),
            ["var123"]
        )
        self.assertEqual(
            testcase.get_contain_variables("$var_name"),
            ["var_name"]
        )
        self.assertEqual(
            testcase.get_contain_variables("var"),
            []
        )
        self.assertEqual(
            testcase.get_contain_variables("a$var"),
            ["var"]
        )
        self.assertEqual(
            testcase.get_contain_variables("$v ar"),
            ["v"]
        )
        self.assertEqual(
            testcase.get_contain_variables(" "),
            []
        )
        self.assertEqual(
            testcase.get_contain_variables("$abc*"),
            ["abc"]
        )
        self.assertEqual(
            testcase.get_contain_variables("${func()}"),
            []
        )
        self.assertEqual(
            testcase.get_contain_variables("${func(1,2)}"),
            []
        )
        self.assertEqual(
            testcase.get_contain_variables("${gen_md5($TOKEN, $data, $random)}"),
            ["TOKEN", "data", "random"]
        )

    def test_parse_variables(self):
        variable_mapping = {
            "var_1": "abc",
            "var_2": "def",
            "var_3": 123,
            "var_4": {"a": 1},
            "var_5": True
        }
        self.assertEqual(
            testcase.parse_variables("$var_1", variable_mapping),
            "abc"
        )
        self.assertEqual(
            testcase.parse_variables("var_1", variable_mapping),
            "var_1"
        )
        self.assertEqual(
            testcase.parse_variables("$var_1#XYZ", variable_mapping),
            "abc#XYZ"
        )
        self.assertEqual(
            testcase.parse_variables("/$var_1/$var_2/var3", variable_mapping),
            "/abc/def/var3"
        )
        self.assertEqual(
            testcase.parse_variables("/$var_1/$var_2/$var_1", variable_mapping),
            "/abc/def/abc"
        )
        self.assertEqual(
            testcase.parse_variables("${func($var_1, $var_2, xyz)}", variable_mapping),
            "${func(abc, def, xyz)}"
        )
        self.assertEqual(
            testcase.parse_variables("$var_3", variable_mapping),
            123
        )
        self.assertEqual(
            testcase.parse_variables("$var_4", variable_mapping),
            {"a": 1}
        )
        self.assertEqual(
            testcase.parse_variables("$var_5", variable_mapping),
            True
        )
        self.assertEqual(
            testcase.parse_variables("abc$var_5", variable_mapping),
            "abcTrue"
        )
        self.assertEqual(
            testcase.parse_variables("abc$var_4", variable_mapping),
            "abc{'a': 1}"
        )

    def test_is_functon(self):
        content = "${func()}"
        self.assertTrue(testcase.is_functon(content))
        content = "${func(5)}"
        self.assertTrue(testcase.is_functon(content))
        content = "${func(1, 2)}"
        self.assertTrue(testcase.is_functon(content))
        content = "${func($a, $b)}"
        self.assertTrue(testcase.is_functon(content))
        content = "${func(a=1, b=2)}"
        self.assertTrue(testcase.is_functon(content))
        content = "${func(1, 2, a=3, b=4)}"
        self.assertTrue(testcase.is_functon(content))
        content = "${func(1, $b, c=$x, d=4)}"
        self.assertTrue(testcase.is_functon(content))
        content = "${func}"
        self.assertFalse(testcase.is_functon(content))
        content = "$abc"
        self.assertFalse(testcase.is_functon(content))
        content = "abc"
        self.assertFalse(testcase.is_functon(content))

    def test_parse_string_value(self):
        str_value = "123"
        self.assertEqual(testcase.parse_string_value(str_value), 123)
        str_value = "12.3"
        self.assertEqual(testcase.parse_string_value(str_value), 12.3)
        str_value = "a123"
        self.assertEqual(testcase.parse_string_value(str_value), "a123")
        str_value = "$var"
        self.assertEqual(testcase.parse_string_value(str_value), "$var")
        str_value = "${func}"
        self.assertEqual(testcase.parse_string_value(str_value), "${func}")

    def test_parse_functon(self):
        content = "${func()}"
        self.assertEqual(
            testcase.parse_function(content),
            {'func_name': 'func', 'args': [], 'kwargs': {}}
        )
        content = "${func(5)}"
        self.assertEqual(
            testcase.parse_function(content),
            {'func_name': 'func', 'args': [5], 'kwargs': {}}
        )
        content = "${func(1, 2)}"
        self.assertEqual(
            testcase.parse_function(content),
            {'func_name': 'func', 'args': [1, 2], 'kwargs': {}}
        )
        content = "${func(a=1, b=2)}"
        self.assertEqual(
            testcase.parse_function(content),
            {'func_name': 'func', 'args': [], 'kwargs': {'a': 1, 'b': 2}}
        )
        content = "${func(a= 1, b =2)}"
        self.assertEqual(
            testcase.parse_function(content),
            {'func_name': 'func', 'args': [], 'kwargs': {'a': 1, 'b': 2}}
        )
        content = "${func(1, 2, a=3, b=4)}"
        self.assertEqual(
            testcase.parse_function(content),
            {'func_name': 'func', 'args': [1, 2], 'kwargs': {'a': 3, 'b': 4}}
        )

    def test_parse_content_with_bindings_variables(self):
        variables_binds = {
            "str_1": "str_value1",
            "str_2": "str_value2"
        }
        self.assertEqual(
            testcase.parse_content_with_bindings("$str_1", variables_binds, {}),
            "str_value1"
        )
        self.assertEqual(
            testcase.parse_content_with_bindings("123$str_1/456", variables_binds, {}),
            "123str_value1/456"
        )

        with self.assertRaises(ParamsError):
            testcase.parse_content_with_bindings("$str_3", variables_binds, {})

        self.assertEqual(
            testcase.parse_content_with_bindings(["$str_1", "str3"], variables_binds, {}),
            ["str_value1", "str3"]
        )
        self.assertEqual(
            testcase.parse_content_with_bindings({"key": "$str_1"}, variables_binds, {}),
            {"key": "str_value1"}
        )

    def test_parse_content_with_bindings_multiple_identical_variables(self):
        variables_binds = {
            "userid": 100,
            "data": 1498
        }
        content = "/users/$userid/training/$data?userId=$userid&data=$data"
        self.assertEqual(
            testcase.parse_content_with_bindings(content, variables_binds, {}),
            "/users/100/training/1498?userId=100&data=1498"
        )

    def test_parse_content_with_bindings_functions(self):
        import random, string
        functions_binds = {
            "gen_random_string": lambda str_len: ''.join(random.choice(string.ascii_letters + string.digits) \
                for _ in range(str_len))
        }

        result = testcase.parse_content_with_bindings("${gen_random_string(5)}", {}, functions_binds)
        self.assertEqual(len(result), 5)

        add_two_nums = lambda a, b=1: a + b
        functions_binds["add_two_nums"] = add_two_nums
        self.assertEqual(
            testcase.parse_content_with_bindings("${add_two_nums(1)}", {}, functions_binds),
            2
        )
        self.assertEqual(
            testcase.parse_content_with_bindings("${add_two_nums(1, 2)}", {}, functions_binds),
            3
        )

    def test_parse_content_with_bindings_testcase(self):
        variables_binds = {
            "uid": "1000",
            "random": "A2dEx",
            "authorization": "a83de0ff8d2e896dbd8efb81ba14e17d",
            "data": {"name": "user", "password": "123456"},
            "expected_status": 201,
            "expected_success": True
        }
        functions_binds = {
            "add_two_nums": lambda a, b=1: a + b
        }
        testcase_template = {
            "url": "http://127.0.0.1:5000/api/users/$uid",
            "method": "POST",
            "headers": {
                "Content-Type": "application/json",
                "authorization": "$authorization",
                "random": "$random",
                "sum": "${add_two_nums(1, 2)}"
            },
            "body": "$data"
        }
        parsed_testcase = testcase.parse_content_with_bindings(
            testcase_template, variables_binds, functions_binds)

        self.assertEqual(
            parsed_testcase["url"],
            "http://127.0.0.1:5000/api/users/%s" % variables_binds["uid"]
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
