# NOTE: Generated By HttpRunner v3.1.6
# FROM: request_methods/report_with_allure.yml


import pytest
import debugtalk


import allure


from httprunner import HttpRunner, Config, Step, RunRequest, RunTestCase


class TestCaseReportWithAllure(HttpRunner):
    @allure.title("report with allure")
    @allure.description("description write here.")
    @allure.link("https://github.com/project/issues/153", name="测试用例")
    @allure.link("https://github.com/project/issues/152", name="研发设计文档")
    @allure.link("https://github.com/project/issues/151", name="产品需求文档")
    @allure.link("https://github.com/project/issues/150")
    @pytest.mark.p1
    @pytest.mark.unsafe
    def test_start(self):
        super().test_start()

    config = (
        Config("report with allure")
        .variables(**{"foo1": "session_bar1"})
        .base_url("https://postman-echo.com")
        .verify(False)
    )

    teststeps = [
        Step(
            RunRequest("get with params")
            .with_variables(
                **{"foo1": "bar1", "foo2": "session_bar2", "sum_v": "${sum_two(1, 2)}"}
            )
            .get("/get")
            .with_params(**{"foo1": "$foo1", "foo2": "$foo2", "sum_v": "$sum_v"})
            .with_headers(**{"User-Agent": "HttpRunner/${get_httprunner_version()}"})
            .extract()
            .with_jmespath("body.args.foo2", "session_foo2")
            .validate()
            .assert_equal("status_code", 200)
            .assert_equal("body.args.sum_v", "3")
        ),
    ]


if __name__ == "__main__":
    TestCaseReportWithAllure().test_start()
