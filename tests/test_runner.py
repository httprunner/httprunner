import os
import requests
from ate import runner, exception, utils
from tests.base import ApiServerUnittest

class TestRunner(ApiServerUnittest):

    def setUp(self):
        self.test_runner = runner.Runner()
        self.reset_all()

        self.testcase_file_path_list = [
            os.path.join(
                os.getcwd(), 'tests/data/demo_testset_hardcode.yml'),
            os.path.join(
                os.getcwd(), 'tests/data/demo_testset_hardcode.json')
        ]

    def reset_all(self):
        url = "%s/api/reset-all" % self.host
        headers = self.get_authenticated_headers()
        return self.api_client.get(url, headers=headers)

    def test_run_single_testcase(self):
        for testcase_file_path in self.testcase_file_path_list:
            testcases = utils.load_testcases(testcase_file_path)
            testcase = testcases[0]["test"]
            success, _ = self.test_runner.run_test(testcase)
            self.assertTrue(success)

            testcase = testcases[1]["test"]
            success, _ = self.test_runner.run_test(testcase)
            self.assertTrue(success)

            testcase = testcases[2]["test"]
            success, _ = self.test_runner.run_test(testcase)
            self.assertTrue(success)

    def test_run_single_testcase_fail(self):
        testcase = {
            "name": "get token",
            "request": {
                "url": "http://127.0.0.1:5000/api/get-token",
                "method": "POST",
                "headers": {
                    "content-type": "application/json",
                    "user_agent": "iOS/10.3",
                    "device_sn": "HZfFBh6tU59EdXJ",
                    "os_platform": "ios",
                    "app_version": "2.8.6"
                },
                "json": {
                    "sign": "f1219719911caae89ccc301679857ebfda115ca2"
                }
            },
            "extract_binds": [
                {"token": "content.token"}
            ],
            "validators": [
                {"check": "status_code", "comparator": "eq", "expected": 205},
                {"check": "content.token", "comparator": "len_eq", "expected": 19}
            ]
        }

        success, diff_content_list = self.test_runner.run_test(testcase)
        self.assertFalse(success)
        self.assertEqual(
            diff_content_list[0],
            {"check": "status_code", "comparator": "eq", "expected": 205, 'value': 200}
        )

    def test_run_testset_hardcode(self):
        for testcase_file_path in self.testcase_file_path_list:
            testsets = utils.load_testcases_by_path(testcase_file_path)
            results = self.test_runner.run_testset(testsets[0])
            self.assertEqual(len(results), 3)
            self.assertEqual(results, [(True, [])] * 3)

    def test_run_testsets_hardcode(self):
        for testcase_file_path in self.testcase_file_path_list:
            testsets = utils.load_testcases_by_path(testcase_file_path)
            results = self.test_runner.run_testsets(testsets)
            self.assertEqual(len(results), 1)
            self.assertEqual(results, [[(True, [])] * 3])

    def test_run_testset_template_variables(self):
        testcase_file_path = os.path.join(
            os.getcwd(), 'tests/data/demo_testset_variables.yml')
        testsets = utils.load_testcases_by_path(testcase_file_path)
        results = self.test_runner.run_testset(testsets[0])
        self.assertEqual(len(results), 3)
        self.assertEqual(results, [(True, [])] * 3)

    def test_run_testset_template_import_functions(self):
        testcase_file_path = os.path.join(
            os.getcwd(), 'tests/data/demo_testset_template_import_functions.yml')
        testsets = utils.load_testcases_by_path(testcase_file_path)
        results = self.test_runner.run_testset(testsets[0])
        self.assertEqual(len(results), 3)
        self.assertEqual(results, [(True, [])] * 3)

    def test_run_testsets_template_import_functions(self):
        testcase_file_path = os.path.join(
            os.getcwd(), 'tests/data/demo_testset_template_import_functions.yml')
        testsets = utils.load_testcases_by_path(testcase_file_path)
        results = self.test_runner.run_testsets(testsets)
        self.assertEqual(len(results), 1)
        self.assertEqual(results, [[(True, [])] * 3])

    def test_run_testsets_template_lambda_functions(self):
        testcase_file_path = os.path.join(
            os.getcwd(), 'tests/data/demo_testset_template_lambda_functions.yml')
        testsets = utils.load_testcases_by_path(testcase_file_path)
        results = self.test_runner.run_testsets(testsets)
        self.assertEqual(len(results), 1)
        self.assertEqual(results, [[(True, [])] * 3])
