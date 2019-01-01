from httprunner.client import HttpSession
from httprunner.compat import bytes
from tests.api_server import HTTPBIN_SERVER
from tests.base import ApiServerUnittest


class TestHttpClient(ApiServerUnittest):
    def setUp(self):
        super(TestHttpClient, self).setUp()
        self.api_client = HttpSession(self.host)
        self.headers = self.get_authenticated_headers()
        self.reset_all()

    def tearDown(self):
        super(TestHttpClient, self).tearDown()

    def reset_all(self):
        url = "%s/api/reset-all" % self.host
        headers = self.get_authenticated_headers()
        return self.api_client.get(url, headers=headers)

    def test_request_with_full_url(self):
        url = "%s/api/users/1000" % self.host
        data = {
            'name': 'user1',
            'password': '123456'
        }
        resp = self.api_client.post(url, json=data, headers=self.headers)
        self.assertEqual(201, resp.status_code)
        self.assertEqual(True, resp.json()['success'])

    def test_request_without_base_url(self):
        url = "/api/users/1000"
        data = {
            'name': 'user1',
            'password': '123456'
        }
        resp = self.api_client.post(url, json=data, headers=self.headers)
        self.assertEqual(201, resp.status_code)
        self.assertEqual(True, resp.json()['success'])

    def test_request_post_data(self):
        url = "/api/users/1000"
        data = {
            'name': 'user1',
            'password': '123456'
        }
        resp = self.api_client.post(url, json=data, headers=self.headers)
        # b'{"name": "user1", "password": "123456"}'
        self.assertIn(b'"name": "user1"', resp.request.body)
        self.assertIn(b'"password": "123456"', resp.request.body)
        resp = self.api_client.post(url, data=data, headers=self.headers)
        # name=user1&password=123456
        self.assertIn("name=user1", resp.request.body)
        self.assertIn("&", resp.request.body)
        self.assertIn("password=123456", resp.request.body)

    def test_request_with_cookies(self):
        url = "/api/users/1000"
        data = {
            'name': 'user1',
            'password': '123456'
        }
        cookies = {
            "a": "1",
            "b": "2"
        }
        resp = self.api_client.get(url, cookies=cookies, headers=self.headers)
        self.assertEqual(resp.request._cookies["a"], "1")
        self.assertEqual(resp.request._cookies["b"], "2")

    def test_request_redirect(self):
        url = "{}/redirect-to?url=https%3A%2F%2Fdebugtalk.com&status_code=302".format(HTTPBIN_SERVER)
        headers = {"accept: text/html"}
        cookies = {
            "a": "1",
            "b": "2"
        }
        resp = self.api_client.get(url, cookies=cookies, headers=self.headers)
        raw_request = resp.history[0].request
        self.assertEqual(raw_request._cookies["a"], "1")
        self.assertEqual(raw_request._cookies["b"], "2")
        redirect_request = resp.request
        self.assertEqual(redirect_request.url, "https://debugtalk.com")
        self.assertEqual(redirect_request._cookies["a"], "1")
        self.assertEqual(redirect_request._cookies["b"], "2")
