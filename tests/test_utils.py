import os
from ate import utils
from ate import exception
from tests.base import ApiServerUnittest

class TestUtils(ApiServerUnittest):

    def test_load_testcases_bad_filepath(self):
        testcase_file_path = os.path.join(os.getcwd(), 'tests/data/demo')
        self.assertEqual(utils.load_testcases(testcase_file_path), [])

    def test_load_json_testcases(self):
        testcase_file_path = os.path.join(
            os.getcwd(), 'tests/data/demo_testset_hardcode.json')
        testcases = utils.load_testcases(testcase_file_path)
        self.assertEqual(len(testcases), 3)
        testcase = testcases[0]["test"]
        self.assertIn('name', testcase)
        self.assertIn('request', testcase)
        self.assertIn('url', testcase['request'])
        self.assertIn('method', testcase['request'])

    def test_load_yaml_testcases(self):
        testcase_file_path = os.path.join(
            os.getcwd(), 'tests/data/demo_testset_hardcode.yml')
        testcases = utils.load_testcases(testcase_file_path)
        self.assertEqual(len(testcases), 3)
        testcase = testcases[0]["test"]
        self.assertIn('name', testcase)
        self.assertIn('request', testcase)
        self.assertIn('url', testcase['request'])
        self.assertIn('method', testcase['request'])

    def test_load_foler_files(self):
        folder = os.path.join(os.getcwd(), 'tests')
        file1 = os.path.join(os.getcwd(), 'tests', 'test_utils.py')
        file2 = os.path.join(os.getcwd(), 'tests', 'data', 'demo_binds.yml')

        files = utils.load_foler_files(folder, ["*.py"])
        self.assertIn(file1, files)
        self.assertNotIn(file2, files)

    def test_load_testcases_by_path_files(self):
        testsets_list = []

        # absolute file path
        path = os.path.join(
            os.getcwd(), 'tests/data/demo_testset_hardcode.json')
        testset_list = utils.load_testcases_by_path(path)
        self.assertEqual(len(testset_list), 1)
        self.assertIn("path", testset_list[0]["config"])
        self.assertEqual(testset_list[0]["config"]["path"], path)
        self.assertEqual(len(testset_list[0]["testcases"]), 3)
        testsets_list.extend(testset_list)

        # relative file path
        path = 'tests/data/demo_testset_hardcode.yml'
        testset_list = utils.load_testcases_by_path(path)
        self.assertEqual(len(testset_list), 1)
        self.assertIn("path", testset_list[0]["config"])
        self.assertIn(path, testset_list[0]["config"]["path"])
        self.assertEqual(len(testset_list[0]["testcases"]), 3)
        testsets_list.extend(testset_list)

        # list/set container with file(s)
        path = [
            os.path.join(os.getcwd(), 'tests/data/demo_testset_hardcode.json'),
            'tests/data/demo_testset_hardcode.yml'
        ]
        testset_list = utils.load_testcases_by_path(path)
        self.assertEqual(len(testset_list), 2)
        self.assertEqual(len(testset_list[0]["testcases"]), 3)
        self.assertEqual(len(testset_list[1]["testcases"]), 3)
        testsets_list.extend(testset_list)
        self.assertEqual(len(testsets_list), 4)

        for testset in testsets_list:
            for testcase in testset["testcases"]:
                self.assertIn('name', testcase)
                self.assertIn('request', testcase)
                self.assertIn('url', testcase['request'])
                self.assertIn('method', testcase['request'])

    def test_load_testcases_by_path_folder(self):
        # absolute folder path
        path = os.path.join(os.getcwd(), 'tests/data')
        testset_list_1 = utils.load_testcases_by_path(path)
        self.assertGreater(len(testset_list_1), 4)

        # relative folder path
        path = 'tests/data/'
        testset_list_2 = utils.load_testcases_by_path(path)
        self.assertEqual(len(testset_list_1), len(testset_list_2))

        # list/set container with file(s)
        path = [
            os.path.join(os.getcwd(), 'tests/data'),
            'tests/data/'
        ]
        testset_list_3 = utils.load_testcases_by_path(path)
        self.assertEqual(len(testset_list_3), 2 * len(testset_list_1))

    def test_load_testcases_by_path_not_exist(self):
        # absolute folder path
        path = os.path.join(os.getcwd(), 'tests/data_not_exist')
        testset_list_1 = utils.load_testcases_by_path(path)
        self.assertEqual(testset_list_1, [])

        # relative folder path
        path = 'tests/data_not_exist'
        testset_list_2 = utils.load_testcases_by_path(path)
        self.assertEqual(testset_list_2, [])

        # list/set container with file(s)
        path = [
            os.path.join(os.getcwd(), 'tests/data_not_exist'),
            'tests/data_not_exist/'
        ]
        testset_list_3 = utils.load_testcases_by_path(path)
        self.assertEqual(testset_list_3, [])

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
        imported_module = utils.get_imported_module("ate.utils")
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
