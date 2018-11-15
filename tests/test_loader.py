
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
        project_mapping = loader.load_project_tests(os.path.join(os.getcwd(), "httprunner"))
        self.assertNotIn("alter_response", project_mapping["functions"])

        project_mapping = loader.load_project_tests(os.path.join(os.getcwd(), "tests"))
        self.assertIn("alter_response", project_mapping["functions"])

        is_status_code_200 = project_mapping["functions"]["is_status_code_200"]
        self.assertTrue(is_status_code_200(200))
        self.assertFalse(is_status_code_200(500))

    def test_load_debugtalk_py(self):
        project_working_directory, debugtalk_functions = loader.load_debugtalk_py("tests/data/demo_testcase.yml")
        self.assertEqual(
            project_working_directory,
            os.path.join(os.getcwd(), "tests")
        )
        self.assertIn("gen_md5", debugtalk_functions)

        project_working_directory, debugtalk_functions = loader.load_debugtalk_py("tests/base.py")
        self.assertEqual(
            project_working_directory,
            os.path.join(os.getcwd(), "tests")
        )
        self.assertIn("gen_md5", debugtalk_functions)

        project_working_directory, debugtalk_functions = loader.load_debugtalk_py("httprunner/__init__.py")
        self.assertEqual(
            project_working_directory,
            os.getcwd()
        )
        self.assertEqual(debugtalk_functions, {})

    def test_load_tests(self):
        testcase_file_path = os.path.join(
            os.getcwd(), 'tests/data/demo_testcase.yml')
        testcases = loader.load_tests(testcase_file_path)
        self.assertIsInstance(testcases, list)
        self.assertEqual(
            testcases[0]["config"]["request"],
            '${get_default_request()}'
        )
        self.assertEqual(testcases[0]["config"]["name"], '123$var_a')
        self.assertIn(
            "sum_two",
            testcases[0]["config"]["functions"]
        )


