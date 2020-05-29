"""
This module handles compatibility issues between testcase format v2 and v3.
"""
import os
from typing import List, Dict, Text, Union

from loguru import logger

from httprunner import exceptions
from httprunner.loader import load_project_meta
from httprunner.utils import sort_dict_by_custom_order


def convert_jmespath(raw: Text) -> Text:
    # content.xx/json.xx => body.xx
    if raw.startswith("content"):
        raw = f"body{raw[len('content'):]}"
    elif raw.startswith("json"):
        raw = f"body{raw[len('json'):]}"

    raw_list = []
    for item in raw.split("."):
        if "-" in item:
            # add quotes for field with separator
            # e.g. headers.Content-Type => headers."Content-Type"
            item = item.strip('"')
            raw_list.append(f'"{item}"')
        elif item.isdigit():
            # convert lst.0.name to lst[0].name
            if len(raw_list) == 0:
                raise exceptions.FileFormatError(
                    f"Invalid jmespath: {raw}, jmespath should startswith headers/body/status_code/cookies"
                )

            last_item = raw_list.pop()
            item = f"{last_item}[{item}]"
            raw_list.append(item)
        else:
            raw_list.append(item)

    return ".".join(raw_list)


def convert_extractors(extractors: Union[List, Dict]) -> Dict:
    """ convert extract list(v2) to dict(v3)

    Args:
        extractors: [{"varA": "content.varA"}, {"varB": "json.varB"}]

    Returns:
        {"varA": "body.varA", "varB": "body.varB"}

    """
    v3_extractors: Dict = {}

    if isinstance(extractors, List):
        for extractor in extractors:
            for k, v in extractor.items():
                v3_extractors[k] = v
    elif isinstance(extractors, Dict):
        v3_extractors = extractors
    else:
        raise exceptions.FileFormatError(f"Invalid extractor: {extractors}")

    for k, v in v3_extractors.items():
        v3_extractors[k] = convert_jmespath(v)

    return v3_extractors


def convert_validators(validators: List) -> List:
    for v in validators:
        if "check" in v and "expect" in v:
            # format1: {"check": "content.abc", "assert": "eq", "expect": 201}
            v["check"] = convert_jmespath(v["check"])

        elif len(v) == 1:
            # format2: {'eq': ['status_code', 201]}
            comparator = list(v.keys())[0]
            v[comparator][0] = convert_jmespath(v[comparator][0])

    return validators


def sort_request_by_custom_order(request: Dict) -> Dict:
    custom_order = [
        "method",
        "url",
        "params",
        "headers",
        "cookies",
        "data",
        "json",
        "files",
        "timeout",
        "allow_redirects",
        "proxies",
        "verify",
        "stream",
        "auth",
        "cert",
    ]
    return sort_dict_by_custom_order(request, custom_order)


def sort_step_by_custom_order(step: Dict) -> Dict:
    custom_order = [
        "name",
        "variables",
        "request",
        "testcase",
        "setup_hooks",
        "teardown_hooks",
        "extract",
        "validate",
        "validate_script",
    ]
    return sort_dict_by_custom_order(step, custom_order)


def ensure_step_attachment(step: Dict) -> Dict:
    test_dict = {
        "name": step["name"],
    }

    if "variables" in step:
        test_dict["variables"] = step["variables"]

    if "setup_hooks" in step:
        test_dict["setup_hooks"] = step["setup_hooks"]

    if "teardown_hooks" in step:
        test_dict["teardown_hooks"] = step["teardown_hooks"]

    if "extract" in step:
        test_dict["extract"] = convert_extractors(step["extract"])

    if "validate" in step:
        test_dict["validate"] = convert_validators(step["validate"])

    if "validate_script" in step:
        test_dict["validate_script"] = step["validate_script"]

    return test_dict


def ensure_testcase_v3_api(api_content: Dict) -> Dict:
    teststep = {
        "request": api_content["request"],
    }
    teststep.update(ensure_step_attachment(api_content))

    teststep = sort_step_by_custom_order(teststep)

    return {
        "config": {"name": api_content["name"]},
        "teststeps": [teststep],
    }


