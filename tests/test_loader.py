
import os
import unittest

from httprunner import exceptions, loader, validator


class TestFileLoader(unittest.TestCase):

    def test_load_yaml_file_file_format_error(self):
        yaml_tmp_file = "tests/data/tmp.yml"
        # create empty yaml file
        with open(yaml_tmp_file, 'w') as f:
            f.write("")

        with self.assertRaises(exceptions.FileFormatError):
            loader.load_yaml_file(yaml_tmp_file)

        os.remove(yaml_tmp_file)

        # create invalid format yaml file
        with open(yaml_tmp_file, 'w') as f:
            f.write("abc")

        with self.assertRaises(exceptions.FileFormatError):
            loader.load_yaml_file(yaml_tmp_file)

        os.remove(yaml_tmp_file)

    def test_load_json_file_file_format_error(self):
        json_tmp_file = "tests/data/tmp.json"
        # create empty file
        with open(json_tmp_file, 'w') as f:
            f.write("")

        with self.assertRaises(exceptions.FileFormatError):
            loader.load_json_file(json_tmp_file)

        os.remove(json_tmp_file)

        # create empty json file
        with open(json_tmp_file, 'w') as f:
            f.write("{}")

        with self.assertRaises(exceptions.FileFormatError):
            loader.load_json_file(json_tmp_file)

        os.remove(json_tmp_file)

        # create invalid format json file
        with open(json_tmp_file, 'w') as f:
            f.write("abc")

        with self.assertRaises(exceptions.FileFormatError):
            loader.load_json_file(json_tmp_file)

        os.remove(json_tmp_file)

    def test_load_testcases_bad_filepath(self):
        testcase_file_path = os.path.join(os.getcwd(), 'tests/data/demo')
        with self.assertRaises(exceptions.FileNotFound):
            loader.load_file(testcase_file_path)

    def test_load_json_testcases(self):
        testcase_file_path = os.path.join(
            os.getcwd(), 'tests/data/demo_testcase_hardcode.json')
        testcases = loader.load_file(testcase_file_path)
        self.assertEqual(len(testcases), 3)
        test = testcases[0]["test"]
        self.assertIn('name', test)
        self.assertIn('request', test)
        self.assertIn('url', test['request'])
        self.assertIn('method', test['request'])

    def test_load_yaml_testcases(self):
        testcase_file_path = os.path.join(
            os.getcwd(), 'tests/data/demo_testcase_hardcode.yml')
        testcases = loader.load_file(testcase_file_path)
        self.assertEqual(len(testcases), 3)
        test = testcases[0]["test"]
        self.assertIn('name', test)
        self.assertIn('request', test)
        self.assertIn('url', test['request'])
        self.assertIn('method', test['request'])

    def test_load_csv_file_one_parameter(self):
        csv_file_path = os.path.join(
            os.getcwd(), 'tests/data/user_agent.csv')
        csv_content = loader.load_file(csv_file_path)
        self.assertEqual(
            csv_content,
            [
                {'user_agent': 'iOS/10.1'},
                {'user_agent': 'iOS/10.2'},
                {'user_agent': 'iOS/10.3'}
            ]
        )

    def test_load_csv_file_multiple_parameters(self):
        csv_file_path = os.path.join(
            os.getcwd(), 'tests/data/account.csv')
        csv_content = loader.load_file(csv_file_path)
        self.assertEqual(
            csv_content,
            [
                {'username': 'test1', 'password': '111111'},
                {'username': 'test2', 'password': '222222'},
                {'username': 'test3', 'password': '333333'}
            ]
        )

    def test_load_folder_files(self):
        folder = os.path.join(os.getcwd(), 'tests')
        file1 = os.path.join(os.getcwd(), 'tests', 'test_utils.py')
        file2 = os.path.join(os.getcwd(), 'tests', 'data', 'demo_binds.yml')

        files = loader.load_folder_files(folder, recursive=False)
        self.assertNotIn(file2, files)

        files = loader.load_folder_files(folder)
        self.assertIn(file2, files)
        self.assertNotIn(file1, files)

        files = loader.load_folder_files(folder)
        api_file = os.path.join(os.getcwd(), 'tests', 'api', 'basic.yml')
        self.assertIn(api_file, files)

        files = loader.load_folder_files("not_existed_foulder", recursive=False)
        self.assertEqual([], files)

        files = loader.load_folder_files(file2, recursive=False)
        self.assertEqual([], files)

    def test_load_dot_env_file(self):
        dot_env_path = os.path.join(
            os.getcwd(), "tests", ".env"
        )
        env_variables_mapping = loader.load_dot_env_file(dot_env_path)
        self.assertIn("PROJECT_KEY", env_variables_mapping)
        self.assertEqual(env_variables_mapping["UserName"], "debugtalk")

    def test_load_custom_dot_env_file(self):
        dot_env_path = os.path.join(
            os.getcwd(), "tests", "data", "test.env"
        )
        env_variables_mapping = loader.load_dot_env_file(dot_env_path)
        self.assertIn("PROJECT_KEY", env_variables_mapping)
        self.assertEqual(env_variables_mapping["UserName"], "test")
        self.assertEqual(env_variables_mapping["content_type"], "application/json; charset=UTF-8")

    def test_load_env_path_not_exist(self):
        dot_env_path = os.path.join(
            os.getcwd(), "tests", "data",
        )
        env_variables_mapping = loader.load_dot_env_file(dot_env_path)
        self.assertEqual(env_variables_mapping, {})

    def test_locate_file(self):
        with self.assertRaises(exceptions.FileNotFound):
            loader.locate_file(os.getcwd(), "debugtalk.py")

        with self.assertRaises(exceptions.FileNotFound):
            loader.locate_file("", "debugtalk.py")

        start_path = os.path.join(os.getcwd(), "tests")
        self.assertEqual(
            loader.locate_file(start_path, "debugtalk.py"),
            os.path.join(
                os.getcwd(), "tests/debugtalk.py"
            )
        )
        self.assertEqual(
            loader.locate_file("tests/", "debugtalk.py"),
            os.path.join(os.getcwd(), "tests", "debugtalk.py")
        )
        self.assertEqual(
            loader.locate_file("tests", "debugtalk.py"),
            os.path.join(os.getcwd(), "tests", "debugtalk.py")
        )
        self.assertEqual(
            loader.locate_file("tests/base.py", "debugtalk.py"),
            os.path.join(os.getcwd(), "tests", "debugtalk.py")
        )
        self.assertEqual(
            loader.locate_file("tests/data/demo_testcase.yml", "debugtalk.py"),
            os.path.join(os.getcwd(), "tests", "debugtalk.py")
        )


