import gevent
import gevent.pywsgi
import requests
import unittest
from . import api_server

class TestApiServer(unittest.TestCase):
    """
    Test case class that sets up an HTTP server which can be used within the tests
    """
    def setUp(self):
        super(TestApiServer, self).setUp()
        self._api_server = gevent.pywsgi.WSGIServer(("127.0.0.1", 0), api_server.app, log=None)
        gevent.spawn(lambda: self._api_server.serve_forever())
        gevent.sleep(0.01)
        self.host = "http://127.0.0.1:%i" % self._api_server.server_port
        self.api_client = requests.Session()

    def tearDown(self):
        super(TestApiServer, self).tearDown()
        self._api_server.stop_accepting()
        self._api_server.stop()

    def clear_users(self):
        url = "%s/api/users" % self.host
        return self.api_client.delete(url)

    def get_users(self):
        url = "%s/api/users" % self.host
        return self.api_client.get(url)

    def create_user(self, uid, name, password):
        url = "%s/api/users/%d" % (self.host, uid)
        data = {
            'name': name,
            'password': password
        }
        return self.api_client.post(url, json=data)

    def get_user(self, uid):
        url = "%s/api/users/%d" % (self.host, uid)
        return self.api_client.get(url)

    def update_user(self, uid, name, password):
        url = "%s/api/users/%d" % (self.host, uid)
        data = {
            'name': name,
            'password': password
        }
        return self.api_client.put(url, json=data)

    def delete_user(self, uid):
        url = "%s/api/users/%d" % (self.host, uid)
        return self.api_client.delete(url)

    def test_clear_users(self):
        resp = self.clear_users()
        self.assertEqual(200, resp.status_code)
        self.assertEqual(True, resp.json()['success'])

    def test_create_user_not_existed(self):
        self.clear_users()
        resp = self.create_user(1000, 'user1', '123456')
        self.assertEqual(201, resp.status_code)

    def test_create_user_existed(self):
        self.clear_users()
        resp = self.create_user(1000, 'user1', '123456')
        resp = self.create_user(1000, 'user1', '123456')
        self.assertEqual(500, resp.status_code)

    def test_get_users_empty(self):
        self.clear_users()
        resp = self.get_users()
        self.assertEqual(200, resp.status_code)
        self.assertEqual(resp.json()['count'], 0)

    def test_get_users_not_empty(self):
        self.clear_users()
        resp = self.create_user(1000, 'user1', '123456')
        resp = self.get_users()
        self.assertEqual(200, resp.status_code)
        self.assertEqual(resp.json()['count'], 1)

        resp = self.create_user(1001, 'user2', '123456')
        resp = self.get_users()
        self.assertEqual(200, resp.status_code)
        self.assertEqual(resp.json()['count'], 2)

    def test_get_user_not_existed(self):
        self.clear_users()
        resp = self.get_user(1000)
        self.assertEqual(404, resp.status_code)
        self.assertEqual(resp.json()['success'], False)

    def test_get_user_existed(self):
        self.clear_users()
        self.create_user(1000, 'user1', '123456')
        resp = self.get_user(1000)
        self.assertEqual(200, resp.status_code)
        self.assertEqual(resp.json()['success'], True)

    def test_update_user_not_existed(self):
        self.clear_users()
        resp = self.update_user(1000, 'user1', '123456')
        self.assertEqual(404, resp.status_code)
        self.assertEqual(resp.json()['success'], False)

    def test_update_user_existed(self):
        self.clear_users()
        self.create_user(1000, 'user1', '123456')
        resp = self.update_user(1000, 'user2', '123456')
        self.assertEqual(200, resp.status_code)
        self.assertEqual(resp.json()['data']['name'], 'user2')

    def test_delete_user_not_existed(self):
        self.clear_users()
        resp = self.delete_user(1000)
        self.assertEqual(404, resp.status_code)
        self.assertEqual(resp.json()['success'], False)

    def test_delete_user_existed(self):
        self.clear_users()
        self.create_user(1000, 'leo', '123456')
        resp = self.delete_user(1000)
        self.assertEqual(200, resp.status_code)
        self.assertEqual(resp.json()['success'], True)
