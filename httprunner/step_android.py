from typing import Text

from loguru import logger
import uiautomator2 as u2

from httprunner.models import IStep, StepResult, TStep, TStepAndroidUI
from httprunner.runner import HttpRunner


def run_android_ui(runner: HttpRunner, step: TStep) -> StepResult:
    step_result = StepResult(
        name=step.name,
        step_type="android_ui",
        success=False,
    )
    logger.info(f"run android ui action: {step.android.method}, param: {step.android.param}")

    return step_result


class StepAndroidControl(IStep):

    def __init__(self, step: TStep):
        self.__step = step

    def start_app(self, package_name: Text) -> "StepAndroidControl":
        return self

    def stop_app(self, package_name: Text) -> "StepAndroidControl":
        return self

    def start_watcher(self) -> "StepAndroidControl":
        return self

    def stop_watcher(self) -> "StepAndroidControl":
        return self

    def start_camera(self) -> "StepAndroidControl":
        return self

    def stop_camera(self) -> "StepAndroidControl":
        return self

    def start_record(self) -> "StepAndroidControl":
        return self

    def stop_record(self) -> "StepAndroidControl":
        return self

    def struct(self) -> TStep:
        return self.__step

    def name(self) -> Text:
        return self.__step.name

    def type(self) -> Text:
        return "android-control"

    def run(self, runner: HttpRunner):
        return run_android_ui(runner, self.__step)


class StepAndroidUI(IStep):

    def __init__(self, step: TStep):
        self.__step = step

    def press_back(self) -> "StepAndroidUI":
        self.__step.android.method = "press"
        self.__step.android.param = "back"
        return self

    def press_home(self) -> "StepAndroidUI":
        self.__step.android.method = "press"
        self.__step.android.param = "home"
        return self

    def sleep(self, time: int) -> "StepAndroidUI":
        self.__step.android.method = "sleep"
        self.__step.android.param = time
        return self

    def swipe_up(self) -> "StepAndroidUI":
        self.__step.android.method = "swipe"
        self.__step.android.param = [0.25, 0.5, 0.75, 0.5]
        return self

    def swipe_down(self) -> "StepAndroidUI":
        self.__step.android.method = "swipe"
        self.__step.android.param = [0.75, 0.5, 0.25, 0.5]
        return self

    def swipe_left(self) -> "StepAndroidUI":
        self.__step.android.method = "swipe"
        self.__step.android.param = [0.5, 0.75, 0.5, 0.25]
        return self

    def swipe_right(self) -> "StepAndroidUI":
        self.__step.android.method = "swipe"
        self.__step.android.param = [0.5, 0.25, 0.5, 0.75]
        return self

    def swipe(self, from_x: float, from_y: float, to_x: float, to_y: float) -> "StepAndroidUI":
        self.__step.android.method = "swipe"
        self.__step.android.param = [from_x, from_y, to_x, to_y]
        return self

    def click(self, text: Text) -> "StepAndroidUI":
        self.__step.android.method = "click"
        self.__step.android.param = text
        return self

    def struct(self) -> TStep:
        return self.__step

    def name(self) -> Text:
        return self.__step.name

    def type(self) -> Text:
        return "android-ui"

    def run(self, runner: HttpRunner):
        return run_android_ui(runner, self.__step)


class RunAndroidUI(object):

    def __init__(self, name: Text):
        self.__step = TStep(name=name)
        self.__step.android = TStepAndroidUI()

    def control(self) -> StepAndroidControl:
        return StepAndroidControl(self.__step)

    def ui(self) -> StepAndroidUI:
        return StepAndroidUI(self.__step)