class TestModuleLoader(unittest.TestCase):

    def test_filter_module_functions(self):
        module_functions = loader.load_module_functions(loader)
        self.assertIn("load_module_functions", module_functions)
        self.assertNotIn("is_py3", module_functions)

    def test_load_debugtalk_module(self):
        loader.load_project_tests(os.path.join(os.getcwd(), "httprunner"))
        project_mapping = loader.project_mapping
        self.assertNotIn("alter_response", project_mapping["functions"])

        loader.load_project_tests(os.path.join(os.getcwd(), "tests"))
        project_mapping = loader.project_mapping
        self.assertIn("alter_response", project_mapping["functions"])

        is_status_code_200 = project_mapping["functions"]["is_status_code_200"]
        self.assertTrue(is_status_code_200(200))
        self.assertFalse(is_status_code_200(500))

    def test_load_debugtalk_py(self):
        loader.load_project_tests("tests/data/demo_testcase.yml")
        project_working_directory = loader.project_mapping["PWD"]
        debugtalk_functions = loader.project_mapping["functions"]
        self.assertEqual(
            project_working_directory,
            os.path.join(os.getcwd(), "tests")
        )
        self.assertIn("gen_md5", debugtalk_functions)

        loader.load_project_tests("tests/base.py")
        project_working_directory = loader.project_mapping["PWD"]
        debugtalk_functions = loader.project_mapping["functions"]
        self.assertEqual(
            project_working_directory,
            os.path.join(os.getcwd(), "tests")
        )
        self.assertIn("gen_md5", debugtalk_functions)

        loader.load_project_tests("httprunner/__init__.py")
        project_working_directory = loader.project_mapping["PWD"]
        debugtalk_functions = loader.project_mapping["functions"]
        self.assertEqual(
            project_working_directory,
            os.getcwd()
        )
        self.assertEqual(debugtalk_functions, {})

    def test_load_tests(self):
        testcase_file_path = os.path.join(
            os.getcwd(), 'tests/data/demo_testcase.yml')
        tests_mapping = loader.load_tests(testcase_file_path)
        testcases = tests_mapping["testcases"]
        self.assertIsInstance(testcases, list)
        self.assertEqual(testcases[0]["config"]["name"], '123$var_a')
        self.assertIn(
            "sum_two",
            tests_mapping["project_mapping"]["functions"]
        )
        self.assertEqual(
            testcases[0]["config"]["variables"]["var_c"],
            "${sum_two(1, 2)}"
        )
        self.assertEqual(
            testcases[0]["config"]["variables"]["PROJECT_KEY"],
            "${ENV(PROJECT_KEY)}"
        )


