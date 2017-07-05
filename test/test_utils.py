import os
from ate import utils
from ate import exception
from test.base import ApiServerUnittest

class TestUtils(ApiServerUnittest):

    def test_load_testcases_bad_filepath(self):
        testcase_file_path = os.path.join(os.getcwd(), 'test/data/demo')
        with self.assertRaises(exception.ParamsError):
            utils.load_testcases(testcase_file_path)

    def test_load_json_testcases(self):
        testcase_file_path = os.path.join(
            os.getcwd(), 'test/data/simple_demo_no_auth.json')
        testcases = utils.load_testcases(testcase_file_path)
        self.assertEqual(len(testcases), 2)
        testcase = testcases[0]["test"]
        self.assertIn('name', testcase)
        self.assertIn('request', testcase)
        self.assertIn('url', testcase['request'])
        self.assertIn('method', testcase['request'])

    def test_load_yaml_testcases(self):
        testcase_file_path = os.path.join(
            os.getcwd(), 'test/data/simple_demo_no_auth.yml')
        testcases = utils.load_testcases(testcase_file_path)
        self.assertEqual(len(testcases), 2)
        testcase = testcases[0]["test"]
        self.assertIn('name', testcase)
        self.assertIn('request', testcase)
        self.assertIn('url', testcase['request'])
        self.assertIn('method', testcase['request'])

    def test_load_foler_files(self):
        folder = os.path.join(os.getcwd(), 'test')
        files = utils.load_foler_files(folder)
        file1 = os.path.join(os.getcwd(), 'test', 'test_utils.py')
        file2 = os.path.join(os.getcwd(), 'test', 'data', 'demo_binds.yml')
        self.assertIn(file1, files)
        self.assertIn(file2, files)

    def test_load_testcases_by_path_files(self):
        testsets_list = []

        # absolute file path
        path = os.path.join(
            os.getcwd(), 'test/data/simple_demo_no_auth.json')
        testset_list = utils.load_testcases_by_path(path)
        self.assertEqual(len(testset_list), 1)
        self.assertEqual(len(testset_list[0]["testcases"]), 2)
        testsets_list.extend(testset_list)

        # relative file path
        path = 'test/data/simple_demo_no_auth.yml'
        testset_list = utils.load_testcases_by_path(path)
        self.assertEqual(len(testset_list), 1)
        self.assertEqual(len(testset_list[0]["testcases"]), 2)
        testsets_list.extend(testset_list)

        # list/set container with file(s)
        path = [
            os.path.join(os.getcwd(), 'test/data/simple_demo_no_auth.json'),
            'test/data/simple_demo_no_auth.yml'
        ]
        testset_list = utils.load_testcases_by_path(path)
        self.assertEqual(len(testset_list), 2)
        self.assertEqual(len(testset_list[0]["testcases"]), 2)
        self.assertEqual(len(testset_list[1]["testcases"]), 2)
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
        path = os.path.join(os.getcwd(), 'test/data')
        testset_list_1 = utils.load_testcases_by_path(path)
        self.assertGreater(len(testset_list_1), 6)

        # relative folder path
        path = 'test/data/'
        testset_list_2 = utils.load_testcases_by_path(path)
        self.assertEqual(len(testset_list_1), len(testset_list_2))

        # list/set container with file(s)
        path = [
            os.path.join(os.getcwd(), 'test/data'),
            'test/data/'
        ]
        testset_list_3 = utils.load_testcases_by_path(path)
        self.assertEqual(len(testset_list_3), 2 * len(testset_list_1))

    def test_load_testcases_by_path_not_exist(self):
        # absolute folder path
        path = os.path.join(os.getcwd(), 'test/data_not_exist')
        testset_list_1 = utils.load_testcases_by_path(path)
        self.assertEqual(testset_list_1, [])

        # relative folder path
        path = 'test/data_not_exist'
        testset_list_2 = utils.load_testcases_by_path(path)
        self.assertEqual(testset_list_2, [])

        # list/set container with file(s)
        path = [
            os.path.join(os.getcwd(), 'test/data_not_exist'),
            'test/data_not_exist/'
        ]
        testset_list_3 = utils.load_testcases_by_path(path)
        self.assertEqual(testset_list_3, [])

    def test_is_variable(self):
        content = "$var"
        self.assertTrue(utils.is_variable(content))
        content = "$var123"
        self.assertTrue(utils.is_variable(content))
        content = "$var_name"
        self.assertTrue(utils.is_variable(content))
        content = "var"
        self.assertFalse(utils.is_variable(content))
        content = "a$var"
        self.assertFalse(utils.is_variable(content))
        content = "$v ar"
        self.assertFalse(utils.is_variable(content))
        content = " "
        self.assertFalse(utils.is_variable(content))
        content = "$abc*"
        self.assertFalse(utils.is_variable(content))

    def test_parse_variable(self):
        content = "$var"
        self.assertEqual(utils.parse_variable(content), "var")
        content = "$var123"
        self.assertEqual(utils.parse_variable(content), "var123")
        content = "$var_name"
        self.assertEqual(utils.parse_variable(content), "var_name")

    def test_is_functon(self):
        content = "${func()}"
        self.assertTrue(utils.is_functon(content))
        content = "${func(5)}"
        self.assertTrue(utils.is_functon(content))
        content = "${func(1, 2)}"
        self.assertTrue(utils.is_functon(content))
        content = "${func($a, $b)}"
        self.assertTrue(utils.is_functon(content))
        content = "${func(a=1, b=2)}"
        self.assertTrue(utils.is_functon(content))
        content = "${func(1, 2, a=3, b=4)}"
        self.assertTrue(utils.is_functon(content))
        content = "${func(1, $b, c=$x, d=4)}"
        self.assertTrue(utils.is_functon(content))
        content = "${func}"
        self.assertFalse(utils.is_functon(content))
        content = "$abc"
        self.assertFalse(utils.is_functon(content))
        content = "abc"
        self.assertFalse(utils.is_functon(content))

    def test_parse_string_value(self):
        str_value = "123"
        self.assertEqual(utils.parse_string_value(str_value), 123)
        str_value = "12.3"
        self.assertEqual(utils.parse_string_value(str_value), 12.3)
        str_value = "a123"
        self.assertEqual(utils.parse_string_value(str_value), "a123")
        str_value = "$var"
        self.assertEqual(utils.parse_string_value(str_value), "$var")
        str_value = "${func}"
        self.assertEqual(utils.parse_string_value(str_value), "${func}")

    def test_parse_functon(self):
        content = "${func()}"
        self.assertEqual(
            utils.parse_function(content),
            {'func_name': 'func', 'args': [], 'kwargs': {}}
        )
        content = "${func(5)}"
        self.assertEqual(
            utils.parse_function(content),
            {'func_name': 'func', 'args': [5], 'kwargs': {}}
        )
        content = "${func(1, 2)}"
        self.assertEqual(
            utils.parse_function(content),
            {'func_name': 'func', 'args': [1, 2], 'kwargs': {}}
        )
        content = "${func(a=1, b=2)}"
        self.assertEqual(
            utils.parse_function(content),
            {'func_name': 'func', 'args': [], 'kwargs': {'a': 1, 'b': 2}}
        )
        content = "${func(a= 1, b =2)}"
        self.assertEqual(
            utils.parse_function(content),
            {'func_name': 'func', 'args': [], 'kwargs': {'a': 1, 'b': 2}}
        )
        content = "${func(1, 2, a=3, b=4)}"
        self.assertEqual(
            utils.parse_function(content),
            {'func_name': 'func', 'args': [1, 2], 'kwargs': {'a': 3, 'b': 4}}
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
        with self.assertRaises(exception.ParamsError):
            utils.query_json(json_content, query)

        query = "ids.5"
        with self.assertRaises(exception.ParamsError):
            utils.query_json(json_content, query)

        query = "person.age"
        result = utils.query_json(json_content, query)
        self.assertEqual(result, 29)

        query = "person.not_exist_key"
        with self.assertRaises(exception.ParamsError):
            utils.query_json(json_content, query)

        query = "person.cities.0"
        result = utils.query_json(json_content, query)
        self.assertEqual(result, "Guangzhou")

        query = "person.name.first_name"
        result = utils.query_json(json_content, query)
        self.assertEqual(result, "Leo")

    def test_match_expected(self):
        self.assertTrue(utils.match_expected(1, 1, "eq"))
        self.assertTrue(utils.match_expected("abc", "abc", "eq"))
        self.assertTrue(utils.match_expected("abc", "abc"))
        self.assertFalse(utils.match_expected(123, "123", "eq"))
        self.assertFalse(utils.match_expected(123, "123"))

        self.assertTrue(utils.match_expected("123", 3, "len_eq"))
        self.assertTrue(utils.match_expected(123, "123", "str_eq"))
        self.assertTrue(utils.match_expected(123, "123", "ne"))

        self.assertTrue(utils.match_expected("123", 2, "len_gt"))
        self.assertTrue(utils.match_expected("123", 3, "len_ge"))
        self.assertTrue(utils.match_expected("123", 4, "len_lt"))
        self.assertTrue(utils.match_expected("123", 3, "len_le"))

        self.assertTrue(utils.match_expected(1, 2, "lt"))
        self.assertTrue(utils.match_expected(1, 1, "le"))
        self.assertTrue(utils.match_expected(2, 1, "gt"))
        self.assertTrue(utils.match_expected(1, 1, "ge"))

        self.assertTrue(utils.match_expected("123abc456", "3ab", "contains"))
        self.assertTrue(utils.match_expected("3ab", "123abc456", "contained_by"))

        self.assertTrue(utils.match_expected("123abc456", "^123.*456$", "regex"))
        self.assertFalse(utils.match_expected("123abc456", "^12b.*456$", "regex"))

        with self.assertRaises(exception.ParamsError):
            utils.match_expected(1, 2, "not_supported_comparator")

        self.assertTrue(utils.match_expected("2017-06-29 17:29:58", 19, "str_len"))
        self.assertTrue(utils.match_expected("2017-06-29 17:29:58", "19", "str_len"))

        self.assertTrue(utils.match_expected("abc123", "ab", "startswith"))
        self.assertTrue(utils.match_expected("123abc", 12, "startswith"))
        self.assertTrue(utils.match_expected(12345, 123, "startswith"))

    def test_deep_update_dict(self):
        origin_dict = {'a': 1, 'b': {'c': 3, 'd': 4}, 'f': 6}
        override_dict = {'a': 2, 'b': {'c': 33, 'e': 5}, 'g': 7}
        updated_dict = utils.deep_update_dict(origin_dict, override_dict)
        self.assertEqual(
            updated_dict,
            {'a': 2, 'b': {'c': 33, 'd': 4, 'e': 5}, 'f': 6, 'g': 7}
        )
