from httprunner.client import HttpSession, prepare_kwargs
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

    def test_prepare_kwargs(self):
        kwargs = {
            "headers": {
                "content-type": "application/json; charset=utf-8"
            },
            "data": {
                "a": 1,
                "b": 2
            }
        }
        prepare_kwargs("POST", kwargs)
        self.assertIn('"a": 1', kwargs["data"])
        self.assertIn('"b": 2', kwargs["data"])
