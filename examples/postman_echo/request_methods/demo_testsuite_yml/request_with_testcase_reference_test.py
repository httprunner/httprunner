# NOTE: Generated By HttpRunner v3.1.3
# FROM: request_methods/request_with_testcase_reference.yml

import sys
from pathlib import Path

sys.path.insert(0, str(Path(__file__).parent.parent))

from httprunner import HttpRunner, Config, Step, RunRequest, RunTestCase

from request_methods.request_with_functions_test import (
    TestCaseRequestWithFunctions as RequestWithFunctions,
)


class TestCaseRequestWithTestcaseReference(HttpRunner):

    config = (
        Config("request with referenced testcase")
        .variables(
            **{
                "foo1": "testcase_ref_bar12",
                "expect_foo1": "testcase_ref_bar12",
                "expect_foo2": "testcase_ref_bar22",
                "foo2": "testcase_ref_bar22",
            }
        )
        .base_url("https://postman-echo.com")
        .verify(False)
        .locust_weight(3)
    )

    teststeps = [
        Step(
            RunTestCase("request with functions")
            .with_variables(
                **{"foo1": "testcase_ref_bar1", "expect_foo1": "testcase_ref_bar1"}
            )
            .setup_hook("${sleep(0.1)}")
            .call(RequestWithFunctions)
            .teardown_hook("${sleep(0.2)}")
            .export(*["foo3"])
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
            .with_data("foo1=$foo1&foo2=$foo3")
            .validate()
            .assert_equal("status_code", 200)
            .assert_equal("body.form.foo1", "bar1")
            .assert_equal("body.form.foo2", "bar21")
        ),
    ]


if __name__ == "__main__":
    TestCaseRequestWithTestcaseReference().test_start()
