# NOTE: Generated By HttpRunner v3.0.9
# FROM: examples/postman_echo/request_methods/request_with_testcase_reference.yml

import os
import sys

sys.path.insert(0, os.getcwd())

from httprunner import HttpRunner, Config, Step, RunRequest, RunTestCase

from examples.postman_echo.request_methods.request_with_functions_test import (
    TestCaseRequestWithFunctions as RequestWithFunctions,
)


class TestCaseRequestWithTestcaseReference(HttpRunner):
    config = (
        Config("request with referenced testcase")
        .variables(**{"foo1": "session_bar1", "var2": "testsuite_val2"})
        .base_url("https://postman-echo.com")
        .verify(False)
    )

    teststeps = [
        Step(
            RunTestCase("request with functions")
            .with_variables(**{"foo1": "override_bar1"})
            .call(RequestWithFunctions)
            .export(*["session_foo2"])
        ),
        Step(
            RunRequest("post form data")
            .with_variables(**{"foo1": "bar1"})
            .post("/post")
            .with_headers(
                **{
                    "User-Agent": "HttpRunner/${get_httprunner_version()}",
                    "Content-Type": "application/x-www-form-urlencoded",
                }
            )
            .with_data("foo1=$foo1&foo2=$session_foo2")
            .validate()
            .assert_equal("status_code", 200)
            .assert_equal("body.form.foo1", "session_bar1")
            .assert_equal("body.form.foo2", "session_bar2")
        ),
    ]


if __name__ == "__main__":
    TestCaseRequestWithTestcaseReference().test_start()
