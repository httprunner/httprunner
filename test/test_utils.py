import os
import unittest
from ate import utils
from ate import exception

class TestUtils(unittest.TestCase):

    def test_load_testcases_bad_filepath(self):
        testcase_file_path = os.path.join(os.getcwd(), 'test/data/demo')
        with self.assertRaises(exception.ParamsError):
            utils.load_testcases(testcase_file_path)

    def test_load_json_testcases(self):
        testcase_file_path = os.path.join(os.getcwd(), 'test/data/demo.json')
        testcases = utils.load_testcases(testcase_file_path)
        self.assertEqual(len(testcases), 2)
        self.assertIn('name', testcases[0])
        self.assertIn('request', testcases[0])
        self.assertIn('response', testcases[0])
        self.assertIn('url', testcases[0]['request'])
        self.assertIn('method', testcases[0]['request'])

    def test_load_yaml_testcases(self):
        testcase_file_path = os.path.join(os.getcwd(), 'test/data/demo.yml')
        testcases = utils.load_testcases(testcase_file_path)
        self.assertEqual(len(testcases), 2)
        self.assertIn('name', testcases[0])
        self.assertIn('request', testcases[0])
        self.assertIn('response', testcases[0])
        self.assertIn('url', testcases[0]['request'])
        self.assertIn('method', testcases[0]['request'])
