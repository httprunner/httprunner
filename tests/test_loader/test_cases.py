
import os
import unittest

from httprunner import exceptions, loader
from httprunner.loader import buildup


class TestModuleLoader(unittest.TestCase):

    def test_filter_module_functions(self):
        module_functions = buildup.load_module_functions(buildup)
        self.assertIn("load_module_functions", module_functions)
        self.assertNotIn("is_py3", module_functions)

    def test_load_debugtalk_module(self):
        project_mapping = buildup.load_project_data(os.path.join(os.getcwd(), "httprunner"))
        self.assertNotIn("alter_response", project_mapping["functions"])

        project_mapping = buildup.load_project_data(os.path.join(os.getcwd(), "tests"))
        self.assertIn("alter_response", project_mapping["functions"])

        is_status_code_200 = project_mapping["functions"]["is_status_code_200"]
        self.assertTrue(is_status_code_200(200))
        self.assertFalse(is_status_code_200(500))

    def test_load_debugtalk_py(self):
        project_mapping = buildup.load_project_data("tests/data/demo_testcase.yml")
        project_working_directory = project_mapping["PWD"]
        debugtalk_functions = project_mapping["functions"]
        self.assertEqual(
            project_working_directory,
            os.path.join(os.getcwd(), "tests")
        )
        self.assertIn("gen_md5", debugtalk_functions)

        project_mapping = buildup.load_project_data("tests/base.py")
        project_working_directory = project_mapping["PWD"]
        debugtalk_functions = project_mapping["functions"]
        self.assertEqual(
            project_working_directory,
            os.path.join(os.getcwd(), "tests")
        )
        self.assertIn("gen_md5", debugtalk_functions)

        project_mapping = buildup.load_project_data("httprunner/__init__.py")
        project_working_directory = project_mapping["PWD"]
        debugtalk_functions = project_mapping["functions"]
        self.assertEqual(
            project_working_directory,
            os.getcwd()
        )
        self.assertEqual(debugtalk_functions, {})


