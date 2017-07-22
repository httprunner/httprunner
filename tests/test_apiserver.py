import random
import requests

from tests.base import ApiServerUnittest


class TestApiServer(ApiServerUnittest):

    def setUp(self):
        super(TestApiServer, self).setUp()
        self.headers = self.get_authenticated_headers()
        self.reset_all()

    def tearDown(self):
        super(TestApiServer, self).tearDown()

    def test_index(self):
        resp = self.api_client.get(self.host)
        self.assertEqual(200, resp.status_code)

    def reset_all(self):
        url = "%s/api/reset-all" % self.host
        return self.api_client.get(url, headers=self.headers)

    def get_users(self):
        url = "%s/api/users" % self.host
        return self.api_client.get(url, headers=self.headers)

    def create_user(self, uid, name, password):
        url = "%s/api/users/%d" % (self.host, uid)
        data = {
            'name': name,
            'password': password
        }
        return self.api_client.post(url, headers=self.headers, json=data)

    def get_user(self, uid):
        url = "%s/api/users/%d" % (self.host, uid)
        return self.api_client.get(url, headers=self.headers)

    def update_user(self, uid, name, password):
        url = "%s/api/users/%d" % (self.host, uid)
        data = {
            'name': name,
            'password': password
        }
        return self.api_client.put(url, headers=self.headers, json=data)

    def delete_user(self, uid):
        url = "%s/api/users/%d" % (self.host, uid)
        return self.api_client.delete(url, headers=self.headers)

    def test_reset_all(self):
        resp = self.reset_all()
        self.assertEqual(200, resp.status_code)
        self.assertEqual(True, resp.json()['success'])

    def test_create_user_not_existed(self):
        resp = self.create_user(1000, 'user1', '123456')
        self.assertEqual(201, resp.status_code)
        self.assertEqual(True, resp.json()['success'])

    def test_create_user_existed(self):
        resp = self.create_user(1000, 'user1', '123456')
        resp = self.create_user(1000, 'user1', '123456')
        self.assertEqual(500, resp.status_code)

    def test_get_users_empty(self):
        resp = self.get_users()
        self.assertEqual(200, resp.status_code)
        self.assertEqual(resp.json()['count'], 0)

    def test_get_users_not_empty(self):
        resp = self.create_user(1000, 'user1', '123456')
        resp = self.get_users()
        self.assertEqual(200, resp.status_code)
        self.assertEqual(resp.json()['count'], 1)

        resp = self.create_user(1001, 'user2', '123456')
        resp = self.get_users()
        self.assertEqual(200, resp.status_code)
        self.assertEqual(resp.json()['count'], 2)

    def test_get_user_not_existed(self):
        resp = self.get_user(1000)
        self.assertEqual(404, resp.status_code)
        self.assertEqual(resp.json()['success'], False)

    def test_get_user_existed(self):
        self.create_user(1000, 'user1', '123456')
        resp = self.get_user(1000)
        self.assertEqual(200, resp.status_code)
        self.assertEqual(resp.json()['success'], True)

    def test_update_user_not_existed(self):
        resp = self.update_user(1000, 'user1', '123456')
        self.assertEqual(404, resp.status_code)
        self.assertEqual(resp.json()['success'], False)

    def test_update_user_existed(self):
        self.create_user(1000, 'user1', '123456')
        resp = self.update_user(1000, 'user2', '123456')
        self.assertEqual(200, resp.status_code)
        self.assertEqual(resp.json()['data']['name'], 'user2')

    def test_delete_user_not_existed(self):
        resp = self.delete_user(1000)
        self.assertEqual(404, resp.status_code)
        self.assertEqual(resp.json()['success'], False)

    def test_delete_user_existed(self):
        self.create_user(1000, 'leo', '123456')
        resp = self.delete_user(1000)
        self.assertEqual(200, resp.status_code)
        self.assertEqual(resp.json()['success'], True)

    def test_get_customized_response_status_code(self):
        status_code = random.randint(200, 511)
        url = "%s/customize-response" % self.host
        expected_response = {
            'status_code': status_code,
        }
        resp = self.api_client.post(
            url,
            headers=self.headers,
            json=expected_response
        )
        self.assertEqual(status_code, resp.status_code)

    def test_get_customized_response_headers(self):
        expected_response = {
            'headers': {
                'abc': 123,
                'def': 456
            }
        }
        url = "%s/customize-response" % self.host
        resp = self.api_client.post(
            url,
            headers=self.headers,
            json=expected_response
        )
        self.assertIn('abc', resp.headers)
        self.assertIn('123', resp.headers['abc'])
