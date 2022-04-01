import os
from typing import Text, Callable

from loguru import logger
from httprunner import exceptions
from httprunner.loader import load_testcase_file

from httprunner.step_request import call_hooks
from httprunner.runner import HttpRunner
from httprunner.models import (
    TStep,
    StepData
)


def run_step_testcase(runner: HttpRunner, step: TStep) -> StepData:
    """run teststep: referenced testcase"""
    step_data = StepData(name=step.name)
    step_variables = step.variables
    step_export = step.export

    # setup hooks
    if step.setup_hooks:
        call_hooks(runner, step.setup_hooks, step_variables, "setup testcase")

    # TODO: override testcase with current step name/variables/export

    ref_case_runner = HttpRunner()
    ref_case_runner.config = step.testcase.config
    ref_case_runner.teststeps = step.testcase.teststeps
    ref_case_runner.with_session(runner.session) \
        .with_case_id(runner.case_id) \
        .with_variables(step_variables) \
        .with_export(step_export) \
        .test_start()

    # teardown hooks
    if step.teardown_hooks:
        call_hooks(runner, step.teardown_hooks, step.variables, "teardown testcase")

    summary = ref_case_runner.get_summary()
    step_data.data = summary.step_datas  # list of step data
    step_data.export_vars = summary.in_out.export_vars
    step_data.success = summary.success

    if step_data.export_vars:
        logger.info(f"export variables: {step_data.export_vars}")

    return step_data


class StepRefCase(object):
    def __init__(self, step: TStep):
        self.__step = step

    def teardown_hook(self, hook: Text, assign_var_name: Text = None) -> "StepRefCase":
        if assign_var_name:
            self.__step.teardown_hooks.append({assign_var_name: hook})
        else:
            self.__step.teardown_hooks.append(hook)

        return self

    def export(self, *var_name: Text) -> "StepRefCase":
        self.__step.export.extend(var_name)
        return self

    def struct(self) -> TStep:
        return self.__step

    def name(self) -> Text:
        return self.__step.name

    def type(self) -> Text:
        return f"request-{self.__step.request.method}"

    def run(self, runner: HttpRunner):
        return run_step_testcase(runner, self.__step)


class RunTestCase(object):
    def __init__(self, name: Text):
        self.__step = TStep(name=name)

    def with_variables(self, **variables) -> "RunTestCase":
        self.__step.variables.update(variables)
        return self

    def setup_hook(self, hook: Text, assign_var_name: Text = None) -> "RunTestCase":
        if assign_var_name:
            self.__step.setup_hooks.append({assign_var_name: hook})
        else:
            self.__step.setup_hooks.append(hook)

        return self

    def call(self, testcase: Callable) -> StepRefCase:
        if hasattr(testcase, "config") and hasattr(testcase, "teststeps"):
            self.__step.testcase = testcase
        elif isinstance(testcase, Text):
            if not os.path.isfile(testcase):
                raise exceptions.ParamsError(f"Invalid testcase path: {testcase}")

            self.__step.testcase = load_testcase_file(testcase)
        else:
            raise exceptions.ParamsError(
                f"Invalid teststep referenced testcase: {testcase}"
            )

        return StepRefCase(self.__step)
