# NOTE: Generated By HttpRunner v4.1.4
# FROM: cookie_manipulation/set_delete_cookies.yml
from httprunner import HttpRunner, Config, Step, RunRequest


class TestCaseSetDeleteCookies(HttpRunner):

    config = (
        Config("set & delete cookies.")
        .variables(**{"foo1": "bar1", "foo2": "bar2"})
        .base_url("https://postman-echo.com")
        .verify(False)
        .export(*["cookie_foo1", "cookie_foo3"])
    )

    teststeps = [
        Step(
            RunRequest("set cookie foo1 & foo2 & foo3")
            .with_variables(**{"foo3": "bar3"})
            .get("/cookies/set")
            .with_params(**{"foo1": "bar111", "foo2": "$foo2", "foo3": "$foo3"})
            .with_headers(**{"User-Agent": "HttpRunner/${get_httprunner_version()}"})
            .extract()
            .with_jmespath("$.cookies.foo1", "cookie_foo1")
            .with_jmespath("$.cookies.foo3", "cookie_foo3")
            .validate()
            .assert_equal("status_code", 200)
            .assert_not_equal("$.cookies.foo3", "$foo3")
        ),
        Step(
            RunRequest("delete cookie foo2")
            .get("/cookies/delete?foo2")
            .with_headers(**{"User-Agent": "HttpRunner/${get_httprunner_version()}"})
            .validate()
            .assert_equal("status_code", 200)
            .assert_not_equal("$.cookies.foo1", "$foo1")
            .assert_equal("$.cookies.foo1", "$cookie_foo1")
            .assert_equal("$.cookies.foo3", "$cookie_foo3")
        ),
    ]


if __name__ == "__main__":
    TestCaseSetDeleteCookies().test_start()
