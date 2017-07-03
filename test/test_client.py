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

    def create_user(self, uid, name, password):
        url = "%s/api/users/%d" % (self.host, uid)
        data = {
            'name': name,
            'password': password
        }
        return self.api_client.post(url, json=data)

    def test_create_user_not_existed(self):
        resp = self.create_user(1000, 'user1', '123456')
        self.assertEqual(201, resp.status_code)
        self.assertEqual(True, resp.json()['success'])

    def test_create_user_existed(self):
        resp = self.create_user(1000, 'user1', '123456')
        resp = self.create_user(1000, 'user1', '123456')
        self.assertEqual(500, resp.status_code)
