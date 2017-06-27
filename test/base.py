import multiprocessing
import time
import unittest

from ate import utils
from test import api_server


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
        data = utils.handle_req_data(data)
        random_str = utils.gen_random_string(5)
        authorization = utils.gen_md5(token, data, random_str)

        headers = {
            'authorization': authorization,
            'random': random_str
        }
        return headers
