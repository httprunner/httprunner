from examples.postman_echo import debugtalk
from examples.postman_echo.request_methods.validate_with_variables_test \
    import TestCaseRequestMethodsValidateWithVariables
from httprunner.runner import TestCaseRunner
from httprunner.schema import TestsConfig, TestStep


class TestCaseRequestMethodsRefTestcase(TestCaseRunner):
    config = TestsConfig(**{
        "name": "request methods testcase: reference testcase",
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
                "foo1": "override_bar1"
            },
            "testcase": TestCaseRequestMethodsValidateWithVariables
        })
    ]


if __name__ == '__main__':
    TestCaseRequestMethodsRefTestcase().run()