class TestSuiteLoader(unittest.TestCase):

    @classmethod
    def setUpClass(cls):
        cls.project_mapping = buildup.load_project_data(os.path.join(os.getcwd(), "tests"))
        cls.tests_def_mapping = buildup.tests_def_mapping

    def test_load_teststep_api(self):
        raw_test = {
            "name": "create user (override).",
            "api": "api/create_user.yml",
            "variables": [
                {"uid": "999"}
            ]
        }
        teststep = buildup.load_teststep(raw_test)
        self.assertEqual(
            "create user (override).",
            teststep["name"]
        )
        self.assertIn("api_def", teststep)
        api_def = teststep["api_def"]
        self.assertEqual(api_def["name"], "create user")
        self.assertEqual(api_def["request"]["url"], "/api/users/$uid")

    def test_load_teststep_testcase(self):
        raw_test = {
            "name": "setup and reset all (override).",
            "testcase": "testcases/setup.yml",
            "variables": [
                {"device_sn": "$device_sn"}
            ]
        }
        testcase = buildup.load_teststep(raw_test)
        self.assertEqual(
            "setup and reset all (override).",
            testcase["name"]
        )
        tests = testcase["testcase_def"]["teststeps"]
        self.assertEqual(len(tests), 2)
        self.assertEqual(tests[0]["name"], "get token (setup)")
        self.assertEqual(tests[1]["name"], "reset all users")

    def test_load_test_file_api(self):
        loaded_content = buildup.load_test_file("tests/api/create_user.yml")
        self.assertEqual(loaded_content["type"], "api")
        self.assertIn("path", loaded_content)
        self.assertIn("request", loaded_content)
        self.assertEqual(loaded_content["request"]["url"], "/api/users/$uid")

    def test_load_test_file_testcase(self):
        for loaded_content in [
            buildup.load_test_file("tests/testcases/setup.yml"),
            buildup.load_test_file("tests/testcases/setup.json")
        ]:
            self.assertEqual(loaded_content["type"], "testcase")
            self.assertIn("path", loaded_content)
            self.assertIn("config", loaded_content)
            self.assertEqual(loaded_content["config"]["name"], "setup and reset all.")
            self.assertIn("teststeps", loaded_content)
            self.assertEqual(len(loaded_content["teststeps"]), 2)

    def test_load_test_file_testcase_v2(self):
        for loaded_content in [
            buildup.load_test_file("tests/testcases/setup.v2.yml"),
            buildup.load_test_file("tests/testcases/setup.v2.json")
        ]:
            self.assertEqual(loaded_content["type"], "testcase")
            self.assertIn("path", loaded_content)
            self.assertIn("config", loaded_content)
            self.assertEqual(loaded_content["config"]["name"], "setup and reset all.")
            self.assertIn("teststeps", loaded_content)
            self.assertEqual(len(loaded_content["teststeps"]), 2)

    def test_load_test_file_testsuite(self):
        for loaded_content in [
            buildup.load_test_file("tests/testsuites/create_users.yml"),
            buildup.load_test_file("tests/testsuites/create_users.json")
        ]:
            self.assertEqual(loaded_content["type"], "testsuite")

            testcases = loaded_content["testcases"]
            self.assertEqual(len(testcases), 2)
            self.assertIn('create user 1000 and check result.', testcases)
            self.assertIn('testcase_def', testcases["create user 1000 and check result."])
            self.assertEqual(
                testcases["create user 1000 and check result."]["testcase_def"]["config"]["name"],
                "create user and check result."
            )

    def test_load_test_file_testsuite_v2(self):
        for loaded_content in [
            buildup.load_test_file("tests/testsuites/create_users.v2.yml"),
            buildup.load_test_file("tests/testsuites/create_users.v2.json")
        ]:
            self.assertEqual(loaded_content["type"], "testsuite")

            testcases = loaded_content["testcases"]
            self.assertEqual(len(testcases), 2)
            self.assertIn('create user 1000 and check result.', testcases)
            self.assertIn('testcase_def', testcases["create user 1000 and check result."])
            self.assertEqual(
                testcases["create user 1000 and check result."]["testcase_def"]["config"]["name"],
                "create user and check result."
            )

    def test_load_tests_api_file(self):
        path = os.path.join(
            os.getcwd(), 'tests/api/create_user.yml')
        tests_mapping = loader.load_cases(path)
        project_mapping = tests_mapping["project_mapping"]
        api_list = tests_mapping["apis"]
        self.assertEqual(len(api_list), 1)
        self.assertEqual(api_list[0]["request"]["url"], "/api/users/$uid")

    def test_load_tests_testcase_file(self):
        # absolute file path
        path = os.path.join(
            os.getcwd(), 'tests/data/demo_testcase_hardcode.json')
        tests_mapping = loader.load_cases(path)
        project_mapping = tests_mapping["project_mapping"]
        testcases_list = tests_mapping["testcases"]
        self.assertEqual(len(testcases_list), 1)
        self.assertEqual(len(testcases_list[0]["teststeps"]), 3)
        self.assertIn("get_sign", project_mapping["functions"])

        # relative file path
        path = 'tests/data/demo_testcase_hardcode.yml'
        tests_mapping = loader.load_cases(path)
        project_mapping = tests_mapping["project_mapping"]
        testcases_list = tests_mapping["testcases"]
        self.assertEqual(len(testcases_list), 1)
        self.assertEqual(len(testcases_list[0]["teststeps"]), 3)
        self.assertIn("get_sign", project_mapping["functions"])

    def test_load_tests_testcase_file_2(self):
        testcase_file_path = os.path.join(
            os.getcwd(), 'tests/data/demo_testcase.yml')
        tests_mapping = loader.load_cases(testcase_file_path)
        testcases = tests_mapping["testcases"]
        self.assertIsInstance(testcases, list)
        self.assertEqual(testcases[0]["config"]["name"], '123t$var_a')
        self.assertIn(
            "sum_two",
            tests_mapping["project_mapping"]["functions"]
        )
        self.assertEqual(
            testcases[0]["config"]["variables"]["var_c"],
            "${sum_two($var_a, $var_b)}"
        )
        self.assertEqual(
            testcases[0]["config"]["variables"]["PROJECT_KEY"],
            "${ENV(PROJECT_KEY)}"
        )

    def test_load_tests_testcase_file_with_api_ref(self):
        path = os.path.join(
            os.getcwd(), 'tests/data/demo_testcase_layer.yml')
        tests_mapping = loader.load_cases(path)
        project_mapping = tests_mapping["project_mapping"]
        testcases_list = tests_mapping["testcases"]
        self.assertIn('device_sn', testcases_list[0]["config"]["variables"])
        self.assertIn("gen_md5", project_mapping["functions"])
        self.assertIn("base_url", testcases_list[0]["config"])
        test_dict0 = testcases_list[0]["teststeps"][0]
        self.assertEqual(
            "get token with $user_agent, $app_version",
            test_dict0["name"]
        )
        self.assertIn("/api/get-token", test_dict0["api_def"]["request"]["url"])
        self.assertIn(
            {'eq': ['status_code', 200]},
            test_dict0["validate"]
        )

    def test_load_tests_testsuite_file_with_testcase_ref(self):
        path = os.path.join(
            os.getcwd(), 'tests/testsuites/create_users.yml')
        tests_mapping = loader.load_cases(path)
        project_mapping = tests_mapping["project_mapping"]
        testsuites_list = tests_mapping["testsuites"]

        self.assertEqual(
            "create users with uid",
            testsuites_list[0]["config"]["name"]
        )
        self.assertEqual(
            '${gen_random_string(15)}',
            testsuites_list[0]["config"]["variables"]['device_sn']
        )
        self.assertIn(
            "create user 1000 and check result.",
            testsuites_list[0]["testcases"]
        )

        self.assertEqual(
            testsuites_list[0]["testcases"]["create user 1000 and check result."]["testcase_def"]["config"]["name"],
            "create user and check result."
        )

    def test_load_tests_folder_path(self):
        # absolute folder path
        path = os.path.join(os.getcwd(), 'tests/data')
        tests_mapping = loader.load_cases(path)
        testcase_list_1 = tests_mapping["testcases"]
        self.assertGreater(len(testcase_list_1), 4)

        # relative folder path
        path = 'tests/data/'
        tests_mapping = loader.load_cases(path)
        testcase_list_2 = tests_mapping["testcases"]
        self.assertEqual(len(testcase_list_1), len(testcase_list_2))

    def test_load_tests_path_not_exist(self):
        # absolute folder path
        path = os.path.join(os.getcwd(), 'tests/data_not_exist')
        with self.assertRaises(exceptions.FileNotFound):
            loader.load_cases(path)

        # relative folder path
        path = 'tests/data_not_exist'
        with self.assertRaises(exceptions.FileNotFound):
            loader.load_cases(path)

    def test_load_project_tests(self):
        buildup.load_project_data(os.path.join(os.getcwd(), "tests"))
        self.assertIn("gen_md5", self.project_mapping["functions"])
        self.assertEqual(self.project_mapping["env"]["PROJECT_KEY"], "ABCDEFGH")
        self.assertEqual(self.project_mapping["PWD"], os.path.abspath(os.path.dirname(os.path.dirname(__file__))))
        self.assertEqual(self.project_mapping["test_path"], os.path.abspath(os.path.dirname(os.path.dirname(__file__))))