class TestSuiteLoader(unittest.TestCase):

    @classmethod
    def setUpClass(cls):
        loader.load_project_tests(os.path.join(os.getcwd(), "tests"))
        cls.project_mapping = loader.project_mapping
        cls.tests_def_mapping = loader.tests_def_mapping

    def test_load_test_testcase(self):
        raw_test = {
            "name": "setup and reset all (override).",
            "testcase": "testcases/setup.yml",
            "variables": [
                {"device_sn": "$device_sn"}
            ],
            "output": ["token", "device_sn"]
        }
        testcase = loader.load_test(raw_test)
        self.assertEqual(
            "setup and reset all (override).",
            testcase["name"]
        )
        tests = testcase["testcase_def"]["tests"]
        self.assertEqual(len(tests), 2)
        self.assertEqual(tests[0]["name"], "get token (setup)")
        self.assertEqual(tests[1]["name"], "reset all users")

    def test_load_testcase(self):
        raw_testcase = loader.load_file("tests/testsuites/create_users.yml")
        testcase = loader.load_testcase(raw_testcase)
        self.assertEqual(testcase["config"]["name"], "create users with uid")
        self.assertIn("device_sn", testcase["config"]["variables"])
        self.assertEqual(len(testcase["tests"]), 2)
        self.assertEqual(testcase["tests"][0]["name"], "create user 1000 and check result.")
        self.assertEqual(testcase["tests"][0]["testcase_def"]["config"]["name"], "create user and check result.")

    def test_load_testcases_by_path_files(self):
        # absolute file path
        path = os.path.join(
            os.getcwd(), 'tests/data/demo_testcase_hardcode.json')
        tests_mapping = loader.load_tests(path)
        project_mapping = tests_mapping["project_mapping"]
        testcases_list = tests_mapping["testcases"]
        self.assertEqual(len(testcases_list), 1)
        self.assertEqual(len(testcases_list[0]["tests"]), 3)
        self.assertIn("get_sign", project_mapping["functions"])

        # relative file path
        path = 'tests/data/demo_testcase_hardcode.yml'
        tests_mapping = loader.load_tests(path)
        project_mapping = tests_mapping["project_mapping"]
        testcases_list = tests_mapping["testcases"]
        self.assertEqual(len(testcases_list), 1)
        self.assertEqual(len(testcases_list[0]["tests"]), 3)
        self.assertIn("get_sign", project_mapping["functions"])

        # TODO: list/set container with file(s)
        # path = [
        #     os.path.join(os.getcwd(), 'tests/data/demo_testcase_hardcode.json'),
        #     'tests/data/demo_testcase_hardcode.yml'
        # ]
        # testcases_list = loader.load_tests(path)
        # self.assertEqual(len(testcases_list), 2)
        # self.assertEqual(len(testcases_list[0]["tests"]), 3)
        # self.assertEqual(len(testcases_list[1]["tests"]), 3)
        # testcases_list.extend(testcases_list)
        # self.assertEqual(len(testcases_list), 4)

        # for testcase in testcases_list:
        #     for test_dict in testcase["tests"]:
        #         self.assertIn('name', test_dict)
        #         self.assertIn('request', test_dict)
        #         self.assertIn('url', test_dict['request'])
        #         self.assertIn('method', test_dict['request'])

    def test_load_testcases_by_path_folder(self):
        # absolute folder path
        path = os.path.join(os.getcwd(), 'tests/data')
        tests_mapping = loader.load_tests(path)
        testcase_list_1 = tests_mapping["testcases"]
        self.assertGreater(len(testcase_list_1), 4)

        # relative folder path
        path = 'tests/data/'
        tests_mapping = loader.load_tests(path)
        testcase_list_2 = tests_mapping["testcases"]
        self.assertEqual(len(testcase_list_1), len(testcase_list_2))

        # TODO: list/set container with file(s)
        # path = [
        #     os.path.join(os.getcwd(), 'tests/data'),
        #     'tests/data/'
        # ]
        # tests_mapping = loader.load_tests(path)
        # testcase_list_3 = tests_mapping["testcases"]
        # self.assertEqual(len(testcase_list_3), 2 * len(testcase_list_1))

    def test_load_testcases_by_path_not_exist(self):
        # absolute folder path
        path = os.path.join(os.getcwd(), 'tests/data_not_exist')
        with self.assertRaises(exceptions.FileNotFound):
            loader.load_tests(path)

        # relative folder path
        path = 'tests/data_not_exist'
        with self.assertRaises(exceptions.FileNotFound):
            loader.load_tests(path)

        # TODO: list/set container with file(s)
        # path = [
        #     os.path.join(os.getcwd(), 'tests/data_not_exist'),
        #     'tests/data_not_exist/'
        # ]
        # with self.assertRaises(exceptions.FileNotFound):
        #     loader.load_tests(path)

    def test_load_testcases_with_api_ref(self):
        path = os.path.join(
            os.getcwd(), 'tests/data/demo_testcase_layer.yml')
        tests_mapping = loader.load_tests(path)
        project_mapping = tests_mapping["project_mapping"]
        testcases_list = tests_mapping["testcases"]
        self.assertIn('device_sn', testcases_list[0]["config"]["variables"])
        self.assertIn("gen_md5", project_mapping["functions"])
        self.assertIn("base_url", testcases_list[0]["config"])
        test_dict0 = testcases_list[0]["tests"][0]
        self.assertEqual(
            "get token with $user_agent, $app_version",
            test_dict0["name"]
        )
        self.assertIn("/api/get-token", test_dict0["api_def"]["request"]["url"])
        self.assertIn(
            {'eq': ['status_code', 200]},
            test_dict0["validate"]
        )

    def test_load_testcases_with_testcase_ref(self):
        path = os.path.join(
            os.getcwd(), 'tests/testsuites/create_users.yml')
        tests_mapping = loader.load_tests(path)
        project_mapping = tests_mapping["project_mapping"]
        testcases_list = tests_mapping["testcases"]

        self.assertEqual(
            "create users with uid",
            testcases_list[0]["config"]["name"]
        )
        self.assertEqual(
            {'device_sn': '${gen_random_string(15)}'},
            testcases_list[0]["config"]["variables"]
        )
        testcase0 = testcases_list[0]["tests"][0]
        self.assertEqual(
            "create user 1000 and check result.",
            testcase0["name"]
        )
        self.assertEqual(
            "create user and check result.",
            testcase0["testcase_def"]["config"]["name"]
        )

        testcase1 = testcases_list[0]["tests"][1]
        self.assertEqual(
            "create user 1001 and check result.",
            testcase1["name"]
        )
        self.assertEqual(
            {'uid': 1001},
            testcase1["variables"]
        )

    def test_load_folder_content(self):
        path = os.path.join(os.getcwd(), "tests", "api")
        items_mapping = loader.load_folder_content(path)
        file_path = os.path.join(os.getcwd(), "tests", "api", "basic.yml")
        self.assertIn(file_path, items_mapping)
        self.assertIsInstance(items_mapping[file_path], list)

    def test_load_api_folder(self):
        path = os.path.join(os.getcwd(), "tests", "api")
        api_definition_mapping = loader.load_api_folder(path)
        self.assertIn("get_token", api_definition_mapping)
        self.assertIn("request", api_definition_mapping["get_token"])

    def test_load_project_tests(self):
        loader.load_project_tests(os.path.join(os.getcwd(), "tests"))
        self.assertIn("get_token", self.tests_def_mapping["api"])
        self.assertEqual(self.project_mapping["env"]["PROJECT_KEY"], "ABCDEFGH")
