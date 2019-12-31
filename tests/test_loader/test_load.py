import os
import unittest

from httprunner import exceptions
from httprunner.loader import load
from httprunner.loader.buildup import load_test_file


class TestFileLoader(unittest.TestCase):

    def test_load_yaml_file_file_format_error(self):
        yaml_tmp_file = "tests/data/tmp.yml"
        # create empty yaml file
        with open(yaml_tmp_file, 'w') as f:
            f.write("")

        with self.assertRaises(exceptions.FileFormatError):
            load_test_file(yaml_tmp_file)

        os.remove(yaml_tmp_file)

        # create invalid format yaml file
        with open(yaml_tmp_file, 'w') as f:
            f.write("abc")

        with self.assertRaises(exceptions.FileFormatError):
            load_test_file(yaml_tmp_file)

        os.remove(yaml_tmp_file)

    def test_load_json_file_file_format_error(self):
        json_tmp_file = "tests/data/tmp.json"
        # create empty file
        with open(json_tmp_file, 'w') as f:
            f.write("")

        with self.assertRaises(exceptions.FileFormatError):
            load_test_file(json_tmp_file)

        os.remove(json_tmp_file)

        # create empty json file
        with open(json_tmp_file, 'w') as f:
            f.write("{}")

        with self.assertRaises(exceptions.FileFormatError):
            load_test_file(json_tmp_file)

        os.remove(json_tmp_file)

        # create invalid format json file
        with open(json_tmp_file, 'w') as f:
            f.write("abc")

        with self.assertRaises(exceptions.FileFormatError):
            load_test_file(json_tmp_file)

        os.remove(json_tmp_file)

    def test_load_testcases_bad_filepath(self):
        testcase_file_path = os.path.join(os.getcwd(), 'tests/data/demo')
        with self.assertRaises(exceptions.FileNotFound):
            load.load_file(testcase_file_path)

    def test_load_json_testcases(self):
        testcase_file_path = os.path.join(
            os.getcwd(), 'tests/data/demo_testcase_hardcode.json')
        testcases = load.load_file(testcase_file_path)
        self.assertEqual(len(testcases), 3)
        test = testcases[0]["test"]
        self.assertIn('name', test)
        self.assertIn('request', test)
        self.assertIn('url', test['request'])
        self.assertIn('method', test['request'])

    def test_load_yaml_testcases(self):
        testcase_file_path = os.path.join(
            os.getcwd(), 'tests/data/demo_testcase_hardcode.yml')
        testcases = load.load_file(testcase_file_path)
        self.assertEqual(len(testcases), 3)
        test = testcases[0]["test"]
        self.assertIn('name', test)
        self.assertIn('request', test)
        self.assertIn('url', test['request'])
        self.assertIn('method', test['request'])

    def test_load_csv_file_one_parameter(self):
        csv_file_path = os.path.join(
            os.getcwd(), 'tests/data/user_agent.csv')
        csv_content = load.load_file(csv_file_path)
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
        csv_content = load.load_file(csv_file_path)
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
        file2 = os.path.join(os.getcwd(), 'tests', 'api', 'reset_all.yml')

        files = load.load_folder_files(folder, recursive=False)
        self.assertEqual(files, [])

        files = load.load_folder_files(folder)
        self.assertIn(file2, files)
        self.assertNotIn(file1, files)

        files = load.load_folder_files("not_existed_foulder", recursive=False)
        self.assertEqual([], files)

        files = load.load_folder_files(file2, recursive=False)
        self.assertEqual([], files)

    def test_load_dot_env_file(self):
        dot_env_path = os.path.join(
            os.getcwd(), "tests", ".env"
        )
        env_variables_mapping = load.load_dot_env_file(dot_env_path)
        self.assertIn("PROJECT_KEY", env_variables_mapping)
        self.assertEqual(env_variables_mapping["UserName"], "debugtalk")

    def test_load_custom_dot_env_file(self):
        dot_env_path = os.path.join(
            os.getcwd(), "tests", "data", "test.env"
        )
        env_variables_mapping = load.load_dot_env_file(dot_env_path)
        self.assertIn("PROJECT_KEY", env_variables_mapping)
        self.assertEqual(env_variables_mapping["UserName"], "test")
        self.assertEqual(env_variables_mapping["content_type"], "application/json; charset=UTF-8")

    def test_load_env_path_not_exist(self):
        dot_env_path = os.path.join(
            os.getcwd(), "tests", "data",
        )
        env_variables_mapping = load.load_dot_env_file(dot_env_path)
        self.assertEqual(env_variables_mapping, {})
