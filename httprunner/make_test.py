import os
import unittest

from httprunner import loader
from httprunner.make import (
    main_make,
    convert_testcase_path,
    pytest_files_made_cache_mapping,
    make_config_chain_style,
    make_teststep_chain_style,
    pytest_files_run_set,
    ensure_file_abs_path_valid,
)


class TestMake(unittest.TestCase):
    def setUp(self) -> None:
        pytest_files_made_cache_mapping.clear()
        pytest_files_run_set.clear()
        loader.project_meta = None
        self.data_dir = os.path.join(os.getcwd(), "examples", "data")

    def test_make_testcase(self):
        path = ["examples/postman_echo/request_methods/request_with_variables.yml"]
        testcase_python_list = main_make(path)
        self.assertEqual(
            testcase_python_list[0],
            os.path.join(
                os.getcwd(),
                os.path.join(
                    "examples",
                    "postman_echo",
                    "request_methods",
                    "request_with_variables_test.py",
                ),
            ),
        )

    def test_make_testcase_with_ref(self):
        path = [
            "examples/postman_echo/request_methods/request_with_testcase_reference.yml"
        ]
        testcase_python_list = main_make(path)
        self.assertEqual(len(testcase_python_list), 1)
        self.assertIn(
            os.path.join(
                os.getcwd(),
                os.path.join(
                    "examples",
                    "postman_echo",
                    "request_methods",
                    "request_with_testcase_reference_test.py",
                ),
            ),
            testcase_python_list,
        )

        with open(
            os.path.join(
                "examples",
                "postman_echo",
                "request_methods",
                "request_with_testcase_reference_test.py",
            )
        ) as f:
            content = f.read()
            self.assertIn(
                """
from request_methods.request_with_functions_test import (
    TestCaseRequestWithFunctions as RequestWithFunctions,
)
""",
                content,
            )
            self.assertIn(
                ".call(RequestWithFunctions)",
                content,
            )

    def test_make_testcase_folder(self):
        path = ["examples/postman_echo/request_methods/"]
        testcase_python_list = main_make(path)
        self.assertIn(
            os.path.join(
                os.getcwd(),
                os.path.join(
                    "examples",
                    "postman_echo",
                    "request_methods",
                    "request_with_functions_test.py",
                ),
            ),
            testcase_python_list,
        )

    def test_ensure_file_path_valid(self):
        self.assertEqual(
            ensure_file_abs_path_valid(os.path.join(self.data_dir, "a-b.c", "2 3.yml")),
            os.path.join(self.data_dir, "a_b_c", "T2_3.yml"),
        )
        loader.project_meta = None
        self.assertEqual(
            ensure_file_abs_path_valid(
                os.path.join(os.getcwd(), "examples", "postman_echo", "request_methods")
            ),
            os.path.join(os.getcwd(), "examples", "postman_echo", "request_methods"),
        )
        loader.project_meta = None
        self.assertEqual(
            ensure_file_abs_path_valid(os.path.join(os.getcwd(), "pyproject.toml")),
            os.path.join(os.getcwd(), "pyproject.toml"),
        )
        loader.project_meta = None
        self.assertEqual(
            ensure_file_abs_path_valid(os.getcwd()),
            os.getcwd(),
        )
        loader.project_meta = None
        self.assertEqual(
            ensure_file_abs_path_valid(os.path.join(self.data_dir, ".csv")),
            os.path.join(self.data_dir, ".csv"),
        )

    def test_convert_testcase_path(self):
        self.assertEqual(
            convert_testcase_path(os.path.join(self.data_dir, "a-b.c", "2 3.yml")),
            (
                os.path.join(self.data_dir, "a_b_c", "T2_3_test.py"),
                "T23",
            ),
        )
        self.assertEqual(
            convert_testcase_path(os.path.join(self.data_dir, "a-b.c", "中文case.yml")),
            (
                os.path.join(self.data_dir, "a_b_c", "中文case_test.py"),
                "中文Case",
            ),
        )

    def test_make_config_chain_style(self):
        config = {
            "name": "request methods testcase: validate with functions",
            "variables": {"foo1": "bar1", "foo2": 22},
            "base_url": "https://postman_echo.com",
            "verify": False,
            "path": "examples/postman_echo/request_methods/validate_with_functions_test.py",
        }
        self.assertEqual(
            make_config_chain_style(config),
            """Config("request methods testcase: validate with functions").variables(**{'foo1': 'bar1', 'foo2': 22}).base_url("https://postman_echo.com").verify(False)""",
        )

    def test_make_teststep_chain_style(self):
        step = {
            "name": "get with params",
            "variables": {
                "foo1": "bar1",
                "foo2": 123,
                "sum_v": "${sum_two(1, 2)}",
            },
            "request": {
                "method": "GET",
                "url": "/get",
                "params": {"foo1": "$foo1", "foo2": "$foo2", "sum_v": "$sum_v"},
                "headers": {"User-Agent": "HttpRunner/${get_httprunner_version()}"},
            },
            "testcase": "CLS_LB(TestCaseDemo)CLS_RB",
            "extract": {
                "session_foo1": "body.args.foo1",
                "session_foo2": "body.args.foo2",
            },
            "validate": [
                {"eq": ["status_code", 200]},
                {"eq": ["body.args.sum_v", "3"]},
            ],
        }
        teststep_chain_style = make_teststep_chain_style(step)
        self.assertEqual(
            teststep_chain_style,
            """Step(RunRequest("get with params").with_variables(**{'foo1': 'bar1', 'foo2': 123, 'sum_v': '${sum_two(1, 2)}'}).get("/get").with_params(**{'foo1': '$foo1', 'foo2': '$foo2', 'sum_v': '$sum_v'}).with_headers(**{'User-Agent': 'HttpRunner/${get_httprunner_version()}'}).extract().with_jmespath('body.args.foo1', 'session_foo1').with_jmespath('body.args.foo2', 'session_foo2').validate().assert_equal("status_code", 200).assert_equal("body.args.sum_v", "3"))""",
        )

    def test_make_requests_with_json_chain_style(self):
        step = {
            "name": "get with params",
            "variables": {
                "foo1": "bar1",
                "foo2": 123,
                "sum_v": "${sum_two(1, 2)}",
                "myjson": {"name": "user", "password": "123456"},
            },
            "request": {
                "method": "GET",
                "url": "/get",
                "params": {"foo1": "$foo1", "foo2": "$foo2", "sum_v": "$sum_v"},
                "headers": {"User-Agent": "HttpRunner/${get_httprunner_version()}"},
                "json": "$myjson",
            },
            "testcase": "CLS_LB(TestCaseDemo)CLS_RB",
            "extract": {
                "session_foo1": "body.args.foo1",
                "session_foo2": "body.args.foo2",
            },
            "validate": [
                {"eq": ["status_code", 200]},
                {"eq": ["body.args.sum_v", "3"]},
            ],
        }
        teststep_chain_style = make_teststep_chain_style(step)
        self.assertEqual(
            teststep_chain_style,
            """Step(RunRequest("get with params").with_variables(**{'foo1': 'bar1', 'foo2': 123, 'sum_v': '${sum_two(1, 2)}', 'myjson': {'name': 'user', 'password': '123456'}}).get("/get").with_params(**{'foo1': '$foo1', 'foo2': '$foo2', 'sum_v': '$sum_v'}).with_headers(**{'User-Agent': 'HttpRunner/${get_httprunner_version()}'}).with_json("$myjson").extract().with_jmespath('body.args.foo1', 'session_foo1').with_jmespath('body.args.foo2', 'session_foo2').validate().assert_equal("status_code", 200).assert_equal("body.args.sum_v", "3"))""",
        )
