import os
import unittest

from httprunner import loader
from httprunner.cli import main_run
from httprunner.client import HttpSession
from httprunner.runner import HttpRunner


class TestHttpRunner(unittest.TestCase):
    def setUp(self):
        loader.project_meta = None
        self.runner = HttpRunner()

    def test_run_testcase_by_path_request_only(self):
        self.runner.run_path(
            "examples/postman_echo/request_methods/request_with_functions.yml"
        )
        result = self.runner.get_summary()
        self.assertTrue(result.success)
        self.assertEqual(result.name, "request methods testcase with functions")
        self.assertEqual(result.step_datas[0].name, "get with params")
        self.assertEqual(len(result.step_datas), 3)

    def test_run_testcase_by_path_ref_testcase(self):
        self.runner.run_path(
            "examples/postman_echo/request_methods/request_with_testcase_reference.yml"
        )
        result = self.runner.get_summary()
        self.assertTrue(result.success)
        self.assertEqual(result.name, "request methods testcase: reference testcase")
        self.assertEqual(result.step_datas[0].name, "request with functions")
        self.assertEqual(len(result.step_datas), 2)

    def test_run_testcase_with_abnormal_path(self):
        exit_code = main_run(["tests/data/a-b.c/2 3.yml"])
        self.assertEqual(exit_code, 0)
        self.assertTrue(os.path.exists("tests/data/a_b_c/__init__.py"))
        self.assertTrue(os.path.exists("tests/data/debugtalk.py"))
        self.assertTrue(os.path.exists("tests/data/a_b_c/T1_test.py"))
        self.assertTrue(os.path.exists("tests/data/a_b_c/T2_3_test.py"))


class TestHttpSession(unittest.TestCase):
    def setUp(self):
        self.session = HttpSession()

    def test_request_http(self):
        self.session.request("get", "http://httpbin.org/get")
        address = self.session.data.address
        self.assertGreater(len(address.server_ip), 0)
        self.assertEqual(address.server_port, 80)
        self.assertGreater(len(address.client_ip), 0)
        self.assertGreater(address.client_port, 10000)

    def test_request_https(self):
        self.session.request("get", "https://httpbin.org/get")
        address = self.session.data.address
        self.assertGreater(len(address.server_ip), 0)
        self.assertEqual(address.server_port, 443)
        self.assertGreater(len(address.client_ip), 0)
        self.assertGreater(address.client_port, 10000)

    def test_request_http_allow_redirects(self):
        self.session.request(
            "get",
            "http://httpbin.org/redirect-to?url=https%3A%2F%2Fgithub.com",
            allow_redirects=True)
        address = self.session.data.address
        self.assertNotEqual(address.server_ip, "N/A")
        self.assertEqual(address.server_port, 443)
        self.assertNotEqual(address.server_ip, "N/A")
        self.assertGreater(address.client_port, 10000)

    def test_request_https_allow_redirects(self):
        self.session.request(
            "get",
            "https://httpbin.org/redirect-to?url=https%3A%2F%2Fgithub.com",
            allow_redirects=True)
        address = self.session.data.address
        self.assertNotEqual(address.server_ip, "N/A")
        self.assertEqual(address.server_port, 443)
        self.assertNotEqual(address.server_ip, "N/A")
        self.assertGreater(address.client_port, 10000)

    def test_request_http_not_allow_redirects(self):
        self.session.request(
            "get",
            "http://httpbin.org/redirect-to?url=https%3A%2F%2Fgithub.com",
            allow_redirects=False)
        address = self.session.data.address
        self.assertEqual(address.server_ip, "N/A")
        self.assertEqual(address.server_port, 0)
        self.assertEqual(address.client_ip, "N/A")
        self.assertEqual(address.client_port, 0)

    def test_request_https_not_allow_redirects(self):
        self.session.request(
            "get",
            "https://httpbin.org/redirect-to?url=https%3A%2F%2Fgithub.com",
            allow_redirects=False)
        address = self.session.data.address
        self.assertEqual(address.server_ip, "N/A")
        self.assertEqual(address.server_port, 0)
        self.assertEqual(address.client_ip, "N/A")
        self.assertEqual(address.client_port, 0)
