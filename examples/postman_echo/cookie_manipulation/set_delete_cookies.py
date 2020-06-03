import unittest
import requests

from httprunner.runner import HttpRunner
from httprunner.schema import TConfig, TStep


class TestCaseSetDeleteCookies(unittest.TestCase):
    config = TConfig(
        **{
            "name": "set & delete cookies.",
            "base_url": "https://postman-echo.com",
            "variables": {"foo1": "bar1", "foo2": "bar2"},
            "verify": False,
            "export": ["cookie_foo1", "cookie_foo3"],
        }
    )

    teststeps = [
        TStep(
            **{
                "name": "set cookie foo1 & foo2 & foo3",
                "variables": {"foo3": "bar3"},
                "request": {
                    "method": "GET",
                    "url": "/cookies/set",
                    "params": {"foo1": "bar111", "foo2": "$foo2", "foo3": "$foo3"},
                    "headers": {"User-Agent": "HttpRunner/${get_httprunner_version()}"},
                },
                "extract": {
                    "cookie_foo1": "$.cookies.foo1",
                    "cookie_foo3": "$.cookies.foo3",
                },
                "validate": [
                    {"eq": ["status_code", 200]},
                    {"eq": ["$.cookies.foo3", "$foo3"]},
                ],
            }
        ),
        TStep(
            **{
                "name": "delete cookie foo2",
                "request": {
                    "method": "GET",
                    "url": "/cookies/delete?foo2",
                    "headers": {"User-Agent": "HttpRunner/${get_httprunner_version()}"},
                },
                "validate": [
                    {"eq": ["status_code", 200]},
                    {"ne": ["$.cookies.foo1", "$foo1"]},
                    {"eq": ["$.cookies.foo1", "$cookie_foo1"]},
                    {"eq": ["$.cookies.foo3", "$cookie_foo3"]},
                ],
            }
        ),
    ]

    def test_start(self):
        s = requests.Session()
        HttpRunner(self.config, self.teststeps, session=s).with_variables(
            foo1="bar123", foo2="bar22"
        ).run()
