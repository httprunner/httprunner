import os
import unittest

from httprunner import compat


class TestCompat(unittest.TestCase):
    def test_convert_jmespath(self):

        self.assertEqual(compat.convert_jmespath("content.abc"), "body.abc")
        self.assertEqual(compat.convert_jmespath("json.abc"), "body.abc")
        self.assertEqual(
            compat.convert_jmespath("headers.Content-Type"), 'headers."Content-Type"'
        )
        self.assertEqual(
            compat.convert_jmespath('headers."Content-Type"'), 'headers."Content-Type"'
        )
        self.assertEqual(
            compat.convert_jmespath("body.data.buildings.0.building_id"),
            "body.data.buildings[0].building_id",
        )
        with self.assertRaises(SystemExit):
            compat.convert_jmespath("2.buildings.0.building_id")

    def test_convert_extractors(self):
        self.assertEqual(
            compat.convert_extractors(
                [{"varA": "content.varA"}, {"varB": "json.varB"}]
            ),
            {"varA": "body.varA", "varB": "body.varB"},
        )
        self.assertEqual(
            compat.convert_extractors([{"varA": "content.0.varA"}]),
            {"varA": "body[0].varA"},
        )
        self.assertEqual(
            compat.convert_extractors({"varA": "content.0.varA"}),
            {"varA": "body[0].varA"},
        )

    def test_convert_validators(self):
        self.assertEqual(
            compat.convert_validators(
                [{"check": "content.abc", "assert": "eq", "expect": 201}]
            ),
            [{"check": "body.abc", "assert": "eq", "expect": 201}],
        )
        self.assertEqual(
            compat.convert_validators([{"eq": ["content.abc", 201]}]),
            [{"eq": ["body.abc", 201]}],
        )
        self.assertEqual(
            compat.convert_validators([{"eq": ["content.0.name", 201]}]),
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
                "config": {"name": "get with params"},
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
        self.assertTrue(
            os.path.isfile("examples/postman_echo/request_methods/conftest.py")
        )

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
