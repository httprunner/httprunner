import unittest

from httprunner.client import HttpSession


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
            allow_redirects=True,
        )
        address = self.session.data.address
        self.assertNotEqual(address.server_ip, "N/A")
        self.assertEqual(address.server_port, 443)
        self.assertNotEqual(address.server_ip, "N/A")
        self.assertGreater(address.client_port, 10000)

    def test_request_https_allow_redirects(self):
        self.session.request(
            "get",
            "https://httpbin.org/redirect-to?url=https%3A%2F%2Fgithub.com",
            allow_redirects=True,
        )
        address = self.session.data.address
        self.assertNotEqual(address.server_ip, "N/A")
        self.assertEqual(address.server_port, 443)
        self.assertNotEqual(address.server_ip, "N/A")
        self.assertGreater(address.client_port, 10000)

    def test_request_http_not_allow_redirects(self):
        self.session.request(
            "get",
            "http://httpbin.org/redirect-to?url=https%3A%2F%2Fgithub.com",
            allow_redirects=False,
        )
        address = self.session.data.address
        self.assertEqual(address.server_ip, "N/A")
        self.assertEqual(address.server_port, 0)
        self.assertEqual(address.client_ip, "N/A")
        self.assertEqual(address.client_port, 0)

    def test_request_https_not_allow_redirects(self):
        self.session.request(
            "get",
            "https://httpbin.org/redirect-to?url=https%3A%2F%2Fgithub.com",
            allow_redirects=False,
        )
        address = self.session.data.address
        self.assertEqual(address.server_ip, "N/A")
        self.assertEqual(address.server_port, 0)
        self.assertEqual(address.client_ip, "N/A")
        self.assertEqual(address.client_port, 0)
