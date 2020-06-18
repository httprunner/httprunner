import random

from locust import task, HttpUser, between

from httprunner.ext.locust import prepare_locust_tests


class HttpRunnerUser(HttpUser):
    host = ""
    wait_time = between(5, 15)

    def on_start(self):
        locust_tests = prepare_locust_tests()
        self.testcase_runners = [
            testcase().with_session(self.client) for testcase in locust_tests
        ]

    @task
    def test_any(self):
        test_runner = random.choice(self.testcase_runners)
        try:
            test_runner.run()
        except Exception as ex:
            self.environment.events.request_failure.fire(
                request_type="Failed",
                name=test_runner.config.name,
                response_time=0,
                response_length=0,
                exception=ex,
            )
