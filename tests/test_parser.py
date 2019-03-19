import os
import time
import unittest

from httprunner import exceptions, loader, parser


class TestParser(unittest.TestCase):

    def test_parse_string_value(self):
        self.assertEqual(parser.parse_string_value("123"), 123)
        self.assertEqual(parser.parse_string_value("12.3"), 12.3)
        self.assertEqual(parser.parse_string_value("a123"), "a123")
        self.assertEqual(parser.parse_string_value("$var"), "$var")
        self.assertEqual(parser.parse_string_value("${func}"), "${func}")

    def test_extract_variables(self):
        self.assertEqual(
            parser.extract_variables("$var"),
            ["var"]
        )
        self.assertEqual(
            parser.extract_variables("$var123"),
            ["var123"]
        )
        self.assertEqual(
            parser.extract_variables("$var_name"),
            ["var_name"]
        )
        self.assertEqual(
            parser.extract_variables("var"),
            []
        )
        self.assertEqual(
            parser.extract_variables("a$var"),
            ["var"]
        )
        self.assertEqual(
            parser.extract_variables("$v ar"),
            ["v"]
        )
        self.assertEqual(
            parser.extract_variables(" "),
            []
        )
        self.assertEqual(
            parser.extract_variables("$abc*"),
            ["abc"]
        )
        self.assertEqual(
            parser.extract_variables("${func()}"),
            []
        )
        self.assertEqual(
            parser.extract_variables("${func(1,2)}"),
            []
        )
        self.assertEqual(
            parser.extract_variables("${gen_md5($TOKEN, $data, $random)}"),
            ["TOKEN", "data", "random"]
        )

    def test_parse_function(self):
        self.assertEqual(
            parser.parse_function("func()"),
            {'func_name': 'func', 'args': [], 'kwargs': {}}
        )
        self.assertEqual(
            parser.parse_function("func(5)"),
            {'func_name': 'func', 'args': [5], 'kwargs': {}}
        )
        self.assertEqual(
            parser.parse_function("func(1, 2)"),
            {'func_name': 'func', 'args': [1, 2], 'kwargs': {}}
        )
        self.assertEqual(
            parser.parse_function("func(a=1, b=2)"),
            {'func_name': 'func', 'args': [], 'kwargs': {'a': 1, 'b': 2}}
        )
        self.assertEqual(
            parser.parse_function("func(a= 1, b =2)"),
            {'func_name': 'func', 'args': [], 'kwargs': {'a': 1, 'b': 2}}
        )
        self.assertEqual(
            parser.parse_function("func(1, 2, a=3, b=4)"),
            {'func_name': 'func', 'args': [1, 2], 'kwargs': {'a': 3, 'b': 4}}
        )
        self.assertEqual(
            parser.parse_function("func($request, 123)"),
            {'func_name': 'func', 'args': ["$request", 123], 'kwargs': {}}
        )
        self.assertEqual(
            parser.parse_function("func( )"),
            {'func_name': 'func', 'args': [], 'kwargs': {}}
        )
        self.assertEqual(
            parser.parse_function("func(hello world, a=3, b=4)"),
            {'func_name': 'func', 'args': ["hello world"], 'kwargs': {'a': 3, 'b': 4}}
        )
        self.assertEqual(
            parser.parse_function("func($request, 12 3)"),
            {'func_name': 'func', 'args': ["$request", '12 3'], 'kwargs': {}}
        )

    def test_parse_validator(self):
        validator = {"check": "status_code", "comparator": "eq", "expect": 201}
        self.assertEqual(
            parser.parse_validator(validator),
            {"check": "status_code", "comparator": "eq", "expect": 201}
        )

        validator = {'eq': ['status_code', 201]}
        self.assertEqual(
            parser.parse_validator(validator),
            {"check": "status_code", "comparator": "eq", "expect": 201}
        )

    def test_extract_functions(self):
        self.assertEqual(
            parser.extract_functions("${func()}"),
            ["func()"]
        )
        self.assertEqual(
            parser.extract_functions("${func(5)}"),
            ["func(5)"]
        )
        self.assertEqual(
            parser.extract_functions("${func(a=1, b=2)}"),
            ["func(a=1, b=2)"]
        )
        self.assertEqual(
            parser.extract_functions("${func(1, $b, c=$x, d=4)}"),
            ["func(1, $b, c=$x, d=4)"]
        )
        self.assertEqual(
            parser.extract_functions("/api/1000?_t=${get_timestamp()}"),
            ["get_timestamp()"]
        )
        self.assertEqual(
            parser.extract_functions("/api/${add(1, 2)}"),
            ["add(1, 2)"]
        )
        self.assertEqual(
            parser.extract_functions("/api/${add(1, 2)}?_t=${get_timestamp()}"),
            ["add(1, 2)", "get_timestamp()"]
        )
        self.assertEqual(
            parser.extract_functions("abc${func(1, 2, a=3, b=4)}def"),
            ["func(1, 2, a=3, b=4)"]
        )

    def test_parse_data(self):
        content = {
            'request': {
                'url': '/api/users/$uid',
                'method': "$method",
                'headers': {'token': '$token'},
                'data': {
                    "null": None,
                    "true": True,
                    "false": False,
                    "empty_str": "",
                    "value": "abc${add_one(3)}def"
                }
            }
        }
        variables_mapping = {
            "uid": 1000,
            "method": "POST",
            "token": "abc123"
        }
        functions_mapping = {
            "add_one": lambda x: x + 1
        }
        result = parser.parse_data(content, variables_mapping, functions_mapping)
        self.assertEqual("/api/users/1000", result["request"]["url"])
        self.assertEqual("abc123", result["request"]["headers"]["token"])
        self.assertEqual("POST", result["request"]["method"])
        self.assertIsNone(result["request"]["data"]["null"])
        self.assertTrue(result["request"]["data"]["true"])
        self.assertFalse(result["request"]["data"]["false"])
        self.assertEqual("", result["request"]["data"]["empty_str"])
        self.assertEqual("abc4def", result["request"]["data"]["value"])

    def test_parse_data_variables(self):
        variables_mapping = {
            "var_1": "abc",
            "var_2": "def",
            "var_3": 123,
            "var_4": {"a": 1},
            "var_5": True,
            "var_6": None
        }
        self.assertEqual(
            parser.parse_data("$var_1", variables_mapping),
            "abc"
        )
        self.assertEqual(
            parser.parse_data("var_1", variables_mapping),
            "var_1"
        )
        self.assertEqual(
            parser.parse_data("$var_1#XYZ", variables_mapping),
            "abc#XYZ"
        )
        self.assertEqual(
            parser.parse_data("/$var_1/$var_2/var3", variables_mapping),
            "/abc/def/var3"
        )
        self.assertEqual(
            parser.parse_data("/$var_1/$var_2/$var_1", variables_mapping),
            "/abc/def/abc"
        )
        self.assertEqual(
            parser.parse_string_variables("${func($var_1, $var_2, xyz)}", variables_mapping, {}),
            "${func(abc, def, xyz)}"
        )
        self.assertEqual(
            parser.parse_data("$var_3", variables_mapping),
            123
        )
        self.assertEqual(
            parser.parse_data("$var_4", variables_mapping),
            {"a": 1}
        )
        self.assertEqual(
            parser.parse_data("$var_5", variables_mapping),
            True
        )
        self.assertEqual(
            parser.parse_data("abc$var_5", variables_mapping),
            "abcTrue"
        )
        self.assertEqual(
            parser.parse_data("abc$var_4", variables_mapping),
            "abc{'a': 1}"
        )
        self.assertEqual(
            parser.parse_data("$var_6", variables_mapping),
            None
        )

        with self.assertRaises(exceptions.VariableNotFound):
            parser.parse_data("/api/$SECRET_KEY", variables_mapping)

        self.assertEqual(
            parser.parse_data(["$var_1", "$var_2"], variables_mapping),
            ["abc", "def"]
        )
        self.assertEqual(
            parser.parse_data({"$var_1": "$var_2"}, variables_mapping),
            {"abc": "def"}
        )

    def test_parse_data_multiple_identical_variables(self):
        variables_mapping = {
            "userid": 100,
            "data": 1498
        }
        content = "/users/$userid/training/$data?userId=$userid&data=$data"
        self.assertEqual(
            parser.parse_data(content, variables_mapping),
            "/users/100/training/1498?userId=100&data=1498"
        )

        variables_mapping = {
            "user": 100,
            "userid": 1000,
            "data": 1498
        }
        content = "/users/$user/$userid/$data?userId=$userid&data=$data"
        self.assertEqual(
            parser.parse_data(content, variables_mapping),
            "/users/100/1000/1498?userId=1000&data=1498"
        )

    def test_parse_data_functions(self):
        import random, string
        functions_mapping = {
            "gen_random_string": lambda str_len: ''.join(random.choice(string.ascii_letters + string.digits) \
                for _ in range(str_len))
        }
        result = parser.parse_data("${gen_random_string(5)}", functions_mapping=functions_mapping)
        self.assertEqual(len(result), 5)

        add_two_nums = lambda a, b=1: a + b
        functions_mapping["add_two_nums"] = add_two_nums
        self.assertEqual(
            parser.parse_data("${add_two_nums(1)}", functions_mapping=functions_mapping),
            2
        )
        self.assertEqual(
            parser.parse_data("${add_two_nums(1, 2)}", functions_mapping=functions_mapping),
            3
        )
        self.assertEqual(
            parser.parse_data("/api/${add_two_nums(1, 2)}", functions_mapping=functions_mapping),
            "/api/3"
        )

        with self.assertRaises(exceptions.FunctionNotFound):
            parser.parse_data("/api/${gen_md5(abc)}")

    def test_parse_data_testcase(self):
        variables = {
            "uid": "1000",
            "random": "A2dEx",
            "authorization": "a83de0ff8d2e896dbd8efb81ba14e17d",
            "data": {"name": "user", "password": "123456"}
        }
        functions = {
            "add_two_nums": lambda a, b=1: a + b,
            "get_timestamp": lambda: int(time.time() * 1000)
        }
        testcase_template = {
            "url": "http://127.0.0.1:5000/api/users/$uid/${add_two_nums(1,2)}",
            "method": "POST",
            "headers": {
                "Content-Type": "application/json",
                "authorization": "$authorization",
                "random": "$random",
                "sum": "${add_two_nums(1, 2)}"
            },
            "body": "$data"
        }
        parsed_testcase = parser.parse_data(testcase_template, variables, functions)
        self.assertEqual(
            parsed_testcase["url"],
            "http://127.0.0.1:5000/api/users/1000/3"
        )
        self.assertEqual(
            parsed_testcase["headers"]["authorization"],
            variables["authorization"]
        )
        self.assertEqual(
            parsed_testcase["headers"]["random"],
            variables["random"]
        )
        self.assertEqual(
            parsed_testcase["body"],
            variables["data"]
        )
        self.assertEqual(
            parsed_testcase["headers"]["sum"],
            3
        )

    def test_substitute_variables(self):
        content = {
            'request': {
                'url': '/api/users/$uid?id=$id',
                'headers': {'token': '$token'}
            }
        }
        variables_mapping = {"$uid": 1000, "$id": 2}
        substituted_data = parser.substitute_variables(content, variables_mapping)
        self.assertEqual(substituted_data["request"]["url"], "/api/users/1000?id=2")
        self.assertEqual(substituted_data["request"]["headers"], {'token': '$token'})

    def test_parse_parameters_raw_list(self):
        parameters = [
            {"user_agent": ["iOS/10.1", "iOS/10.2", "iOS/10.3"]},
            {"username-password": [("user1", "111111"), ["test2", "222222"]]}
        ]
        variables_mapping = {}
        functions_mapping = {}
        cartesian_product_parameters = parser.parse_parameters(parameters)
        self.assertEqual(
            len(cartesian_product_parameters),
            3 * 2
        )
        self.assertEqual(
            cartesian_product_parameters[0],
            {'user_agent': 'iOS/10.1', 'username': 'user1', 'password': '111111'}
        )

    def test_parse_parameters_custom_function(self):
        parameters = [
            {"user_agent": "${get_user_agent()}"},
            {"app_version": "${gen_app_version()}"},
            {"username-password": "${get_account()}"},
            {"username2-password2": "${get_account_in_tuple()}"}
        ]
        dot_env_path = os.path.join(
            os.getcwd(), "tests", ".env"
        )
        loader.load_dot_env_file(dot_env_path)
        from tests import debugtalk
        cartesian_product_parameters = parser.parse_parameters(
            parameters,
            functions_mapping=loader.load_module_functions(debugtalk)
        )
        self.assertIn(
            {
                'user_agent': 'iOS/10.1',
                'app_version': '2.8.5',
                'username': 'user1',
                'password': '111111',
                'username2': 'user1',
                'password2': '111111'
            },
            cartesian_product_parameters
        )
        self.assertEqual(
            len(cartesian_product_parameters),
            2 * 2 * 2 * 2
        )

    def test_parse_parameters_parameterize(self):
        loader.load_project_tests(os.path.join(os.getcwd(), "tests"))
        parameters = [
            {"app_version": "${parameterize(data/app_version.csv)}"},
            {"username-password": "${parameterize(data/account.csv)}"}
        ]
        cartesian_product_parameters = parser.parse_parameters(parameters)
        self.assertEqual(
            len(cartesian_product_parameters),
            2 * 3
        )

    def test_parse_parameters_mix(self):
        loader.load_project_tests(os.path.join(os.getcwd(), "tests"))
        project_mapping = loader.project_mapping

        parameters = [
            {"user_agent": ["iOS/10.1", "iOS/10.2", "iOS/10.3"]},
            {"app_version": "${gen_app_version()}"},
            {"username-password": "${parameterize(data/account.csv)}"}
        ]
        cartesian_product_parameters = parser.parse_parameters(
            parameters, functions_mapping=project_mapping["functions"])
        self.assertEqual(
            len(cartesian_product_parameters),
            3 * 2 * 3
        )

    def test_parse_tests_testcase(self):
        testcase_file_path = os.path.join(
            os.getcwd(), 'tests/data/demo_testcase.yml')
        tests_mapping = loader.load_tests(testcase_file_path)
        testcases = tests_mapping["testcases"]
        self.assertEqual(
            testcases[0]["config"]["variables"]["var_c"],
            "${sum_two(1, 2)}"
        )
        self.assertEqual(
            testcases[0]["config"]["variables"]["PROJECT_KEY"],
            "${ENV(PROJECT_KEY)}"
        )
        parsed_tests_mapping = parser.parse_tests(tests_mapping)
        parsed_testcases = parsed_tests_mapping["testcases"]
        self.assertIsInstance(parsed_testcases, list)
        test_dict1 = parsed_testcases[0]["teststeps"][0]
        self.assertEqual(test_dict1["variables"]["var_c"], 3)
        self.assertEqual(test_dict1["variables"]["PROJECT_KEY"], "ABCDEFGH")
        self.assertEqual(test_dict1["variables"]["var_d"], test_dict1["variables"]["var_e"])
        self.assertEqual(parsed_testcases[0]["config"]["name"], '1230')

    def test_parse_tests_override_variables(self):
        tests_mapping = {
            'testcases': [
                {
                    "config": {
                        'name': '',
                        'variables': [
                            {"password": "123456"},
                            {"creator": "user_test_001"}
                        ]
                    },
                    "teststeps": [
                        {
                            'name': 'testcase1',
                            "variables": [
                                {"creator": "user_test_002"},
                                {"username": "$creator"}
                            ],
                            'request': {'url': '/api1', 'method': 'GET'}
                        }
                    ]
                }
            ]
        }
        parsed_tests_mapping = parser.parse_tests(tests_mapping)
        test_dict1_variables = parsed_tests_mapping["testcases"][0]["teststeps"][0]["variables"]
        self.assertEqual(test_dict1_variables["creator"], "user_test_001")
        self.assertEqual(test_dict1_variables["username"], "user_test_001")

    def test_parse_tests_base_url_priority(self):
        """ base_url & verify: priority test_dict > config
        """
        tests_mapping = {
            'testcases': [
                {
                    "config": {
                        'name': '',
                        "base_url": "$host",
                        'variables': {
                            "host": "https://debugtalk.com"
                        },
                        "verify": False
                    },
                    "teststeps": [
                        {
                            'name': 'testcase1',
                            "base_url": "https://httprunner.org",
                            'request': {'url': '/api1', 'method': 'GET', "verify": True}
                        }
                    ]
                }
            ]
        }
        parsed_tests_mapping = parser.parse_tests(tests_mapping)
        test_dict = parsed_tests_mapping["testcases"][0]["teststeps"][0]
        self.assertEqual(test_dict["request"]["url"], "https://httprunner.org/api1")
        self.assertEqual(test_dict["request"]["verify"], True)

    def test_parse_tests_base_url_path_with_variable(self):
        tests_mapping = {
            'testcases': [
                {
                    "config": {
                        'name': '',
                        "base_url": "$host1",
                        'variables': {
                            "host1": "https://debugtalk.com"
                        }
                    },
                    "teststeps": [
                        {
                            'name': 'testcase1',
                            "variables": {
                                "host2": "https://httprunner.org"
                            },
                            'request': {'url': '$host2/api1', 'method': 'GET'}
                        }
                    ]
                }
            ]
        }
        parsed_tests_mapping = parser.parse_tests(tests_mapping)
        test_dict = parsed_tests_mapping["testcases"][0]["teststeps"][0]
        self.assertEqual(test_dict["request"]["url"], "https://httprunner.org/api1")

    def test_parse_tests_base_url_test_dict(self):
        tests_mapping = {
            'testcases': [
                {
                    "config": {
                        'name': '',
                        "base_url": "$host1",
                        'variables': {
                            "host1": "https://debugtalk.com"
                        }
                    },
                    "teststeps": [
                        {
                            'name': 'testcase1',
                            "base_url": "$host2",
                            "variables": {
                                "host2": "https://httprunner.org"
                            },
                            'request': {'url': '/api1', 'method': 'GET'}
                        }
                    ]
                }
            ]
        }
        parsed_tests_mapping = parser.parse_tests(tests_mapping)
        test_dict = parsed_tests_mapping["testcases"][0]["teststeps"][0]
        self.assertEqual(test_dict["request"]["url"], "https://httprunner.org/api1")

    def test_parse_data_with_variables(self):
        variables = {
            "host2": "https://httprunner.org",
            "num3": "${sum_two($num2, 4)}",
            "num2": "${sum_two($num1, 3)}",
            "num1": "${sum_two(1, 2)}"
        }
        from tests.debugtalk import sum_two
        functions = {
            "sum_two": sum_two
        }
        parsed_testcase = parser.parse_data(variables, variables, functions)
        self.assertEqual(parsed_testcase["num3"], 10)
        self.assertEqual(parsed_testcase["num2"], 6)
        self.assertEqual(parsed_testcase["num1"], 3)

    def test_parse_data_with_variables_not_found(self):
        variables = {
            "host": "https://httprunner.org",
            "num4": "${sum_two($num0, 5)}",
            "num3": "${sum_two($num2, 4)}",
            "num2": "${sum_two($num1, 3)}",
            "num1": "${sum_two(1, 2)}"
        }
        from tests.debugtalk import sum_two
        functions = {
            "sum_two": sum_two
        }
        with self.assertRaises(exceptions.VariableNotFound):
            parser.parse_data(variables, variables, functions)

        parsed_testcase = parser.parse_data(
            variables,
            variables,
            functions,
            raise_if_variable_not_found=False
        )
        self.assertEqual(parsed_testcase["num3"], 10)
        self.assertEqual(parsed_testcase["num2"], 6)
        self.assertEqual(parsed_testcase["num1"], 3)
        self.assertEqual(parsed_testcase["num4"], "${sum_two($num0, 5)}")

    def test_parse_tests_variable_with_function(self):
        from tests.debugtalk import sum_two, gen_random_string
        tests_mapping = {
            "project_mapping": {
                "functions": {
                    "sum_two": sum_two,
                    "gen_random_string": gen_random_string
                }
            },
            'testcases': [
                {
                    "config": {
                        'name': '',
                        "base_url": "$host1",
                        'variables': {
                            "host1": "https://debugtalk.com",
                            "var_a": "${gen_random_string(5)}",
                            "var_b": "$var_a"
                        }
                    },
                    "teststeps": [
                        {
                            'name': 'testcase1',
                            "base_url": "$host2",
                            "variables": {
                                "host2": "https://httprunner.org",
                                "num3": "${sum_two($num2, 4)}",
                                "num2": "${sum_two($num1, 3)}",
                                "num1": "${sum_two(1, 2)}",
                                "str1": "${gen_random_string(5)}",
                                "str2": "$str1"
                            },
                            'request': {
                                'url': '/api1/?num1=$num1&num2=$num2&num3=$num3',
                                'method': 'GET'
                            }
                        }
                    ]
                }
            ]
        }
        parsed_tests_mapping = parser.parse_tests(tests_mapping)
        test_dict = parsed_tests_mapping["testcases"][0]["teststeps"][0]
        self.assertEqual(test_dict["variables"]["num3"], 10)
        self.assertEqual(test_dict["variables"]["num2"], 6)
        self.assertEqual(test_dict["variables"]["str1"], test_dict["variables"]["str2"])
        self.assertEqual(
            test_dict["request"]["url"],
            "https://httprunner.org/api1/?num1=3&num2=6&num3=10"
        )

    def test_parse_tests_variable_not_found(self):
        from tests.debugtalk import sum_two
        tests_mapping = {
            "project_mapping": {
                "functions": {
                    "sum_two": sum_two
                }
            },
            'testcases': [
                {
                    "config": {
                        'name': '',
                        "base_url": "$host1",
                        'variables': {
                            "host1": "https://debugtalk.com"
                        }
                    },
                    "teststeps": [
                        {
                            'name': 'testcase1',
                            "base_url": "$host2",
                            "variables": {
                                "host2": "https://httprunner.org",
                                "num4": "${sum_two($num0, 5)}",
                                "num3": "${sum_two($num2, 4)}",
                                "num2": "${sum_two($num1, 3)}",
                                "num1": "${sum_two(1, 2)}"
                            },
                            'request': {
                                'url': '/api1/?num1=$num1&num2=$num2&num3=$num3&num4=$num4',
                                'method': 'GET'
                            }
                        }
                    ]
                }
            ]
        }
        parsed_tests_mapping = parser.parse_tests(tests_mapping)
        test_dict = parsed_tests_mapping["testcases"][0]["teststeps"][0]
        self.assertEqual(test_dict["variables"]["num3"], 10)
        self.assertEqual(test_dict["variables"]["num2"], 6)
        self.assertEqual(test_dict["variables"]["num4"], "${sum_two($num0, 5)}")
        self.assertEqual(
            test_dict["request"]["url"],
            "https://httprunner.org/api1/?num1=3&num2=6&num3=10&num4=${sum_two($num0, 5)}"
        )

    def test_parse_tests_base_url_teststep_empty(self):
        """ base_url & verify: priority test_dict > config
        """
        tests_mapping = {
            'testcases': [
                {
                    "config": {
                        'name': '',
                        "base_url": "$host",
                        'variables': {
                            "host": "https://debugtalk.com"
                        },
                        "verify": False
                    },
                    "teststeps": [
                        {
                            'name': 'testcase1',
                            "base_url": "",
                            'request': {'url': '/api1', 'method': 'GET', "verify": True}
                        }
                    ]
                }
            ]
        }
        parsed_tests_mapping = parser.parse_tests(tests_mapping)
        test_dict = parsed_tests_mapping["testcases"][0]["teststeps"][0]
        self.assertEqual(test_dict["request"]["url"], "https://debugtalk.com/api1")
        self.assertEqual(test_dict["request"]["verify"], True)

    def test_parse_tests_verify_config_set(self):
        """ verify priority: test_dict > config
        """
        tests_mapping = {
            'testcases': [
                {
                    "config": {
                        'name': 'bugfix verify',
                        "base_url": "https://httpbin.org/",
                        "verify": False
                    },
                    "teststeps": [
                        {
                            'name': 'testcase1',
                            'request': {'url': '/headers', 'method': 'GET'}
                        }
                    ]
                }
            ]
        }
        parsed_tests_mapping = parser.parse_tests(tests_mapping)
        test_dict = parsed_tests_mapping["testcases"][0]["teststeps"][0]
        self.assertEqual(test_dict["request"]["verify"], False)

    def test_parse_tests_verify_config_unset(self):
        """ verify priority: test_dict > config
        """
        tests_mapping = {
            'testcases': [
                {
                    "config": {
                        'name': 'bugfix verify',
                        "base_url": "https://httpbin.org/",
                    },
                    "teststeps": [
                        {
                            'name': 'testcase1',
                            'request': {'url': '/headers', 'method': 'GET'}
                        }
                    ]
                }
            ]
        }
        parsed_tests_mapping = parser.parse_tests(tests_mapping)
        test_dict = parsed_tests_mapping["testcases"][0]["teststeps"][0]
        self.assertEqual(test_dict["request"]["verify"], True)

    def test_parse_tests_verify_step_set_false(self):
        """ verify priority: test_dict > config
        """
        tests_mapping = {
            'testcases': [
                {
                    "config": {
                        'name': 'bugfix verify',
                        "base_url": "https://httpbin.org/",
                        "verify": True
                    },
                    "teststeps": [
                        {
                            'name': 'testcase1',
                            'request': {'url': '/headers', 'method': 'GET', "verify": False}
                        }
                    ]
                }
            ]
        }
        parsed_tests_mapping = parser.parse_tests(tests_mapping)
        test_dict = parsed_tests_mapping["testcases"][0]["teststeps"][0]
        self.assertEqual(test_dict["request"]["verify"], False)

    def test_parse_tests_verify_nested_testcase_unset(self):
        tests_mapping = {
            'testcases': [
                {
                    'config': {
                        'name': 'inquiry price',
                        'verify': False
                    },
                    'teststeps': [
                        {
                            'name': 'login system',
                            'testcase': 'testcases/deps/login.yml',
                            'testcase_def': {
                                'config': {
                                    'name': 'login system'
                                },
                                'teststeps': [
                                    {
                                        'name': '/',
                                        'request': {
                                            'method': 'GET',
                                            'url': 'https://httpbin.org/'
                                        }
                                    }
                                ]
                            }
                        }
                    ]
                }
            ]
        }
        parsed_tests_mapping = parser.parse_tests(tests_mapping)
        test_dict = parsed_tests_mapping["testcases"][0]["teststeps"][0]
        self.assertEqual(test_dict["teststeps"][0]["request"]["verify"], False)

    def test_parse_environ(self):
        os.environ["PROJECT_KEY"] = "ABCDEFGH"
        content = {
            "variables": [
                {"PROJECT_KEY": "${ENV(PROJECT_KEY)}"}
            ]
        }
        result = parser.parse_data(content)

        content = {
            "variables": [
                {"PROJECT_KEY": "${ENV(PROJECT_KEY, abc)}"}
            ]
        }
        with self.assertRaises(exceptions.ParamsError):
            parser.parse_data(content)

        content = {
            "variables": [
                {"PROJECT_KEY": "${ENV(abc=123)}"}
            ]
        }
        with self.assertRaises(exceptions.ParamsError):
            parser.parse_data(content)

    def test_extend_with_api(self):
        loader.load_project_tests(os.path.join(os.getcwd(), "tests"))
        raw_testinfo = {
            "name": "get token",
            "base_url": "https://debugtalk.com",
            "api": "api/get_token.yml",
        }
        api_def_dict = loader.load_teststep(raw_testinfo)
        test_block = {
            "name": "override block",
            "times": 3,
            "variables": [
                {"var": 123}
            ],
            "base_url": "https://httprunner.org",
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

        extended_block = parser._extend_with_api(test_block, api_def_dict)
        self.assertEqual(extended_block["base_url"], "https://debugtalk.com")
        self.assertEqual(extended_block["name"], "override block")
        self.assertEqual({'var': 123}, extended_block["variables"])
        self.assertIn({'check': 'status_code', 'expect': 201, 'comparator': 'eq'}, extended_block["validate"])
        self.assertIn({'check': 'content.token', 'comparator': 'len_eq', 'expect': 32}, extended_block["validate"])
        self.assertEqual(extended_block["times"], 3)
