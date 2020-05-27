"""
This module handles compatibility issues between testcase format v2 and v3.
"""

from typing import List, Dict, Text


def convert_jmespath(raw: Text) -> Text:
    if raw.startswith("content"):
        return f"body{raw[len('content'):]}"
    elif raw.startswith("json"):
        return f"body{raw[len('json'):]}"

    return raw


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
        teststep = {}
        if "api" in step:
            teststep["testcase"] = step.pop("api")

        teststep["extract"] = convert_extractors(step.pop("extract", {}))
        teststep["validate"] = convert_validators(step.pop("validate", []))
        teststep.update(step)
        v3_content["teststeps"].append(teststep)

    return v3_content
