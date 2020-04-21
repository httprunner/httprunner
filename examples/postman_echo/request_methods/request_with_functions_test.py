from httprunner.v3.runner import TestCaseRunner
from httprunner.v3.schema import TestsConfig, TestStep
from examples.postman_echo import debugtalk


class TestCaseRequestMethodsWithFunctions(TestCaseRunner):
    config = TestsConfig(**{
        "name": "request methods testcase with functions",
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
                {"eq": ["body.args.foo1", "session_bar1"]},
                {"eq": ["body.args.foo2", "session_bar2"]},
                {"eq": ["body.args.sum_v", 3]}
            ]
        }),
        TestStep(**{
            "name": "post raw text",
            "variables": {
                "foo1": "hello world",
                "foo3": "$session_foo2"
            },
            "request": {
                "method": "POST",
                "url": "/post",
                "data": "This is expected to be sent back as part of response body: $foo1-$foo3.",
                "headers": {
                    "User-Agent": "HttpRunner/${get_httprunner_version()}",
                    "Content-Type": "text/plain"
                }
            },
            "validate": [
                {"eq": ["status_code", 200]},
                {"eq": [
                    "body.data",
                    "This is expected to be sent back as part of response body: session_bar1-session_bar2."
                ]},
            ]
        }),
        TestStep(**{
            "name": "post form data",
            "variables": {
                "foo1": "session_bar1",
                "foo2": "bar2"
            },
            "request": {
                "method": "POST",
                "url": "/post",
                "data": "foo1=$foo1&foo2=$foo2",
                "headers": {
                    "User-Agent": "HttpRunner/${get_httprunner_version()}",
                    "Content-Type": "application/x-www-form-urlencoded"
                }
            },
            "validate": [
                {"eq": ["status_code", 200]},
                {"eq": ["body.form.foo1", "session_bar1"]},
                {"eq": ["body.form.foo2", "bar2"]}
            ]
        })
    ]


if __name__ == '__main__':
    TestCaseRequestMethodsWithFunctions().run()
