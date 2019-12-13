import logging
import random

from locust import HttpLocust, TaskSet, task
from locust.events import request_failure

from httprunner.exceptions import MyBaseError, MyBaseFailure
from httprunner.ext.locusts.utils import prepare_locust_tests
from httprunner.runner import Runner

logging.getLogger().setLevel(logging.CRITICAL)
logging.getLogger('locust.main').setLevel(logging.INFO)
logging.getLogger('locust.runners').setLevel(logging.INFO)


class WebPageTasks(TaskSet):
    def on_start(self):
        config = {}
        self.test_runner = Runner(config, self.client)

    @task
    def test_any(self):
        test_dict = random.choice(self.locust.tests)
        try:
            self.test_runner.run_test(test_dict)
        except (AssertionError, MyBaseError, MyBaseFailure) as ex:
            request_failure.fire(
                request_type=self.test_runner.exception_request_type,
                name=self.test_runner.exception_name,
                response_time=0,
                exception=ex
            )


class WebPageUser(HttpLocust):
    host = ""
    task_set = WebPageTasks
    min_wait = 10
    max_wait = 30

    # file_path is generated on locusts startup
    file_path = "$TESTCASE_FILE"
    tests = prepare_locust_tests(file_path)
