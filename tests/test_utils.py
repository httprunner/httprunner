import os
import shutil

from httprunner import exceptions, loader, utils
from httprunner.compat import OrderedDict
from tests.base import ApiServerUnittest


class TestUtils(ApiServerUnittest):

    def test_remove_prefix(self):
        full_url = "http://debugtalk.com/post/123"
        prefix = "http://debugtalk.com"
        self.assertEqual(
            utils.remove_prefix(full_url, prefix),
            "/post/123"
        )

    def test_set_os_environ(self):
        self.assertNotIn("abc", os.environ)
        variables_mapping = {
            "abc": "123"
        }
        utils.set_os_environ(variables_mapping)
        self.assertIn("abc", os.environ)
        self.assertEqual(os.environ["abc"], "123")

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
        with self.assertRaises(exceptions.ExtractFailure):
            utils.query_json(json_content, query)

        query = "ids.5"
        with self.assertRaises(exceptions.ExtractFailure):
            utils.query_json(json_content, query)

        query = "person.age"
        result = utils.query_json(json_content, query)
        self.assertEqual(result, 29)

        query = "person.not_exist_key"
        with self.assertRaises(exceptions.ExtractFailure):
            utils.query_json(json_content, query)

        query = "person.cities.0"
        result = utils.query_json(json_content, query)
        self.assertEqual(result, "Guangzhou")

        query = "person.name.first_name"
        result = utils.query_json(json_content, query)
        self.assertEqual(result, "Leo")

        query = "person.name.first_name.0"
        result = utils.query_json(json_content, query)
        self.assertEqual(result, "L")

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
        from httprunner import built_in
        module_mapping = loader.load_python_module(built_in)
        functions_mapping = module_mapping["functions"]

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
        with self.assertRaises(exceptions.ParamsError):
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

    def test_cartesian_product_one(self):
        parameters_content_list = [
            [
                {"a": 1},
                {"a": 2}
            ]
        ]
        product_list = utils.gen_cartesian_product(*parameters_content_list)
        self.assertEqual(
            product_list,
            [
                {"a": 1},
                {"a": 2}
            ]
        )

    def test_cartesian_product_multiple(self):
        parameters_content_list = [
            [
                {"a": 1},
                {"a": 2}
            ],
            [
                {"x": 111, "y": 112},
                {"x": 121, "y": 122}
            ]
        ]
        product_list = utils.gen_cartesian_product(*parameters_content_list)
        self.assertEqual(
            product_list,
            [
                {'a': 1, 'x': 111, 'y': 112},
                {'a': 1, 'x': 121, 'y': 122},
                {'a': 2, 'x': 111, 'y': 112},
                {'a': 2, 'x': 121, 'y': 122}
            ]
        )

    def test_cartesian_product_empty(self):
        parameters_content_list = []
        product_list = utils.gen_cartesian_product(*parameters_content_list)
        self.assertEqual(product_list, [])
