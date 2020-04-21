from httprunner.v3.runner import TestCaseRunner
from httprunner.v3.schema import TestsConfig, TestStep
from examples.postman_echo import debugtalk


class TestCaseRequestMethodsValidateWithFunctions(TestCaseRunner):
    config = TestsConfig(**{
        "name": "request methods testcase: validate with functions",
        "variables": {
            "foo1": "session_bar1"
        },
        "functions": {
            "get_httprunner_version": debugtalk.get_httprunner_version,
            "sum_two": debugtalk.sum_two
        },
        "base_url": "https://postman-echo.com",
        "verify": False
    })

    teststeps = [
        TestStep(**{
            "name": "get with params",
            "variables": {
                "foo1": "bar1",
                "foo2": "session_bar2",
                "sum_v": "${sum_two(1, 2)}"
            },
            "request": {
                "method": "GET",
                "url": "/get",
                "params": {
                    "foo1": "$foo1",
                    "foo2": "$foo2",
                    "sum_v": "$sum_v"
                },
                "headers": {
                    "User-Agent": "HttpRunner/${get_httprunner_version()}"
                }
            },
            "extract": {
                "session_foo2": "body.args.foo2"
            },
            "validate": [
                {"eq": ["status_code", 200]},
                {"eq": ["body.args.sum_v", 3]},
                {"less_than": ["body.args.sum_v", "${sum_two(2, 2)}"]}
            ]
        })
    ]


if __name__ == '__main__':
    TestCaseRequestMethodsValidateWithFunctions().run()
