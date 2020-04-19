from httprunner.v3.runner import TestCaseRunner
from httprunner.v3.schema import TestsConfig, TestStep


class TestCaseRequestMethodsHardcode(TestCaseRunner):
    config = TestsConfig(**{
        "name": "request methods testcase in hardcode",
        "base_url": "https://postman-echo.com",
        "verify": False
    })

    teststeps = [
        TestStep(**{
            "name": "get with params",
            "request": {
                "method": "GET",
                "url": "/get",
                "params": {
                    "foo1": "bar1",
                    "foo2": "bar2"
                },
                "headers": {
                    "User-Agent": "HttpRunner/3.0"
                }
            },
            "validate": [
                {"eq": ["status_code", 200]},
                {"eq": ["headers.Server", "nginx"]}
            ]
        }),
        TestStep(**{
            "name": "post raw text",
            "request": {
                "method": "POST",
                "url": "/post",
                "data": "This is expected to be sent back as part of response body.",
                "headers": {
                    "User-Agent": "HttpRunner/3.0",
                    "Content-Type": "text/plain"
                }
            },
            "validate": [
                {"eq": ["status_code", 200]}
            ]
        }),
        TestStep(**{
            "name": "post form data",
            "request": {
                "method": "POST",
                "url": "/post",
                "data": "foo1=bar1&foo2=bar2",
                "headers": {
                    "User-Agent": "HttpRunner/3.0",
                    "Content-Type": "application/x-www-form-urlencoded"
                }
            },
            "validate": [
                {"eq": ["status_code", 200]}
            ]
        }),
        TestStep(**{
            "name": "put request",
            "request": {
                "method": "PUT",
                "url": "/put",
                "data": "This is expected to be sent back as part of response body.",
                "headers": {
                    "User-Agent": "HttpRunner/3.0",
                    "Content-Type": "text/plain"
                }
            },
            "validate": [
                {"eq": ["status_code", 200]}
            ]
        })
    ]


if __name__ == '__main__':
    TestCaseRequestMethodsHardcode().run()
