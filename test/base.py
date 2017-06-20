import multiprocessing
import time
import unittest
from . import api_server

class ApiServerUnittest(unittest.TestCase):
    """
    Test case class that sets up an HTTP server which can be used within the tests
    """
    @classmethod
    def setUpClass(cls):
        cls.api_server_process = multiprocessing.Process(
            target=api_server.app.run
        )
        cls.api_server_process.start()
        time.sleep(0.1)

    @classmethod
    def tearDownClass(cls):
        cls.api_server_process.terminate()
