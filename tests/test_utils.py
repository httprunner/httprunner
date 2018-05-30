import os
import shutil
import unittest

from httprunner import exception, utils
from httprunner.compat import OrderedDict
from httprunner.utils import FileUtils
from tests.base import ApiServerUnittest


class TestFileUtils(unittest.TestCase):

    def test_load_yaml_file_file_format_error(self):
        yaml_tmp_file = "tests/data/tmp.yml"
        # create empty yaml file
        with open(yaml_tmp_file, 'w') as f:
            f.write("")

        with self.assertRaises(exception.FileFormatError):
            FileUtils._load_yaml_file(yaml_tmp_file)

        os.remove(yaml_tmp_file)

        # create invalid format yaml file
        with open(yaml_tmp_file, 'w') as f:
            f.write("abc")

        with self.assertRaises(exception.FileFormatError):
            FileUtils._load_yaml_file(yaml_tmp_file)

        os.remove(yaml_tmp_file)


    def test_load_json_file_file_format_error(self):
        json_tmp_file = "tests/data/tmp.json"
        # create empty file
        with open(json_tmp_file, 'w') as f:
            f.write("")

        with self.assertRaises(exception.FileFormatError):
            FileUtils._load_json_file(json_tmp_file)

        os.remove(json_tmp_file)

        # create empty json file
        with open(json_tmp_file, 'w') as f:
            f.write("{}")

        with self.assertRaises(exception.FileFormatError):
            FileUtils._load_json_file(json_tmp_file)

        os.remove(json_tmp_file)

        # create invalid format json file
        with open(json_tmp_file, 'w') as f:
            f.write("abc")

        with self.assertRaises(exception.FileFormatError):
            FileUtils._load_json_file(json_tmp_file)

        os.remove(json_tmp_file)

    def test_load_testcases_bad_filepath(self):
        testcase_file_path = os.path.join(os.getcwd(), 'tests/data/demo')
        with self.assertRaises(exception.FileNotFoundError):
            FileUtils.load_file(testcase_file_path)

    def test_load_json_testcases(self):
        testcase_file_path = os.path.join(
            os.getcwd(), 'tests/data/demo_testset_hardcode.json')
        testcases = FileUtils.load_file(testcase_file_path)
        self.assertEqual(len(testcases), 3)
        test = testcases[0]["test"]
        self.assertIn('name', test)
        self.assertIn('request', test)
        self.assertIn('url', test['request'])
        self.assertIn('method', test['request'])

    def test_load_yaml_testcases(self):
        testcase_file_path = os.path.join(
            os.getcwd(), 'tests/data/demo_testset_hardcode.yml')
        testcases = FileUtils.load_file(testcase_file_path)
        self.assertEqual(len(testcases), 3)
        test = testcases[0]["test"]
        self.assertIn('name', test)
        self.assertIn('request', test)
        self.assertIn('url', test['request'])
        self.assertIn('method', test['request'])

    def test_load_csv_file_one_parameter(self):
        csv_file_path = os.path.join(
            os.getcwd(), 'tests/data/user_agent.csv')
        csv_content = FileUtils.load_file(csv_file_path)
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
        csv_content = FileUtils.load_file(csv_file_path)
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

        files = FileUtils.load_folder_files(folder, recursive=False)
        self.assertNotIn(file2, files)

        files = FileUtils.load_folder_files(folder)
        self.assertIn(file2, files)
        self.assertNotIn(file1, files)

        files = FileUtils.load_folder_files(folder)
        api_file = os.path.join(os.getcwd(), 'tests', 'api', 'basic.yml')
        self.assertIn(api_file, files)

        files = FileUtils.load_folder_files("not_existed_foulder", recursive=False)
        self.assertEqual([], files)

        files = FileUtils.load_folder_files(file2, recursive=False)
        self.assertEqual([], files)


