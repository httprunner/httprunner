# -*- coding: utf-8 -*-
"""

  @Date     :     2022/4/7
  @File     :     request_with_retry.py
  @Author   :     duanchao.bill
  @Desc     :

"""

from httprunner import HttpRunner, Config, Step, RunRequest, RunTestCase


class TestCaseRetry(HttpRunner):
    config = (
        Config("request methods testcase in hardcode")
        .base_url("https://postman-echo.com")
        .verify(False)
    )

    teststeps = [
        Step(
            RunRequest("run with retry")
            .with_retry(retry_times=1, retry_interval=1)
            .get("/get")
            .with_params(**{"foo1": "${fake_randnum()}"})
            .with_headers(**{"User-Agent": "HttpRunner/3.0"})
            .validate()
            .assert_equal("body.args.foo1", "2")
        )
    ]
