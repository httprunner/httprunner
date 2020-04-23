import io
import os
import unittest

from httprunner import loader, utils


class TestUtils(unittest.TestCase):

    def test_set_os_environ(self):
        self.assertNotIn("abc", os.environ)
        variables_mapping = {
            "abc": "123"
        }
        utils.set_os_environ(variables_mapping)
        self.assertIn("abc", os.environ)
        self.assertEqual(os.environ["abc"], "123")

    def current_validators(self):
        from httprunner.builtin import comparators
        functions_mapping = loader.load.load_module_functions(comparators)

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
        # Because the Numbers in a CSV file are by default treated as strings, 
        # you need to convert them to Numbers, and we'll test that out here.
        functions_mapping["length_equals"]("123", '3')
        with self.assertRaises(AssertionError):
            functions_mapping["length_equals"]("123", 'abc')
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

    def test_deepcopy_dict(self):
        license_path = os.path.join(
            os.path.dirname(os.path.dirname(__file__)),
            "LICENSE"
        )
        data = {
            'a': 1,
            'b': [2, 4],
            'c': lambda x: x+1,
            'd': open(license_path),
            'f': {
                'f1': {'a1': 2},
                'f2': io.open(license_path, 'rb'),
            }
        }
        new_data = utils.deepcopy_dict(data)
        data["a"] = 0
        self.assertEqual(new_data["a"], 1)
        data["f"]["f1"] = 123
        self.assertEqual(new_data["f"]["f1"], {'a1': 2})
        self.assertNotEqual(id(new_data["b"]), id(data["b"]))
        self.assertEqual(id(new_data["c"]), id(data["c"]))
        # self.assertEqual(id(new_data["d"]), id(data["d"]))

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

    def test_print_info(self):
        info_mapping = {
            "a": 1,
            "t": (1, 2),
            "b": {
                "b1": 123
            },
            "c": None,
            "d": [4, 5]
        }
        utils.print_info(info_mapping)

    def test_prepare_dump_json_file_path_for_folder(self):
        # hrun tests/httpbin/a.b.c/ --save-tests
        test_path = os.path.join("tests", "httpbin", "a.b.c")
        self.assertEqual(
            utils.prepare_log_file_abs_path(test_path, "loaded.json"),
            os.path.join(os.getcwd(), "logs", "tests/httpbin/a.b.c/all.loaded.json")
        )

    def test_prepare_dump_json_file_path_for_file(self):
        # hrun tests/httpbin/a.b.c/rpc.yml --save-tests
        test_path = os.path.join("tests", "httpbin", "a.b.c", "rpc.yml")
        self.assertEqual(
            utils.prepare_log_file_abs_path(test_path, "loaded.json"),
            os.path.join(os.getcwd(), "logs", "tests/httpbin/a.b.c/rpc.loaded.json")
        )

    def test_prepare_dump_json_file_path_for_passed_testcase(self):
        test_path = ""
        self.assertEqual(
            utils.prepare_log_file_abs_path(test_path, "loaded.json"),
            os.path.join(os.getcwd(), "logs", "tests_mapping.loaded.json")
        )