def ensure_testcase_v3(test_content: Dict) -> Dict:
    v3_content = {"config": test_content["config"], "teststeps": []}

    for step in test_content["teststeps"]:
        teststep = {}

        if "request" in step:
            teststep["request"] = step.pop("request")
        elif "api" in step:
            teststep["testcase"] = step.pop("api")
        elif "testcase" in step:
            teststep["testcase"] = step.pop("testcase")

        teststep.update(ensure_step_attachment(step))
        teststep = sort_step_by_custom_order(teststep)
        v3_content["teststeps"].append(teststep)

    return v3_content


def ensure_cli_args(args: List) -> List:
    """ ensure compatibility with deprecated cli args in v2
    """
    # remove deprecated --failfast
    if "--failfast" in args:
        args.pop(args.index("--failfast"))

    # convert --report-file to --html
    if "--report-file" in args:
        index = args.index("--report-file")
        args[index] = "--html"
        args.append("--self-contained-html")

    # keep compatibility with --save-tests in v2
    if "--save-tests" in args:
        args.pop(args.index("--save-tests"))
        generate_conftest_for_summary(args)

    return args


def generate_conftest_for_summary(args: List):

    for arg in args:
        if os.path.exists(arg):
            test_path = arg
            # FIXME: several test paths maybe specified
            break
    else:
        raise exceptions.FileNotFound(f"No test path specified!")

    project_meta = load_project_meta(test_path)
    conftest_path = os.path.join(project_meta.PWD, "conftest.py")
    if os.path.isfile(conftest_path):
        return

    conftest_content = '''# NOTICE: Generated By HttpRunner.
import json
import os
import time

import pytest
from loguru import logger

from httprunner.utils import get_platform


@pytest.fixture(scope="session", autouse=True)
def session_fixture(request):
    """setup and teardown each task"""
    logger.info(f"start running testcases ...")

    start_at = time.time()

    yield

    logger.info(f"task finished, generate task summary for --save-tests")

    summary = {
        "success": True,
        "stat": {
            "testcases": {"total": 0, "success": 0, "fail": 0},
            "teststeps": {"total": 0, "failures": 0, "successes": 0},
        },
        "time": {"start_at": start_at, "duration": time.time() - start_at},
        "platform": get_platform(),
        "details": [],
    }

    for item in request.node.items:
        testcase_summary = item.instance.get_summary()
        summary["success"] &= testcase_summary.success

        summary["stat"]["testcases"]["total"] += 1
        summary["stat"]["teststeps"]["total"] += len(testcase_summary.step_datas)
        if testcase_summary.success:
            summary["stat"]["testcases"]["success"] += 1
            summary["stat"]["teststeps"]["successes"] += len(
                testcase_summary.step_datas
            )
        else:
            summary["stat"]["testcases"]["fail"] += 1
            summary["stat"]["teststeps"]["successes"] += (
                len(testcase_summary.step_datas) - 1
            )
            summary["stat"]["teststeps"]["failures"] += 1

        summary["details"].append(testcase_summary.dict())

    summary_path = "{{SUMMARY_PATH_PLACEHOLDER}}"
    summary_dir = os.path.dirname(summary_path)
    os.makedirs(summary_dir, exist_ok=True)

    with open(summary_path, "w", encoding="utf-8") as f:
        json.dump(summary, f, indent=4)

    logger.info(f"generated task summary: {summary_path}")

'''

    test_path = os.path.abspath(test_path)
    logs_dir_path = os.path.join(project_meta.PWD, "logs")
    test_path_relative_path = test_path[len(project_meta.PWD) + 1 :]

    if os.path.isdir(test_path):
        file_foder_path = os.path.join(logs_dir_path, test_path_relative_path)
        dump_file_name = "all.summary.json"
    else:
        file_relative_folder_path, test_file = os.path.split(test_path_relative_path)
        file_foder_path = os.path.join(logs_dir_path, file_relative_folder_path)
        test_file_name, _ = os.path.splitext(test_file)
        dump_file_name = f"{test_file_name}.summary.json"

    summary_path = os.path.join(file_foder_path, dump_file_name)
    conftest_content = conftest_content.replace(
        "{{SUMMARY_PATH_PLACEHOLDER}}", summary_path
    )

    with open(conftest_path, "w", encoding="utf-8") as f:
        f.write(conftest_content)

    logger.info("generated conftest.py to generate summary.json")
