import os
import shutil
from collections import OrderedDict

from httprunner import exception, utils
from tests.base import ApiServerUnittest


class TestUtils(ApiServerUnittest):

    def test_remove_prefix(self):
        full_url = "http://debugtalk.com/post/123"
        prefix = "http://debugtalk.com"
        self.assertEqual(
            utils.remove_prefix(full_url, prefix),
            "/post/123"
        )

    def test_load_folder_files(self):
        folder = os.path.join(os.getcwd(), 'tests')
        file1 = os.path.join(os.getcwd(), 'tests', 'test_utils.py')
        file2 = os.path.join(os.getcwd(), 'tests', 'data', 'demo_binds.yml')

        files = utils.load_folder_files(folder, recursive=False)
        self.assertNotIn(file2, files)

        files = utils.load_folder_files(folder)
        self.assertIn(file2, files)
        self.assertNotIn(file1, files)

        files_1 = utils.load_folder_files(folder)
        api_file = os.path.join(os.getcwd(), 'tests', 'api', 'demo.yml')
        self.assertEqual(files_1[0], api_file)

        files_2 = utils.load_folder_files(folder)
        api_file = os.path.join(os.getcwd(), 'tests', 'api', 'demo.yml')
        self.assertEqual(files_2[0], api_file)
        self.assertEqual(len(files_1), len(files_2))

        files = utils.load_folder_files("not_existed_foulder", recursive=False)
        self.assertEqual([], files)

        files = utils.load_folder_files(file2, recursive=False)
        self.assertEqual([], files)

    def test_query_json(self):
        json_content = {
            "ids": [1, 2, 3, 4],
            "person": {
                "name": {
                    "first_name": "Leo",
                    "last_name": "Lee",
                },
                "age": 29,
                "cities": ["Guangzhou", "Shenzhen"]
            }
        }
        query = "ids.2"
        result = utils.query_json(json_content, query)
        self.assertEqual(result, 3)

        query = "ids.str_key"
        with self.assertRaises(exception.ParseResponseError):
            utils.query_json(json_content, query)

        query = "ids.5"
        with self.assertRaises(exception.ParseResponseError):
            utils.query_json(json_content, query)

        query = "person.age"
        result = utils.query_json(json_content, query)
        self.assertEqual(result, 29)

        query = "person.not_exist_key"
        with self.assertRaises(exception.ParseResponseError):
            utils.query_json(json_content, query)

        query = "person.cities.0"
        result = utils.query_json(json_content, query)
        self.assertEqual(result, "Guangzhou")

        query = "person.name.first_name"
        result = utils.query_json(json_content, query)
        self.assertEqual(result, "Leo")

    def test_query_json_content_is_text(self):
        json_content = ""
        query = "key"
        with self.assertRaises(exception.ResponseError):
            utils.query_json(json_content, query)

        json_content = "<html><body>content</body></html>"
        query = "key"
        with self.assertRaises(exception.ParseResponseError):
            utils.query_json(json_content, query)

    def test_match_expected(self):
        self.assertTrue(utils.match_expected(1, 1, "eq"))
        self.assertTrue(utils.match_expected("abc", "abc", "=="))
        self.assertTrue(utils.match_expected("abc", "abc"))

        with self.assertRaises(exception.ValidationError):
            utils.match_expected(123, "123", "eq")
        with self.assertRaises(exception.ValidationError):
            utils.match_expected(123, "123")

        self.assertTrue(utils.match_expected(1, 2, "lt"))
        self.assertTrue(utils.match_expected(1, 1, "le"))
        self.assertTrue(utils.match_expected(2, 1, "gt"))
        self.assertTrue(utils.match_expected(1, 1, "ge"))
        self.assertTrue(utils.match_expected(123, "123", "ne"))

        self.assertTrue(utils.match_expected("123", 3, "len_eq"))
        self.assertTrue(utils.match_expected("123", 2, "len_gt"))
        self.assertTrue(utils.match_expected("123", 3, "len_ge"))
        self.assertTrue(utils.match_expected("123", 4, "len_lt"))
        self.assertTrue(utils.match_expected("123", 3, "len_le"))

        self.assertTrue(utils.match_expected("123abc456", "3ab", "contains"))
        self.assertTrue(utils.match_expected(['1', '2'], "1", "contains"))
        self.assertTrue(utils.match_expected({'a':1, 'b':2}, "a", "contains"))
        self.assertTrue(utils.match_expected("3ab", "123abc456", "contained_by"))

        self.assertTrue(utils.match_expected("123abc456", "^123\w+456$", "regex"))
        with self.assertRaises(exception.ValidationError):
            utils.match_expected("123abc456", "^12b.*456$", "regex")

        with self.assertRaises(exception.ParamsError):
            utils.match_expected(1, 2, "not_supported_comparator")

        self.assertTrue(utils.match_expected("abc123", "ab", "startswith"))
        self.assertTrue(utils.match_expected("123abc", 12, "startswith"))
        self.assertTrue(utils.match_expected(12345, 123, "startswith"))
        self.assertTrue(utils.match_expected("abc123", 23, "endswith"))
        self.assertTrue(utils.match_expected("123abc", "abc", "endswith"))
        self.assertTrue(utils.match_expected(12345, 45, "endswith"))

        self.assertTrue(utils.match_expected(None, None, "eq"))
        with self.assertRaises(exception.ValidationError):
            utils.match_expected(None, 3, "len_eq")
        with self.assertRaises(exception.ValidationError):
            utils.match_expected("abc", None, "gt")

    def test_deep_update_dict(self):
        origin_dict = {'a': 1, 'b': {'c': 3, 'd': 4}, 'f': 6}
        override_dict = {'a': 2, 'b': {'c': 33, 'e': 5}, 'g': 7}
        updated_dict = utils.deep_update_dict(origin_dict, override_dict)
        self.assertEqual(
            updated_dict,
            {'a': 2, 'b': {'c': 33, 'd': 4, 'e': 5}, 'f': 6, 'g': 7}
        )

    def test_get_imported_module(self):
        imported_module = utils.get_imported_module("os")
        self.assertIn("walk", dir(imported_module))

    def test_filter_module_functions(self):
        imported_module = utils.get_imported_module("httprunner.utils")
        self.assertIn("PYTHON_VERSION", dir(imported_module))

        functions_dict = utils.filter_module(imported_module, "function")
        self.assertIn("filter_module", functions_dict)
        self.assertNotIn("PYTHON_VERSION", functions_dict)

    def test_get_imported_module_from_file(self):
        imported_module = utils.get_imported_module_from_file("tests/data/debugtalk.py")
        self.assertIn("gen_md5", dir(imported_module))

        functions_dict = utils.filter_module(imported_module, "function")
        self.assertIn("gen_md5", functions_dict)
        self.assertNotIn("PYTHON_VERSION", functions_dict)

        with self.assertRaises(exception.FileNotFoundError):
            utils.get_imported_module_from_file("tests/data/debugtalk2.py")

    def test_search_conf_function(self):
        gen_md5 = utils.search_conf_item("tests/data/demo_binds.yml", "function", "gen_md5")
        self.assertTrue(utils.is_function(("gen_md5", gen_md5)))
        self.assertEqual(gen_md5("abc"), "900150983cd24fb0d6963f7d28e17f72")

        gen_md5 = utils.search_conf_item("tests/data/subfolder/test.yml", "function", "gen_md5")
        self.assertTrue(utils.is_function(("_", gen_md5)))
        self.assertEqual(gen_md5("abc"), "900150983cd24fb0d6963f7d28e17f72")

        with self.assertRaises(exception.FunctionNotFound):
            utils.search_conf_item("tests/data/subfolder/test.yml", "function", "func_not_exist")

        with self.assertRaises(exception.FunctionNotFound):
            utils.search_conf_item("/user/local/bin", "function", "gen_md5")

    def test_search_conf_variable(self):
        SECRET_KEY = utils.search_conf_item("tests/data/demo_binds.yml", "variable", "SECRET_KEY")
        self.assertTrue(utils.is_variable(("SECRET_KEY", SECRET_KEY)))
        self.assertEqual(SECRET_KEY, "DebugTalk")

        SECRET_KEY = utils.search_conf_item("tests/data/subfolder/test.yml", "variable", "SECRET_KEY")
        self.assertTrue(utils.is_variable(("SECRET_KEY", SECRET_KEY)))
        self.assertEqual(SECRET_KEY, "DebugTalk")

        with self.assertRaises(exception.VariableNotFound):
            utils.search_conf_item("tests/data/subfolder/test.yml", "variable", "variable_not_exist")

        with self.assertRaises(exception.VariableNotFound):
            utils.search_conf_item("/user/local/bin", "variable", "SECRET_KEY")

    def test_is_variable(self):
        var1 = 123
        var2 = "abc"
        self.assertTrue(utils.is_variable(("var1", var1)))
        self.assertTrue(utils.is_variable(("var2", var2)))

        __var = 123
        self.assertFalse(utils.is_variable(("__var", __var)))

        func = lambda x: x + 1
        self.assertFalse(utils.is_variable(("func", func)))

        self.assertFalse(utils.is_variable(("os", os)))
        self.assertFalse(utils.is_variable(("utils", utils)))

    def test_handle_config_key_case(self):
        origin_dict = {
            "Name": "test",
            "Request": {
                "url": "http://127.0.0.1:5000",
                "METHOD": "POST",
                "Headers": {
                    "Accept": "application/json",
                    "User-Agent": "ios/9.3"
                }
            }
        }
        new_dict = utils.lower_config_dict_key(origin_dict)
        self.assertIn("name", new_dict)
        self.assertIn("request", new_dict)
        self.assertIn("method", new_dict["request"])
        self.assertIn("headers", new_dict["request"])
        self.assertIn("Accept", new_dict["request"]["headers"])
        self.assertIn("User-Agent", new_dict["request"]["headers"])

        origin_dict = {
            "Name": "test",
            "Request": "$default_request"
        }
        new_dict = utils.lower_config_dict_key(origin_dict)
        self.assertIn("$default_request", new_dict["request"])

    def test_lower_dict_keys(self):
        request_dict = {
            "url": "http://127.0.0.1:5000",
            "METHOD": "POST",
            "Headers": {
                "Accept": "application/json",
                "User-Agent": "ios/9.3"
            }
        }
        new_request_dict = utils.lower_dict_keys(request_dict)
        self.assertIn("method", new_request_dict)
        self.assertIn("headers", new_request_dict)
        self.assertIn("Accept", new_request_dict["headers"])
        self.assertIn("User-Agent", new_request_dict["headers"])

        request_dict = "$default_request"
        new_request_dict = utils.lower_dict_keys(request_dict)
        self.assertEqual("$default_request", request_dict)

        request_dict = None
        new_request_dict = utils.lower_dict_keys(request_dict)
        self.assertEqual(None, request_dict)

    def test_convert_to_order_dict(self):
        map_list = [
            {"a": 1},
            {"b": 2}
        ]
        ordered_dict = utils.convert_to_order_dict(map_list)
        self.assertIsInstance(ordered_dict, dict)
        self.assertIn("a", ordered_dict)

    def test_update_ordered_dict(self):
        map_list = [
            {"a": 1},
            {"b": 2}
        ]
        ordered_dict = utils.convert_to_order_dict(map_list)
        override_mapping = {"a": 3, "c": 4}
        new_dict = utils.update_ordered_dict(ordered_dict, override_mapping)
        self.assertEqual(3, new_dict["a"])
        self.assertEqual(4, new_dict["c"])

    def test_override_variables_binds(self):
        map_list = [
            {"a": 1},
            {"b": 2}
        ]
        override_mapping = {"a": 3, "c": 4}
        new_dict = utils.override_variables_binds(map_list, override_mapping)
        self.assertEqual(3, new_dict["a"])
        self.assertEqual(4, new_dict["c"])

        map_list = OrderedDict(
            {
                "a": 1,
                "b": 2
            }
        )
        override_mapping = {"a": 3, "c": 4}
        new_dict = utils.override_variables_binds(map_list, override_mapping)
        self.assertEqual(3, new_dict["a"])
        self.assertEqual(4, new_dict["c"])

        map_list = "invalid"
        override_mapping = {"a": 3, "c": 4}
        with self.assertRaises(exception.ParamsError):
            utils.override_variables_binds(map_list, override_mapping)

    def test_create_scaffold(self):
        project_path = os.path.join(os.getcwd(), "projectABC")
        utils.create_scaffold(project_path)
        self.assertTrue(os.path.isdir(os.path.join(project_path, "tests")))
        self.assertTrue(os.path.isdir(os.path.join(project_path, "tests", "api")))
        self.assertTrue(os.path.isdir(os.path.join(project_path, "tests", "suite")))
        self.assertTrue(os.path.isdir(os.path.join(project_path, "tests", "testcases")))
        self.assertTrue(os.path.isfile(os.path.join(project_path, "tests", "debugtalk.py")))
        shutil.rmtree(project_path)