class TestSuiteLoader(unittest.TestCase):

    @classmethod
    def setUpClass(cls):
        cls.project_mapping = loader.load_project_tests(os.path.join(os.getcwd(), "tests"))

    def test_load_teststeps(self):
        test_block = {
            "name": "setup and reset all (override).",
            "suite": "setup_and_reset($device_sn)",
            "output": ["token", "device_sn"]
        }
        teststeps = loader._load_teststeps(test_block, self.project_mapping)
        self.assertEqual(len(teststeps), 2)
        self.assertEqual(teststeps[0]["name"], "get token")
        self.assertEqual(teststeps[1]["name"], "reset all users")

    def test_load_testcase(self):
        raw_testcase = loader.load_file("tests/testcases/smoketest.yml")
        testcase = loader._load_testcase(raw_testcase, self.project_mapping)
        self.assertEqual(testcase["config"]["name"], "smoketest")
        self.assertIn("device_sn", testcase["config"]["variables"][0])
        self.assertEqual(len(testcase["teststeps"]), 8)
        self.assertEqual(testcase["teststeps"][0]["name"], "get token")

    def test_get_block_by_name(self):
        ref_call = "get_user($uid, $token)"
        block = loader._get_block_by_name(ref_call, "def-api", self.project_mapping)
        self.assertEqual(block["request"]["url"], "/api/users/$uid")
        self.assertEqual(block["function_meta"]["func_name"], "get_user")
        self.assertEqual(block["function_meta"]["args"], ['$uid', '$token'])

    def test_get_block_by_name_args_mismatch(self):
        ref_call = "get_user($uid, $token, $var)"
        with self.assertRaises(exceptions.ParamsError):
            loader._get_block_by_name(ref_call, "def-api", self.project_mapping)

    def test_override_block(self):
        def_block = loader._get_block_by_name(
            "get_token($user_agent, $device_sn, $os_platform, $app_version)",
            "def-api",
            self.project_mapping
        )
        test_block = {
            "name": "override block",
            "times": 3,
            "variables": [
                {"var": 123}
            ],
            'request': {
                'url': '/api/get-token',
                'method': 'POST',
                'headers': {'user_agent': '$user_agent', 'device_sn': '$device_sn', 'os_platform': '$os_platform', 'app_version': '$app_version'},
                'json': {'sign': '${get_sign($user_agent, $device_sn, $os_platform, $app_version)}'}
            },
            'validate': [
                {'eq': ['status_code', 201]},
                {'len_eq': ['content.token', 32]}
            ]
        }

        extended_block = loader._extend_block(test_block, def_block)
        self.assertEqual(extended_block["name"], "override block")
        self.assertIn({'var': 123}, extended_block["variables"])
        self.assertIn({'check': 'status_code', 'expect': 201, 'comparator': 'eq'}, extended_block["validate"])
        self.assertIn({'check': 'content.token', 'comparator': 'len_eq', 'expect': 32}, extended_block["validate"])
        self.assertEqual(extended_block["times"], 3)

    def test_get_test_definition_api(self):
        api_def = loader._get_test_definition("get_headers", "def-api", self.project_mapping)
        self.assertEqual(api_def["request"]["url"], "/headers")
        self.assertEqual(len(api_def["setup_hooks"]), 2)
        self.assertEqual(len(api_def["teardown_hooks"]), 1)

        with self.assertRaises(exceptions.ApiNotFound):
            loader._get_test_definition("get_token_XXX", "def-api", self.project_mapping)

    def test_get_test_definition_suite(self):
        api_def = loader._get_test_definition("create_and_check", "def-testcase", self.project_mapping)
        self.assertEqual(api_def["config"]["name"], "create user and check result.")

        with self.assertRaises(exceptions.TestcaseNotFound):
            loader._get_test_definition("create_and_check_XXX", "def-testcase", self.project_mapping)

    def test_extend_validators(self):
        def_validators = [
            {'eq': ['v1', 200]},
            {"check": "s2", "expect": 16, "comparator": "len_eq"}
        ]
        current_validators = [
            {"check": "v1", "expect": 201},
            {'len_eq': ['s3', 12]}
        ]

        extended_validators = loader._extend_validators(def_validators, current_validators)
        self.assertIn(
            {"check": "v1", "expect": 201, "comparator": "eq"},
            extended_validators
        )
        self.assertIn(
            {"check": "s2", "expect": 16, "comparator": "len_eq"},
            extended_validators
        )
        self.assertIn(
            {"check": "s3", "expect": 12, "comparator": "len_eq"},
            extended_validators
        )

    def test_extend_validators_with_dict(self):
        def_validators = [
            {'eq': ["a", {"v": 1}]},
            {'eq': [{"b": 1}, 200]}
        ]
        current_validators = [
            {'len_eq': ['s3', 12]},
            {'eq': [{"b": 1}, 201]}
        ]

        extended_validators = loader._extend_validators(def_validators, current_validators)
        self.assertEqual(len(extended_validators), 3)
        self.assertIn({'check': {'b': 1}, 'expect': 201, 'comparator': 'eq'}, extended_validators)
        self.assertNotIn({'check': {'b': 1}, 'expect': 200, 'comparator': 'eq'}, extended_validators)

    def test_extend_variables(self):
        def_variables = [{"var1": "val1"}, {"var2": "val2"}]
        ref_variables = [{"var1": "val111"}, {"var3": "val3"}]

        extended_variables = loader._extend_variables(def_variables, ref_variables)
        self.assertIn(
            {"var1": "val111"},
            extended_variables
        )
        self.assertIn(
            {"var2": "val2"},
            extended_variables
        )
        self.assertIn(
            {"var3": "val3"},
            extended_variables
        )

    def test_load_testcases_by_path_files(self):
        testcases_list = []

        # absolute file path
        path = os.path.join(
            os.getcwd(), 'tests/data/demo_testcase_hardcode.json')
        testcases_list = loader.load_tests(path)
        self.assertEqual(len(testcases_list), 1)
        self.assertEqual(len(testcases_list[0]["teststeps"]), 3)
        self.assertIn("get_sign", testcases_list[0]["config"]["functions"])

        # relative file path
        path = 'tests/data/demo_testcase_hardcode.yml'
        testcases_list = loader.load_tests(path)
        self.assertEqual(len(testcases_list), 1)
        self.assertEqual(len(testcases_list[0]["teststeps"]), 3)
        self.assertIn("get_sign", testcases_list[0]["config"]["functions"])

        # list/set container with file(s)
        path = [
            os.path.join(os.getcwd(), 'tests/data/demo_testcase_hardcode.json'),
            'tests/data/demo_testcase_hardcode.yml'
        ]
        testcases_list = loader.load_tests(path)
        self.assertEqual(len(testcases_list), 2)
        self.assertEqual(len(testcases_list[0]["teststeps"]), 3)
        self.assertEqual(len(testcases_list[1]["teststeps"]), 3)
        testcases_list.extend(testcases_list)
        self.assertEqual(len(testcases_list), 4)

        for testcase in testcases_list:
            for teststep in testcase["teststeps"]:
                self.assertIn('name', teststep)
                self.assertIn('request', teststep)
                self.assertIn('url', teststep['request'])
                self.assertIn('method', teststep['request'])

    def test_load_testcases_by_path_folder(self):
        # absolute folder path
        path = os.path.join(os.getcwd(), 'tests/data')
        testcase_list_1 = loader.load_tests(path)
        self.assertGreater(len(testcase_list_1), 4)

        # relative folder path
        path = 'tests/data/'
        testcase_list_2 = loader.load_tests(path)
        self.assertEqual(len(testcase_list_1), len(testcase_list_2))

        # list/set container with file(s)
        path = [
            os.path.join(os.getcwd(), 'tests/data'),
            'tests/data/'
        ]
        testcase_list_3 = loader.load_tests(path)
        self.assertEqual(len(testcase_list_3), 2 * len(testcase_list_1))

    def test_load_testcases_by_path_not_exist(self):
        # absolute folder path
        path = os.path.join(os.getcwd(), 'tests/data_not_exist')
        with self.assertRaises(exceptions.FileNotFound):
            loader.load_tests(path)

        # relative folder path
        path = 'tests/data_not_exist'
        with self.assertRaises(exceptions.FileNotFound):
            loader.load_tests(path)

        # list/set container with file(s)
        path = [
            os.path.join(os.getcwd(), 'tests/data_not_exist'),
            'tests/data_not_exist/'
        ]
        with self.assertRaises(exceptions.FileNotFound):
            loader.load_tests(path)

    def test_load_testcases_by_path_layered(self):
        path = os.path.join(
            os.getcwd(), 'tests/data/demo_testcase_layer.yml')
        testcases_list = loader.load_tests(path)
        self.assertIn("variables", testcases_list[0]["config"])
        self.assertIn("request", testcases_list[0]["config"])
        self.assertIn("request", testcases_list[0]["teststeps"][0])
        self.assertIn("url", testcases_list[0]["teststeps"][0]["request"])
        self.assertIn("validate", testcases_list[0]["teststeps"][0])

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
        self.assertIn("function_meta", api_definition_mapping["get_token"])

    def test_load_testcases_folder(self):
        path = os.path.join(os.getcwd(), "tests", "suite")
        testcases_definition_mapping = loader.load_test_folder(path)

        self.assertIn("setup_and_reset", testcases_definition_mapping)
        self.assertIn("create_and_check", testcases_definition_mapping)
        self.assertEqual(
            testcases_definition_mapping["setup_and_reset"]["config"]["name"],
            "setup and reset all."
        )
        self.assertEqual(
            testcases_definition_mapping["setup_and_reset"]["function_meta"]["func_name"],
            "setup_and_reset"
        )

    def test_load_testsuites_folder(self):
        path = os.path.join(os.getcwd(), "tests", "testcases")
        testsuites_definition_mapping = loader.load_test_folder(path)

        testsute_path = os.path.join(os.getcwd(), "tests", "testcases", "smoketest.yml")
        self.assertIn(
            testsute_path,
            testsuites_definition_mapping
        )
        self.assertEqual(
            testsuites_definition_mapping[testsute_path]["config"]["name"],
            "smoketest"
        )

    def test_load_project_tests(self):
        project_mapping = loader.load_project_tests(os.path.join(os.getcwd(), "tests"))
        self.assertIn("get_token", project_mapping["def-api"])
        self.assertIn("setup_and_reset", project_mapping["def-testcase"])
        self.assertEqual(project_mapping["env"]["PROJECT_KEY"], "ABCDEFGH")

    def test_load_locust_tests(self):
        path = os.path.join(
            os.getcwd(), 'tests/data/demo_locust.yml')
        locust_tests = loader.load_locust_tests(path)
        self.assertEqual(len(locust_tests["tests"]), 10)
        self.assertEqual(locust_tests["tests"][0][0]["name"], "index")
        self.assertEqual(locust_tests["tests"][9][0]["name"], "user-agent")
