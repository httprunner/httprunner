import unittest

from starlette.testclient import TestClient

from httprunner.app.main import app

client = TestClient(app)


class TestDebug(unittest.TestCase):
    def test_debug_single_testcase(self):
        json_data = {
            "project_meta": {
                "debugtalk_py": "\ndef hello(name):\n    print(f'hello, {name}')\n",
                "variables": {},
                "env": {},
            },
            "testcase": {
                "config": {
                    "name": "test demo for debug service",
                    "verify": False,
                    "base_url": "",
                    "variables": {},
                    "setup_hooks": [],
                    "teardown_hooks": [],
                    "export": [],
                },
                "teststeps": [
                    {
                        "name": "get index page",
                        "request": {
                            "method": "GET",
                            "url": "https://httpbin.org/",
                            "params": {},
                            "headers": {},
                            "json": {},
                            "cookies": {},
                            "timeout": 30,
                            "allow_redirects": True,
                            "verify": False,
                        },
                        "extract": {},
                        "validate": [],
                    }
                ],
            },
        }
        response = client.post("/hrun/debug/testcase", json=json_data)
        assert response.status_code == 200
        assert response.json()["code"] == 0
