import os
import unittest

from httprunner import exceptions, loader


class TestLoader(unittest.TestCase):
    def test_load_testcase_file(self):
        path = "examples/postman_echo/request_methods/request_with_variables.yml"
        testcase_obj = loader.load_testcase_file(path)
        self.assertEqual(
            testcase_obj.config.name, "request methods testcase with variables"
        )
        self.assertEqual(len(testcase_obj.teststeps), 4)

    def test_load_json_file_file_format_error(self):
        json_tmp_file = "tmp.json"
        # create empty file
        with open(json_tmp_file, "w") as f:
            f.write("")

        with self.assertRaises(exceptions.FileFormatError):
            loader._load_json_file(json_tmp_file)

        os.remove(json_tmp_file)

        # create empty json file
        with open(json_tmp_file, "w") as f:
            f.write("{}")

        loader._load_json_file(json_tmp_file)
        os.remove(json_tmp_file)

        # create invalid format json file
        with open(json_tmp_file, "w") as f:
            f.write("abc")

        with self.assertRaises(exceptions.FileFormatError):
            loader._load_json_file(json_tmp_file)

        os.remove(json_tmp_file)

    def test_load_testcases_bad_filepath(self):
        testcase_file_path = os.path.join(os.getcwd(), "examples/data/demo")
        with self.assertRaises(exceptions.FileNotFound):
            loader.load_testcase_file(testcase_file_path)

    def test_load_csv_file_one_parameter(self):
        csv_file_path = os.path.join(os.getcwd(), "examples/httpbin/user_agent.csv")
        csv_content = loader.load_csv_file(csv_file_path)
        self.assertEqual(
            csv_content,
            [
                {"user_agent": "iOS/10.1"},
                {"user_agent": "iOS/10.2"},
                {"user_agent": "iOS/10.3"},
            ],
        )

    def test_load_csv_file_multiple_parameters(self):
        csv_file_path = os.path.join(os.getcwd(), "examples/httpbin/account.csv")
        csv_content = loader.load_csv_file(csv_file_path)
        self.assertEqual(
            csv_content,
            [
                {"username": "test1", "password": "111111"},
                {"username": "test2", "password": "222222"},
                {"username": "test3", "password": "333333"},
            ],
        )

    def test_load_folder_files(self):
        folder = os.path.join(os.getcwd(), "examples")
        file1 = os.path.join(os.getcwd(), "examples", "test_utils.py")
        file2 = os.path.join(os.getcwd(), "examples", "httpbin", "hooks.yml")

        files = loader.load_folder_files(folder, recursive=False)
        self.assertEqual(files, [])

        files = loader.load_folder_files(folder)
        self.assertIn(file2, files)
        self.assertNotIn(file1, files)

        files = loader.load_folder_files("not_existed_foulder", recursive=False)
        self.assertEqual([], files)

        files = loader.load_folder_files(file2, recursive=False)
        self.assertEqual([], files)

    def test_load_custom_dot_env_file(self):
        dot_env_path = os.path.join(os.getcwd(), "examples", "httpbin", "test.env")
        env_variables_mapping = loader.load_dot_env_file(dot_env_path)
        self.assertIn("PROJECT_KEY", env_variables_mapping)
        self.assertEqual(env_variables_mapping["UserName"], "test")
        self.assertEqual(
            env_variables_mapping["content_type"], "application/json; charset=UTF-8"
        )

    def test_load_env_path_not_exist(self):
        dot_env_path = os.path.join(
            os.getcwd(),
            "tests",
            "data",
        )
        env_variables_mapping = loader.load_dot_env_file(dot_env_path)
        self.assertEqual(env_variables_mapping, {})

    def test_locate_file(self):
        with self.assertRaises(exceptions.FileNotFound):
            loader.locate_file(os.getcwd(), "debugtalk.py")

        with self.assertRaises(exceptions.FileNotFound):
            loader.locate_file("", "debugtalk.py")

        start_path = os.path.join(os.getcwd(), "examples", "httpbin")
        self.assertEqual(
            loader.locate_file(start_path, "debugtalk.py"),
            os.path.join(os.getcwd(), "examples", "httpbin", "debugtalk.py"),
        )
        self.assertEqual(
            loader.locate_file("examples/httpbin/", "debugtalk.py"),
            os.path.join(os.getcwd(), "examples", "httpbin", "debugtalk.py"),
        )
        self.assertEqual(
            loader.locate_file("examples/httpbin/", "debugtalk.py"),
            os.path.join(os.getcwd(), "examples", "httpbin", "debugtalk.py"),
        )
