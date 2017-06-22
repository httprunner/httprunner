import hashlib
import multiprocessing
import random
import string
import time
import unittest

from . import api_server


class ApiServerUnittest(unittest.TestCase):
    """ Test case class that sets up an HTTP server which can be used within the tests
    """

    authentication = False

    @classmethod
    def setUpClass(cls):
        api_server.AUTHENTICATION = cls.authentication
        cls.api_server_process = multiprocessing.Process(
            target=api_server.app.run
        )
        cls.api_server_process.start()
        time.sleep(0.1)

    @classmethod
    def tearDownClass(cls):
        cls.api_server_process.terminate()

    def prepare_headers(self, data=""):
        token = api_server.TOKEN
        random_str = ''.join(
            random.choice(string.ascii_uppercase + string.digits) for _ in range(5))

        authorization_str = "".join([token, data, random_str])
        authorization = hashlib.md5(authorization_str.encode('utf-8')).hexdigest()
        headers = {
            'authorization': authorization,
            'random': random_str
        }
        return headers
