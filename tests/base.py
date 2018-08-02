import multiprocessing
import time
import unittest

import requests
from httprunner import utils
from tests.api_server import app as flask_app

try:
    from httpbin import app as httpbin_app
    HTTPBIN_HOST = "127.0.0.1"
    HTTPBIN_PORT = 3458
except ImportError:
    HTTPBIN_HOST = "httpbin.org"
    HTTPBIN_PORT = 80

FLASK_APP_PORT = 5000
HTTPBIN_SERVER = "http://{}:{}".format(HTTPBIN_HOST, HTTPBIN_PORT)


def run_flask():
    flask_app.run(port=FLASK_APP_PORT)

def run_httpbin():
    if HTTPBIN_HOST == "127.0.0.1":
        httpbin_app.run(host=HTTPBIN_HOST, port=HTTPBIN_PORT)


class ApiServerUnittest(unittest.TestCase):
    """ Test case class that sets up an HTTP server which can be used within the tests
    """

    @classmethod
    def setUpClass(cls):
        cls.host = "http://127.0.0.1:5000"
        cls.flask_process = multiprocessing.Process(
            target=run_flask
        )
        cls.httpbin_process = multiprocessing.Process(
            target=run_httpbin
        )
        cls.flask_process.start()
        cls.httpbin_process.start()
        time.sleep(0.1)
        cls.api_client = requests.Session()

    @classmethod
    def tearDownClass(cls):
        cls.flask_process.terminate()
        cls.httpbin_process.terminate()

    def get_token(self, user_agent, device_sn, os_platform, app_version):
        url = "%s/api/get-token" % self.host
        headers = {
            'Content-Type': 'application/json',
            'User-Agent': user_agent,
            'device_sn': device_sn,
            'os_platform': os_platform,
            'app_version': app_version
        }
        data = {
            'sign': utils.get_sign(user_agent, device_sn, os_platform, app_version)
        }

        resp = self.api_client.post(url, json=data, headers=headers)
        resp_json = resp.json()
        self.assertTrue(resp_json["success"])
        self.assertIn("token", resp_json)
        self.assertEqual(len(resp_json["token"]), 16)
        return resp_json["token"]

    def get_authenticated_headers(self):
        user_agent = 'iOS/10.3'
        device_sn = utils.gen_random_string(15)
        os_platform = 'ios'
        app_version = '2.8.6'

        token = self.get_token(user_agent, device_sn, os_platform, app_version)
        headers = {
            'device_sn': device_sn,
            'token': token
        }
        return headers
