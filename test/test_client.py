from ate.client import HttpSession
from test.base import ApiServerUnittest

class TestHttpClient(ApiServerUnittest):
    def setUp(self):
        super(TestHttpClient, self).setUp()
        self.host = "http://127.0.0.1:5000"
        self.api_client = HttpSession(self.host)
        self.clear_users()

    def tearDown(self):
        super(TestHttpClient, self).tearDown()

    def clear_users(self):
        url = "%s/api/users" % self.host
        return self.api_client.delete(url)

    def test_request_with_full_url(self):
        url = "%s/api/users/1000" % self.host
        data = {
            'name': 'user1',
            'password': '123456'
        }
        resp = self.api_client.post(url, json=data)
        self.assertEqual(201, resp.status_code)
        self.assertEqual(True, resp.json()['success'])

    def test_request_without_base_url(self):
        url = "/api/users/1000"
        data = {
            'name': 'user1',
            'password': '123456'
        }
        resp = self.api_client.post(url, json=data)
        self.assertEqual(201, resp.status_code)
        self.assertEqual(True, resp.json()['success'])
