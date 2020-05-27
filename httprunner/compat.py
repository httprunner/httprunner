"""
This module handles compatibility issues between testcase format v2 and v3.
"""

from typing import List, Dict, Text


def convert_jmespath(raw: Text) -> Text:
    # content.xx/json.xx => body.xx
    if raw.startswith("content"):
        return f"body{raw[len('content'):]}"
    elif raw.startswith("json"):
        return f"body{raw[len('json'):]}"

    # add quotes for field with separator, e.g. headers.Content-Type
    raw_list = []
    for item in raw.split("."):
        if "-" in item:
            raw_list.append(f'"{item}"')
        else:
            raw_list.append(item)

    return ".".join(raw_list)


def convert_extractors(extractors: List) -> Dict:
    """ convert extract list(v2) to dict(v3)

    Args:
        extractors: [{"varA": "content.varA"}, {"varB": "json.varB"}]

    Returns:
        {"varA": "body.varA", "varB": "body.varB"}

    """
    if isinstance(extractors, Dict):
        return extractors

    v3_extractors = {}
    for extractor in extractors:
        for k, v in extractor.items():
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


def ensure_testcase_v3_api(api_content: Dict) -> Dict:

    return {
        "config": {"name": api_content["name"]},
        "teststeps": [
            {
                "name": api_content["name"],
                "variables": api_content.get("variables", {}),
                "request": api_content["request"],
                "validate": convert_validators(api_content.get("validate", [])),
                "extract": convert_extractors(api_content.get("extract", {})),
            }
        ],
    }


def ensure_testcase_v3(test_content: Dict) -> Dict:

    v3_content = {"config": test_content["config"], "teststeps": []}

    for step in test_content["teststeps"]:
        teststep = {"name": step.pop("name", "")}
        if "variables" in step:
            teststep["variables"] = step.pop("variables")

        if "request" in step:
            teststep["request"] = step.pop("request")
        elif "api" in step:
            teststep["testcase"] = step.pop("api")
        elif "testcase" in step:
            teststep["testcase"] = step.pop("testcase")

        if "extract" in step:
            teststep["extract"] = convert_extractors(step.pop("extract"))

        if "validate" in step:
            teststep["validate"] = convert_validators(step.pop("validate"))

        teststep.update(step)
        v3_content["teststeps"].append(teststep)

    return v3_content
