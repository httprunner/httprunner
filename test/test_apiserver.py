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

    def tearDown(self):
        super(TestApiServer, self).tearDown()
        self._api_server.stop_accepting()
        self._api_server.stop()

    def clear_users(self):
        url = "%s/api/user/clear" % self.host
        resp = requests.get(url)
        return resp

    def add_user(self, uid, name, password):
        url = "%s/api/user/add" % self.host
        data = {
            'uid': uid,
            'name': name,
            'password': password
        }
        resp = requests.post(url, json=data)
        return resp

    def test_clear_users(self):
        resp = self.clear_users()
        self.assertEqual(200, resp.status_code)
        self.assertEqual(resp.text, "ok")

    def test_add_user_not_existed(self):
        self.clear_users()
        resp = self.add_user(1000, 'leo', '123456')
        self.assertEqual(200, resp.status_code)
        self.assertEqual(resp.text, "ok")

        url = "%s/api/user/1000" % self.host
        resp = requests.get(url)
        self.assertEqual(200, resp.status_code)
        self.assertNotEqual(resp.json(), {})

    def test_add_user_existed(self):
        self.clear_users()
        resp = self.add_user(1000, 'leo', '123456')
        self.assertEqual(200, resp.status_code)
        self.assertEqual(resp.text, "ok")

        self.add_user(1000, 'leo2', '123456')
        url = "%s/api/user/1000" % self.host
        resp = requests.get(url)
        self.assertEqual(200, resp.status_code)
        self.assertEqual(resp.json()['name'], 'leo2')

    def test_get_user_not_existed(self):
        self.clear_users()
        url = "%s/api/user/1000" % self.host
        resp = requests.get(url)
        self.assertEqual(200, resp.status_code)
        self.assertEqual(resp.json(), {})

    def test_get_user_existed(self):
        self.clear_users()
        self.add_user(1000, 'leo', '123456')
        url = "%s/api/user/1000" % self.host
        resp = requests.get(url)
        self.assertEqual(200, resp.status_code)
        self.assertEqual(resp.json()['name'], 'leo')

    def test_delete_user_not_existed(self):
        self.clear_users()
        url = "%s/api/user/1000" % self.host
        resp = requests.delete(url)
        self.assertEqual(200, resp.status_code)
        self.assertEqual(resp.text, "not_existed")

    def test_delete_user_existed(self):
        self.clear_users()
        resp = self.add_user(1000, 'leo', '123456')
        url = "%s/api/user/1000" % self.host
        resp = requests.delete(url)
        self.assertEqual(200, resp.status_code)
        self.assertEqual(resp.text, "ok")
