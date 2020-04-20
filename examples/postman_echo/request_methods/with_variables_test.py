from httprunner.v3.runner import TestCaseRunner
from httprunner.v3.schema import TestsConfig, TestStep


class TestCaseRequestMethodsWithVariables(TestCaseRunner):
    config = TestsConfig(**{
        "name": "request methods testcase with variables",
        "base_url": "https://postman-echo.com",
        "verify": False
    })

    teststeps = [
        TestStep(**{
            "name": "get with params",
            "variables": {
                "foo1": "bar1",
                "foo2": "bar2"
            },
            "request": {
                "method": "GET",
                "url": "/get",
                "params": {
                    "foo1": "$foo1",
                    "foo2": "$foo2"
                },
                "headers": {
                    "User-Agent": "HttpRunner/3.0"
                }
            },
            "validate": [
                {"eq": ["status_code", 200]},
                {"eq": ["body.args.foo1", "bar1"]},
                {"eq": ["body.args.foo2", "bar2"]}
            ]
        }),
        TestStep(**{
            "name": "post raw text",
            "variables": {
                "foo1": "hello world"
            },
            "request": {
                "method": "POST",
                "url": "/post",
                "data": "This is expected to be sent back as part of response body: $foo1.",
                "headers": {
                    "User-Agent": "HttpRunner/3.0",
                    "Content-Type": "text/plain"
                }
            },
            "validate": [
                {"eq": ["status_code", 200]},
                {"eq": ["body.data", "This is expected to be sent back as part of response body: hello world."]},
            ]
        }),
        TestStep(**{
            "name": "post form data",
            "variables": {
                "foo1": "bar1",
                "foo2": "bar2"
            },
            "request": {
                "method": "POST",
                "url": "/post",
                "data": "foo1=$foo1&foo2=$foo2",
                "headers": {
                    "User-Agent": "HttpRunner/3.0",
                    "Content-Type": "application/x-www-form-urlencoded"
                }
            },
            "validate": [
                {"eq": ["status_code", 200]},
                {"eq": ["body.form.foo1", "bar1"]},
                {"eq": ["body.form.foo2", "bar2"]}
            ]
        })
    ]


if __name__ == '__main__':
    TestCaseRequestMethodsWithVariables().run()
