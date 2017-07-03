import os
import requests
from ate import runner, exception, utils
from test.base import ApiServerUnittest

class TestRunnerV2(ApiServerUnittest):

    authentication = True

    def setUp(self):
        self.test_runner = runner.Runner()
        self.clear_users()

    def clear_users(self):
        url = "http://127.0.0.1:5000/api/users"
        return requests.delete(url, headers=self.prepare_headers())

    def test_run_single_testcase_yaml(self):
        testcase_file_path = os.path.join(
            os.getcwd(), 'test/data/simple_demo_auth_hardcode.yml')
        testcases = utils.load_testcases(testcase_file_path)
        testcase = testcases[0]["test"]
        success, _ = self.test_runner.run_test(testcase)
        self.assertTrue(success)

    def test_run_single_testcase_json(self):
        testcase_file_path = os.path.join(
            os.getcwd(), 'test/data/simple_demo_auth_hardcode.json')
        testcases = utils.load_testcases(testcase_file_path)
        testcase = testcases[0]["test"]
        success, _ = self.test_runner.run_test(testcase)
        self.assertTrue(success)

    def test_run_testset_auth_yaml(self):
        testcase_file_path = os.path.join(
            os.getcwd(), 'test/data/simple_demo_auth_hardcode.yml')
        testsets = utils.load_testcases_by_path(testcase_file_path)
        results = self.test_runner.run_testset(testsets[0])
        self.assertEqual(len(results), 2)
        self.assertEqual(results, [(True, []), (True, [])])

    def test_run_testsets_auth_yaml(self):
        testcase_file_path = os.path.join(
            os.getcwd(), 'test/data/simple_demo_auth_hardcode.yml')
        testsets = utils.load_testcases_by_path(testcase_file_path)
        results = self.test_runner.run_testsets(testsets)
        self.assertEqual(len(results), 1)
        self.assertEqual(results[0], [(True, []), (True, [])])

    def test_run_testset_auth_json(self):
        testcase_file_path = os.path.join(
            os.getcwd(), 'test/data/simple_demo_auth_hardcode.json')
        testsets = utils.load_testcases_by_path(testcase_file_path)
        results = self.test_runner.run_testset(testsets[0])
        self.assertEqual(len(results), 2)
        self.assertEqual(results, [(True, []), (True, [])])

    def test_run_testsets_auth_json(self):
        testcase_file_path = os.path.join(
            os.getcwd(), 'test/data/simple_demo_auth_hardcode.json')
        testsets = utils.load_testcases_by_path(testcase_file_path)
        results = self.test_runner.run_testsets(testsets)
        self.assertEqual(len(results), 1)
        self.assertEqual(results[0], [(True, []), (True, [])])

    def test_run_testcase_template_yaml(self):
        testcase_file_path = os.path.join(
            os.getcwd(), 'test/data/demo_template_separate.yml')
        testcases = utils.load_testcases(testcase_file_path)
        success, _ = self.test_runner.run_test(testcases[0]["test"])
        self.assertTrue(success)
        success, _ = self.test_runner.run_test(testcases[1]["test"])
        self.assertTrue(success)

    def test_run_testset_template_yaml(self):
        testcase_file_path = os.path.join(
            os.getcwd(), 'test/data/demo_template_sets.yml')
        testsets = utils.load_testcases_by_path(testcase_file_path)
        results = self.test_runner.run_testset(testsets[0])
        self.assertEqual(len(results), 2)
        self.assertEqual(results, [(True, []), (True, [])])

    def test_run_testsets_template_yaml(self):
        testcase_file_path = os.path.join(
            os.getcwd(), 'test/data/demo_template_sets.yml')
        testsets = utils.load_testcases_by_path(testcase_file_path)
        results = self.test_runner.run_testsets(testsets)
        self.assertEqual(len(results), 1)
        self.assertEqual(results[0], [(True, []), (True, [])])

    def test_run_testset_template_import_functions(self):
        testcase_file_path = os.path.join(
            os.getcwd(), 'test/data/demo_import_functions.yml')
        testsets = utils.load_testcases_by_path(testcase_file_path)
        results = self.test_runner.run_testset(testsets[0])
        self.assertEqual(len(results), 2)
        self.assertEqual(results, [(True, []), (True, [])])

    def test_run_testsets_template_import_functions(self):
        testcase_file_path = os.path.join(
            os.getcwd(), 'test/data/demo_import_functions.yml')
        testsets = utils.load_testcases_by_path(testcase_file_path)
        results = self.test_runner.run_testsets(testsets)
        self.assertEqual(len(results), 1)
        self.assertEqual(results[0], [(True, []), (True, [])])
