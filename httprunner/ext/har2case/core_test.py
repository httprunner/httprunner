import os

from httprunner.ext.har2case.core import HarParser
from httprunner.ext.har2case.utils import load_har_log_entries
from httprunner.ext.har2case.utils_test import TestUtils


class TestHar(TestUtils):
    def setUp(self):
        self.har_path = os.path.join(os.path.dirname(__file__), "data", "demo.har")
        self.har_parser = HarParser(self.har_path)

    def test_prepare_teststep(self):
        log_entries = load_har_log_entries(self.har_path)
        teststep_dict = self.har_parser._prepare_teststep(log_entries[0])
        self.assertIn("name", teststep_dict)
        self.assertIn("request", teststep_dict)
        self.assertIn("validate", teststep_dict)

        validators_mapping = {
            validator["eq"][0]: validator["eq"][1]
            for validator in teststep_dict["validate"]
        }
        self.assertEqual(validators_mapping["status_code"], 200)
        self.assertEqual(validators_mapping["body.IsSuccess"], True)
        self.assertEqual(validators_mapping["body.Code"], 200)
        self.assertEqual(validators_mapping["body.Message"], None)

    def test_prepare_teststeps(self):
        teststeps = self.har_parser._prepare_teststeps()
        self.assertIsInstance(teststeps, list)
        self.assertIn("name", teststeps[0])
        self.assertIn("request", teststeps[0])
        self.assertIn("validate", teststeps[0])

    def test_gen_testcase_yaml(self):
        yaml_file = os.path.join(os.path.dirname(__file__), "data", "demo.yaml")

        self.har_parser.gen_testcase(file_type="YAML")
        self.assertTrue(os.path.isfile(yaml_file))
        os.remove(yaml_file)

    def test_gen_testcase_json(self):
        json_file = os.path.join(os.path.dirname(__file__), "data", "demo.json")

        self.har_parser.gen_testcase(file_type="JSON")
        self.assertTrue(os.path.isfile(json_file))
        os.remove(json_file)

    def test_filter(self):
        filter_str = "httprunner"
        har_parser = HarParser(self.har_path, filter_str)
        teststeps = har_parser._prepare_teststeps()
        self.assertEqual(
            teststeps[0]["request"]["url"],
            "https://httprunner.top/api/v1/Account/Login",
        )

        filter_str = "debugtalk"
        har_parser = HarParser(self.har_path, filter_str)
        teststeps = har_parser._prepare_teststeps()
        self.assertEqual(teststeps, [])

    def test_exclude(self):
        exclude_str = "debugtalk"
        har_parser = HarParser(self.har_path, exclude_str=exclude_str)
        teststeps = har_parser._prepare_teststeps()
        self.assertEqual(
            teststeps[0]["request"]["url"],
            "https://httprunner.top/api/v1/Account/Login",
        )

        exclude_str = "httprunner"
        har_parser = HarParser(self.har_path, exclude_str=exclude_str)
        teststeps = har_parser._prepare_teststeps()
        self.assertEqual(teststeps, [])

    def test_exclude_multiple(self):
        exclude_str = "httprunner|v2"
        har_parser = HarParser(self.har_path, exclude_str=exclude_str)
        teststeps = har_parser._prepare_teststeps()
        self.assertEqual(teststeps, [])

        exclude_str = "http2|v1"
        har_parser = HarParser(self.har_path, exclude_str=exclude_str)
        teststeps = har_parser._prepare_teststeps()
        self.assertEqual(teststeps, [])

    def test_make_request_data_params(self):
        testcase_dict = {"name": "", "request": {}, "validate": []}
        entry_json = {
            "request": {
                "method": "POST",
                "postData": {
                    "mimeType": "application/x-www-form-urlencoded; charset=utf-8",
                    "params": [{"name": "a", "value": 1}, {"name": "b", "value": "2"}],
                },
            }
        }
        self.har_parser._make_request_data(testcase_dict, entry_json)
        self.assertEqual(testcase_dict["request"]["data"]["a"], 1)
        self.assertEqual(testcase_dict["request"]["data"]["b"], "2")

    def test_make_request_data_json(self):
        testcase_dict = {"name": "", "request": {}, "validate": []}
        entry_json = {
            "request": {
                "method": "POST",
                "postData": {
                    "mimeType": "application/json; charset=utf-8",
                    "text": '{"a":"1","b":"2"}',
                },
            }
        }
        self.har_parser._make_request_data(testcase_dict, entry_json)
        self.assertEqual(testcase_dict["request"]["json"], {"a": "1", "b": "2"})

    def test_make_request_data_text_empty(self):
        testcase_dict = {"name": "", "request": {}, "validate": []}
        entry_json = {
            "request": {
                "method": "POST",
                "postData": {"mimeType": "application/json; charset=utf-8", "text": ""},
            }
        }
        self.har_parser._make_request_data(testcase_dict, entry_json)
        self.assertEqual(testcase_dict["request"]["data"], "")

    def test_make_validate(self):
        testcase_dict = {"name": "", "request": {}, "validate": []}
        entry_json = {
            "request": {},
            "response": {
                "status": 200,
                "headers": [
                    {
                        "name": "Content-Type",
                        "value": "application/json; charset=utf-8",
                    },
                ],
                "content": {
                    "size": 71,
                    "mimeType": "application/json; charset=utf-8",
                    # raw response content text is application/jose type
                    "text": "ZXlKaGJHY2lPaUpTVTBFeFh6VWlMQ0psYm1NaU9pSkJNVEk0UTBKRExV",
                    "encoding": "base64",
                },
            },
        }
        self.har_parser._make_validate(testcase_dict, entry_json)
        self.assertEqual(testcase_dict["validate"][0], {"eq": ["status_code", 200]})
        self.assertEqual(
            testcase_dict["validate"][1],
            {"eq": ["headers.Content-Type", "application/json; charset=utf-8"]},
        )

    def test_make_testcase(self):
        har_path = os.path.join(
            os.path.dirname(__file__), "data", "demo-quickstart.har"
        )
        har_parser = HarParser(har_path)
        testcase = har_parser._make_testcase()
        self.assertIsInstance(testcase, dict)
        self.assertIn("config", testcase)
        self.assertIn("teststeps", testcase)
        self.assertEqual(len(testcase["teststeps"]), 2)
