import os
import unittest

from httprunner import compat, exceptions, loader
from httprunner.utils import HTTP_BIN_URL


class TestCompat(unittest.TestCase):
    def setUp(self):
        loader.project_meta = None

    def test_convert_variables(self):
        raw_variables = {"var1": 1, "var2": "val2"}
        self.assertEqual(
            compat.convert_variables(raw_variables, "examples/data/a-b.c/1.yml"),
            {"var1": 1, "var2": "val2"},
        )
        raw_variables = "${get_variables()}"
        self.assertEqual(
            compat.convert_variables(raw_variables, "examples/data/a-b.c/1.yml"),
            {"foo1": "session_bar1"},
        )

        with self.assertRaises(exceptions.TestCaseFormatError):
            raw_variables = [{"var1": 1}, {"var2": "val2", "var3": 3}]
            compat.convert_variables(raw_variables, "examples/data/a-b.c/1.yml")
        with self.assertRaises(exceptions.TestCaseFormatError):
            compat.convert_variables(None, "examples/data/a-b.c/1.yml")

    def test_convert_request(self):
        request_with_json_body = {
            "method": "POST",
            "url": "https://postman-echo.com/post",
            "headers": {"Content-Type": "application/json"},
            "body": {"k1": "v1", "k2": "v2"},
        }
        self.assertEqual(
            compat._convert_request(request_with_json_body),
            {
                "method": "POST",
                "url": "https://postman-echo.com/post",
                "headers": {"Content-Type": "application/json"},
                "json": {"k1": "v1", "k2": "v2"},
            },
        )

        request_with_text_body = {
            "method": "POST",
            "url": "https://postman-echo.com/post",
            "headers": {"Content-Type": "text/plain"},
            "body": "have a nice day",
        }
        self.assertEqual(
            compat._convert_request(request_with_text_body),
            {
                "method": "POST",
                "url": "https://postman-echo.com/post",
                "headers": {"Content-Type": "text/plain"},
                "data": "have a nice day",
            },
        )

    def test_convert_jmespath(self):
        self.assertEqual(compat._convert_jmespath("content.abc"), "body.abc")
        self.assertEqual(compat._convert_jmespath("json.abc"), "body.abc")
        self.assertEqual(
            compat._convert_jmespath("headers.Content-Type"), 'headers."Content-Type"'
        )
        self.assertEqual(
            compat._convert_jmespath("headers.User-Agent"), 'headers."User-Agent"'
        )
        self.assertEqual(
            compat._convert_jmespath('headers."Content-Type"'), 'headers."Content-Type"'
        )
        self.assertEqual(
            compat._convert_jmespath("body.users[-1]"),
            "body.users[-1]",
        )
        self.assertEqual(
            compat._convert_jmespath("body.result.WorkNode_-1"),
            "body.result.WorkNode_-1",
        )

    def test_convert_extractors(self):
        self.assertEqual(
            compat._convert_extractors(
                [{"varA": "content.varA"}, {"varB": "json.varB"}]
            ),
            {"varA": "body.varA", "varB": "body.varB"},
        )
        self.assertEqual(
            compat._convert_extractors([{"varA": "content[0].varA"}]),
            {"varA": "body[0].varA"},
        )
        self.assertEqual(
            compat._convert_extractors({"varA": "content[0].varA"}),
            {"varA": "body[0].varA"},
        )

    def test_convert_validators(self):
        self.assertEqual(
            compat._convert_validators(
                [{"check": "content.abc", "assert": "eq", "expect": 201}]
            ),
            [{"check": "body.abc", "assert": "eq", "expect": 201}],
        )
        self.assertEqual(
            compat._convert_validators([{"eq": ["content.abc", 201]}]),
            [{"eq": ["body.abc", 201]}],
        )
        self.assertEqual(
            compat._convert_validators([{"eq": ["content[0].name", 201]}]),
            [{"eq": ["body[0].name", 201]}],
        )

    def test_ensure_testcase_v4_api(self):
        api_content = {
            "name": "get with params",
            "request": {
                "method": "GET",
                "url": "/get",
                "params": {"foo1": "bar1", "foo2": "bar2"},
                "headers": {"User-Agent": "HttpRunner/3.0"},
            },
            "extract": [{"varA": "content.varA"}, {"user_agent": "headers.User-Agent"}],
            "validate": [{"eq": ["content.varB", 200]}, {"lt": ["json[0].varC", 0]}],
        }
        self.assertEqual(
            compat.ensure_testcase_v4_api(api_content),
            {
                "config": {
                    "name": "get with params",
                    "export": ["varA", "user_agent"],
                },
                "teststeps": [
                    {
                        "name": "get with params",
                        "request": {
                            "method": "GET",
                            "url": "/get",
                            "params": {"foo1": "bar1", "foo2": "bar2"},
                            "headers": {"User-Agent": "HttpRunner/3.0"},
                        },
                        "extract": {
                            "varA": "body.varA",
                            "user_agent": 'headers."User-Agent"',
                        },
                        "validate": [
                            {"eq": ["body.varB", 200]},
                            {"lt": ["body[0].varC", 0]},
                        ],
                    }
                ],
            },
        )

    def test_ensure_testcase_v4(self):
        testcase_content = {
            "config": {"name": "xxx", "base_url": HTTP_BIN_URL},
            "teststeps": [
                {
                    "name": "get with params",
                    "request": {
                        "method": "GET",
                        "url": "/get",
                        "params": {"foo1": "bar1", "foo2": "bar2"},
                        "headers": {"User-Agent": "HttpRunner/3.0"},
                    },
                    "extract": [
                        {"varA": "content.varA"},
                        {"user_agent": "headers.User-Agent"},
                    ],
                    "validate": [
                        {"eq": ["content.varB", 200]},
                        {"lt": ["json[0].varC", 0]},
                    ],
                }
            ],
        }
        self.assertEqual(
            compat.ensure_testcase_v4(testcase_content),
            {
                "config": {"name": "xxx", "base_url": HTTP_BIN_URL},
                "teststeps": [
                    {
                        "name": "get with params",
                        "request": {
                            "method": "GET",
                            "url": "/get",
                            "params": {"foo1": "bar1", "foo2": "bar2"},
                            "headers": {"User-Agent": "HttpRunner/3.0"},
                        },
                        "extract": {
                            "varA": "body.varA",
                            "user_agent": 'headers."User-Agent"',
                        },
                        "validate": [
                            {"eq": ["body.varB", 200]},
                            {"lt": ["body[0].varC", 0]},
                        ],
                    }
                ],
            },
        )

    def test_ensure_cli_args(self):
        args1 = ["examples/postman_echo/request_methods/hardcode.yml", "--failfast"]
        self.assertEqual(
            compat.ensure_cli_args(args1),
            ["examples/postman_echo/request_methods/hardcode.yml"],
        )

        args2 = ["examples/postman_echo/request_methods/hardcode.yml", "--save-tests"]
        self.assertEqual(
            compat.ensure_cli_args(args2),
            ["examples/postman_echo/request_methods/hardcode.yml"],
        )
        self.assertTrue(os.path.isfile("examples/postman_echo/conftest.py"))

        args3 = [
            "examples/postman_echo/request_methods/hardcode.yml",
            "--report-file",
            "report.html",
        ]
        self.assertEqual(
            compat.ensure_cli_args(args3),
            [
                "examples/postman_echo/request_methods/hardcode.yml",
                "--html",
                "report.html",
                "--self-contained-html",
            ],
        )

        args4 = [
            "examples/postman_echo/request_methods/hardcode.yml",
            "--failfast",
            "--save-tests",
            "--report-file",
            "report.html",
        ]
        self.assertEqual(
            compat.ensure_cli_args(args4),
            [
                "examples/postman_echo/request_methods/hardcode.yml",
                "--html",
                "report.html",
                "--self-contained-html",
            ],
        )

    def test_ensure_file_path(self):
        self.assertEqual(
            compat.ensure_path_sep("demo\\test.yml"), os.sep.join(["demo", "test.yml"])
        )
        self.assertEqual(
            compat.ensure_path_sep(os.path.join(os.getcwd(), "demo\\test.yml")),
            os.path.join(os.getcwd(), os.sep.join(["demo", "test.yml"])),
        )
        self.assertEqual(
            compat.ensure_path_sep("demo/test.yml"), os.sep.join(["demo", "test.yml"])
        )
        self.assertEqual(
            compat.ensure_path_sep(os.path.join(os.getcwd(), "demo/test.yml")),
            os.path.join(os.getcwd(), os.sep.join(["demo", "test.yml"])),
        )
