import os

import jinja2
from loguru import logger

from httprunner.new_loader import load_testcase_file

__TMPL__ = """
from httprunner.runner import TestCaseRunner
from httprunner.schema import TestsConfig, TestStep


class {{ class_name }}(TestCaseRunner):
    config = TestsConfig(**{{ config }})

    teststeps = [
        {% for teststep in teststeps %}
            TestStep(**{{ teststep }}),
        {% endfor %}
    ]


if __name__ == '__main__':
    {{ class_name }}().run()
"""


def make_testcase(path: str) -> str:
    testcase = load_testcase_file(path)
    template = jinja2.Template(__TMPL__)

    raw_file_name, _ = os.path.splitext(os.path.basename(path))
    # convert title case, e.g. request_with_variables => RequestWithVariables
    name_in_title_case = raw_file_name.title().replace("_", "")

    data = {
        "class_name": f"TestCase{name_in_title_case}",
        "config": testcase["config"],
        "teststeps": testcase["teststeps"],
    }
    content = template.render(data)

    testcase_dir = os.path.dirname(path)
    testcase_python_path = os.path.join(testcase_dir, f"{raw_file_name}_test.py")
    with open(testcase_python_path, "w") as f:
        f.write(content)

    logger.info(f"generated testcase: {testcase_python_path}")
    return testcase_python_path