class TestUtils(ApiServerUnittest):

    def test_remove_prefix(self):
        full_url = "http://debugtalk.com/post/123"
        prefix = "http://debugtalk.com"
        self.assertEqual(
            utils.remove_prefix(full_url, prefix),
            "/post/123"
        )

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

    def test_get_uniform_comparator(self):
        self.assertEqual(utils.get_uniform_comparator("eq"), "equals")
        self.assertEqual(utils.get_uniform_comparator("=="), "equals")
        self.assertEqual(utils.get_uniform_comparator("lt"), "less_than")
        self.assertEqual(utils.get_uniform_comparator("le"), "less_than_or_equals")
        self.assertEqual(utils.get_uniform_comparator("gt"), "greater_than")
        self.assertEqual(utils.get_uniform_comparator("ge"), "greater_than_or_equals")
        self.assertEqual(utils.get_uniform_comparator("ne"), "not_equals")

        self.assertEqual(utils.get_uniform_comparator("str_eq"), "string_equals")
        self.assertEqual(utils.get_uniform_comparator("len_eq"), "length_equals")
        self.assertEqual(utils.get_uniform_comparator("count_eq"), "length_equals")

        self.assertEqual(utils.get_uniform_comparator("len_gt"), "length_greater_than")
        self.assertEqual(utils.get_uniform_comparator("count_gt"), "length_greater_than")
        self.assertEqual(utils.get_uniform_comparator("count_greater_than"), "length_greater_than")

        self.assertEqual(utils.get_uniform_comparator("len_ge"), "length_greater_than_or_equals")
        self.assertEqual(utils.get_uniform_comparator("count_ge"), "length_greater_than_or_equals")
        self.assertEqual(utils.get_uniform_comparator("count_greater_than_or_equals"), "length_greater_than_or_equals")

        self.assertEqual(utils.get_uniform_comparator("len_lt"), "length_less_than")
        self.assertEqual(utils.get_uniform_comparator("count_lt"), "length_less_than")
        self.assertEqual(utils.get_uniform_comparator("count_less_than"), "length_less_than")

        self.assertEqual(utils.get_uniform_comparator("len_le"), "length_less_than_or_equals")
        self.assertEqual(utils.get_uniform_comparator("count_le"), "length_less_than_or_equals")
        self.assertEqual(utils.get_uniform_comparator("count_less_than_or_equals"), "length_less_than_or_equals")

    def current_validators(self):
        imported_module = utils.get_imported_module("httprunner.built_in")
        functions_mapping = utils.filter_module(imported_module, "function")

        functions_mapping["equals"](None, None)
        functions_mapping["equals"](1, 1)
        functions_mapping["equals"]("abc", "abc")
        with self.assertRaises(AssertionError):
            functions_mapping["equals"]("123", 123)

        functions_mapping["less_than"](1, 2)
        functions_mapping["less_than_or_equals"](2, 2)

        functions_mapping["greater_than"](2, 1)
        functions_mapping["greater_than_or_equals"](2, 2)

        functions_mapping["not_equals"](123, "123")

        functions_mapping["length_equals"]("123", 3)
        functions_mapping["length_greater_than"]("123", 2)
        functions_mapping["length_greater_than_or_equals"]("123", 3)

        functions_mapping["contains"]("123abc456", "3ab")
        functions_mapping["contains"](['1', '2'], "1")
        functions_mapping["contains"]({'a':1, 'b':2}, "a")
        functions_mapping["contained_by"]("3ab", "123abc456")

        functions_mapping["regex_match"]("123abc456", "^123\w+456$")
        with self.assertRaises(AssertionError):
            functions_mapping["regex_match"]("123abc456", "^12b.*456$")

        functions_mapping["startswith"]("abc123", "ab")
        functions_mapping["startswith"]("123abc", 12)
        functions_mapping["startswith"](12345, 123)

        functions_mapping["endswith"]("abc123", 23)
        functions_mapping["endswith"]("123abc", "abc")
        functions_mapping["endswith"](12345, 45)

        functions_mapping["type_match"](580509390, int)
        functions_mapping["type_match"](580509390, "int")
        functions_mapping["type_match"]([], list)
        functions_mapping["type_match"]([], "list")
        functions_mapping["type_match"]([1], "list")
        functions_mapping["type_match"]({}, "dict")
        functions_mapping["type_match"]({"a": 1}, "dict")

    def test_deep_update_dict(self):
        origin_dict = {'a': 1, 'b': {'c': 3, 'd': 4}, 'f': 6, 'h': 123}
        override_dict = {'a': 2, 'b': {'c': 33, 'e': 5}, 'g': 7, 'h': None}
        updated_dict = utils.deep_update_dict(origin_dict, override_dict)
        self.assertEqual(
            updated_dict,
            {'a': 2, 'b': {'c': 33, 'd': 4, 'e': 5}, 'f': 6, 'g': 7, 'h': 123}
        )

    def test_get_imported_module(self):
        imported_module = utils.get_imported_module("os")
        self.assertIn("walk", dir(imported_module))

    def test_filter_module_functions(self):
        imported_module = utils.get_imported_module("httprunner.utils")
        self.assertIn("is_py3", dir(imported_module))

        functions_dict = utils.filter_module(imported_module, "function")
        self.assertIn("filter_module", functions_dict)
        self.assertNotIn("is_py3", functions_dict)

    def test_get_imported_module_from_file(self):
        imported_module = utils.get_imported_module_from_file("tests/debugtalk.py")
        self.assertIn("gen_md5", dir(imported_module))

        functions_dict = utils.filter_module(imported_module, "function")
        self.assertIn("gen_md5", functions_dict)
        self.assertNotIn("urllib", functions_dict)

        with self.assertRaises(exception.FileNotFoundError):
            utils.get_imported_module_from_file("tests/debugtalk2.py")

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
        self.assertIn("accept", new_dict["request"]["headers"])
        self.assertIn("user-agent", new_dict["request"]["headers"])

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
