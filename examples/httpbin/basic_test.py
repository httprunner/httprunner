# NOTICE: Generated By HttpRunner. DO'NOT EDIT!
from httprunner import HttpRunner, TConfig, TStep


class TestCaseBasic(HttpRunner):
    config = TConfig(
        **{
            "name": "basic test with httpbin",
            "base_url": "https://httpbin.org/",
            "path": "examples/httpbin/basic_test.py",
        }
    )

    teststeps = [
        TStep(
            **{
                "name": "headers",
                "request": {"url": "/headers", "method": "GET"},
                "validate": [
                    {"eq": ["status_code", 200]},
                    {"eq": ["body.headers.Host", "httpbin.org"]},
                ],
            }
        ),
        TStep(
            **{
                "name": "user-agent",
                "request": {"url": "/user-agent", "method": "GET"},
                "validate": [{"eq": ["status_code", 200]}],
            }
        ),
        TStep(
            **{
                "name": "get without params",
                "request": {"url": "/get", "method": "GET"},
                "validate": [{"eq": ["status_code", 200]}, {"eq": ["body.args", {}]}],
            }
        ),
        TStep(
            **{
                "name": "get with params in url",
                "request": {"url": "/get?a=1&b=2", "method": "GET"},
                "validate": [
                    {"eq": ["status_code", 200]},
                    {"eq": ["body.args", {"a": "1", "b": "2"}]},
                ],
            }
        ),
        TStep(
            **{
                "name": "get with params in params field",
                "request": {"url": "/get", "params": {"a": 1, "b": 2}, "method": "GET"},
                "validate": [
                    {"eq": ["status_code", 200]},
                    {"eq": ["body.args", {"a": "1", "b": "2"}]},
                ],
            }
        ),
        TStep(
            **{
                "name": "set cookie",
                "request": {"url": "/cookies/set?name=value", "method": "GET"},
                "validate": [{"eq": ["status_code", 200]}],
            }
        ),
        TStep(
            **{
                "name": "extract cookie",
                "request": {"url": "/cookies", "method": "GET"},
                "validate": [{"eq": ["status_code", 200]}],
            }
        ),
        TStep(
            **{
                "name": "post data",
                "request": {
                    "url": "/post",
                    "method": "POST",
                    "headers": {"Content-Type": "application/json"},
                    "data": "abc",
                },
                "validate": [{"eq": ["status_code", 200]}],
            }
        ),
        TStep(
            **{
                "name": "validate body length",
                "request": {"url": "/spec.json", "method": "GET"},
                "validate": [{"len_eq": ["body", 9]}],
            }
        ),
    ]


if __name__ == "__main__":
    TestCaseBasic().test_start()
