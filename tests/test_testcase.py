import unittest

from ate.exception import ParamsError
from ate.testcase import parse_content_with_bindings


class TestcaseParserUnittest(unittest.TestCase):

    def test_parse_content_with_bindings_variables(self):
        variables_binds = {
            "str_1": "str_value1",
            "str_2": "str_value2"
        }
        self.assertEqual(
            parse_content_with_bindings("$str_1", variables_binds, {}),
            "str_value1"
        )
        self.assertEqual(
            parse_content_with_bindings("123$str_1/456", variables_binds, {}),
            "123str_value1/456"
        )

        with self.assertRaises(ParamsError):
            parse_content_with_bindings("$str_3", variables_binds, {})

        self.assertEqual(
            parse_content_with_bindings(["$str_1", "str3"], variables_binds, {}),
            ["str_value1", "str3"]
        )
        self.assertEqual(
            parse_content_with_bindings({"key": "$str_1"}, variables_binds, {}),
            {"key": "str_value1"}
        )

    def test_parse_content_with_bindings_multiple_identical_variables(self):
        variables_binds = {
            "userid": 100,
            "data": 1498
        }
        content = "/users/$userid/training/$data?userId=$userid&data=$data"
        self.assertEqual(
            parse_content_with_bindings(content, variables_binds, {}),
            "/users/100/training/1498?userId=100&data=1498"
        )

    def test_parse_content_with_bindings_functions(self):
        import random, string
        functions_binds = {
            "gen_random_string": lambda str_len: ''.join(random.choice(string.ascii_letters + string.digits) \
                for _ in range(str_len))
        }

        result = parse_content_with_bindings("${gen_random_string(5)}", {}, functions_binds)
        self.assertEqual(len(result), 5)

        add_two_nums = lambda a, b=1: a + b
        functions_binds["add_two_nums"] = add_two_nums
        self.assertEqual(
            parse_content_with_bindings("${add_two_nums(1)}", {}, functions_binds),
            2
        )
        self.assertEqual(
            parse_content_with_bindings("${add_two_nums(1, 2)}", {}, functions_binds),
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
        testcase = {
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
        parsed_testcase = parse_content_with_bindings(testcase, variables_binds, functions_binds)

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
