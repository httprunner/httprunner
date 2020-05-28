"""
This module handles compatibility issues between testcase format v2 and v3.
"""

from typing import List, Dict, Text, Union

from httprunner import exceptions
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
            # add quotes for field with separator, e.g. headers.Content-Type
            raw_list.append(f'"{item}"')
        elif item.isdigit():
            # convert lst.0.name to lst[0].name
            raw_list.append(f"[{item}]")
        else:
            raw_list.append(item)

    # lst.[0].name => lst[0].name
    return ".".join(raw_list).replace(".[", "[")


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
    custom_order = ["name", "variables", "request", "testcase", "extract", "validate"]
    return sort_dict_by_custom_order(step, custom_order)


def ensure_step_attachment(step: Dict) -> Dict:
    test_dict = {
        "name": step["name"],
    }

    if "variables" in step:
        test_dict["variables"] = step["variables"]

    if "extract" in step:
        test_dict["extract"] = convert_extractors(step["extract"])

    if "validate" in step:
        test_dict["validate"] = convert_validators(step["validate"])

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

    # remove deprecated --save-tests
    if "--save-tests" in args:
        args.pop(args.index("--save-tests"))

    return args
