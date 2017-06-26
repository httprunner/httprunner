import unittest

from ate.testcase import TestcaseParser
from ate import exception


class TestcaseParserUnittest(unittest.TestCase):

    def setUp(self):
        self.variables_binds = {
            "uid": "1000",
            "random": "A2dEx",
            "authorization": "a83de0ff8d2e896dbd8efb81ba14e17d",
            "json": {
                "name": "user1",
                "password": "123456"
            },
            "expected_status": 201,
            "expected_success": True
        }
        self.testcase_parser = TestcaseParser(self.variables_binds)

    def test_parse_testcase_template(self):
        testcase = {
            "request": {
                "url": "http://127.0.0.1:5000/api/users/${uid}",
                "method": "POST",
                "headers": {
                    "Content-Type": "application/json",
                    "authorization": "${authorization}",
                    "random": "${random}"
                },
                "body": "${json}"
            },
            "response": {
                "status_code": "${expected_status}",
                "headers": {
                    "Content-Type": "application/json"
                },
                "body": {
                    "success": "${expected_success}",
                    "msg": "user created successfully."
                }
            }
        }
        parsed_testcase = self.testcase_parser.parse(testcase)

        self.assertEqual(
            parsed_testcase["request"]["url"],
            "http://127.0.0.1:5000/api/users/%s" % self.variables_binds["uid"]
        )
        self.assertEqual(
            parsed_testcase["request"]["headers"]["authorization"],
            self.variables_binds["authorization"]
        )
        self.assertEqual(
            parsed_testcase["request"]["headers"]["random"],
            self.variables_binds["random"]
        )
        self.assertEqual(
            parsed_testcase["request"]["body"],
            self.variables_binds["json"]
        )
        self.assertEqual(
            parsed_testcase["response"]["status_code"],
            self.variables_binds["expected_status"]
        )
        self.assertEqual(
            parsed_testcase["response"]["body"]["success"],
            self.variables_binds["expected_success"]
        )

    def test_parse_testcase_template_miss_bind_variable(self):
        testcase = {
            "request": {
                "url": "http://127.0.0.1:5000/api/users/${uid}",
                "method": "${method}"
            }
        }
        with self.assertRaises(exception.ParamsError):
            self.testcase_parser.parse(testcase)

    def test_parse_testcase_with_new_variable_binds(self):
        testcase = {
            "request": {
                "url": "http://127.0.0.1:5000/api/users/${uid}",
                "method": "${method}"
            }
        }
        new_variable_binds = {
            "method": "GET"
        }
        parsed_testcase = self.testcase_parser.parse(testcase, new_variable_binds)

        self.assertIn("method", self.testcase_parser.variables_binds)
        self.assertEqual(
            parsed_testcase["request"]["method"],
            new_variable_binds["method"]
        )
