import os
import unittest

from httprunner import compat, exceptions, loader


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
            compat._convert_jmespath("body.data.buildings.0.building_id"),
            "body.data.buildings[0].building_id",
        )
        self.assertEqual(
            compat._convert_jmespath("body.users[-1]"),
            "body.users[-1]",
        )
        self.assertEqual(
            compat._convert_jmespath("body.result.WorkNode_-1"),
            "body.result.WorkNode_-1",
        )
        with self.assertRaises(SystemExit):
            compat._convert_jmespath("2.buildings.0.building_id")

    def test_convert_extractors(self):
        self.assertEqual(
            compat._convert_extractors(
                [{"varA": "content.varA"}, {"varB": "json.varB"}]
            ),
            {"varA": "body.varA", "varB": "body.varB"},
        )
        self.assertEqual(
            compat._convert_extractors([{"varA": "content.0.varA"}]),
            {"varA": "body[0].varA"},
        )
        self.assertEqual(
            compat._convert_extractors({"varA": "content.0.varA"}),
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
            compat._convert_validators([{"eq": ["content.0.name", 201]}]),
            [{"eq": ["body[0].name", 201]}],
        )

    def test_ensure_testcase_v3_api(self):
        api_content = {
            "name": "get with params",
            "request": {
                "method": "GET",
                "url": "/get",
                "params": {"foo1": "bar1", "foo2": "bar2"},
                "headers": {"User-Agent": "HttpRunner/3.0"},
            },
            "extract": [{"varA": "content.varA"}, {"user_agent": "headers.User-Agent"}],
            "validate": [{"eq": ["content.varB", 200]}, {"lt": ["json.0.varC", 0]}],
        }
        self.assertEqual(
            compat.ensure_testcase_v3_api(api_content),
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

    def test_ensure_testcase_v3(self):
        testcase_content = {
            "config": {"name": "xxx", "base_url": "https://httpbin.org"},
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
                        {"lt": ["json.0.varC", 0]},
                    ],
                }
            ],
        }
        self.assertEqual(
            compat.ensure_testcase_v3(testcase_content),
            {
                "config": {"name": "xxx", "base_url": "https://httpbin.org"},
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
