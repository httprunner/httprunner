from httprunner.runner import TestCaseRunner
from httprunner.schema import TestsConfig, TestStep


class TestCaseRequestMethodsValidateWithVariables(TestCaseRunner):
    config = TestsConfig(**{
        "name": "request methods testcase: validate with variables",
        "variables": {
            "foo1": "session_bar1"
        },
        "base_url": "https://postman-echo.com",
        "verify": False
    })

    teststeps = [
        TestStep(**{
            "name": "get with params",
            "variables": {
                "foo1": "bar1",
                "foo2": "session_bar2"
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
            "extract": {
                "session_foo2": "body.args.foo2"
            },
            "validate": [
                {"eq": ["status_code", 200]},
                {"eq": ["body.args.foo1", "session_bar1"]},
                {"eq": ["body.args.foo1", "$foo1"]},
                {"eq": ["body.args.foo2", "session_bar2"]},
                {"eq": ["body.args.foo2", "$foo2"]}
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
                    "User-Agent": "HttpRunner/3.0",
                    "Content-Type": "text/plain"
                }
            },
            "validate": [
                {"eq": ["status_code", 200]},
                {"eq": [
                    "body.data",
                    "This is expected to be sent back as part of response body: session_bar1-session_bar2."
                ]},
                {"eq": [
                    "body.data",
                    "This is expected to be sent back as part of response body: $foo1-$foo3."
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
                    "User-Agent": "HttpRunner/3.0",
                    "Content-Type": "application/x-www-form-urlencoded"
                }
            },
            "validate": [
                {"eq": ["status_code", 200]},
                {"eq": ["body.form.foo1", "session_bar1"]},
                {"eq": ["body.form.foo1", "$foo1"]},
                {"eq": ["body.form.foo2", "bar2"]},
                {"eq": ["body.form.foo2", "$foo2"]}
            ]
        })
    ]


if __name__ == '__main__':
    runner = TestCaseRequestMethodsValidateWithVariables().run()
    print(runner.case_datas)
