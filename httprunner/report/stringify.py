import json
from base64 import b64encode
from collections import Iterable
from typing import List

from jinja2 import escape
from requests.cookies import RequestsCookieJar

from httprunner.v3.schema import TestSuiteSummary, SessionData


def dumps_json(value):
    """ dumps json value to indented string

    Args:
        value (dict): raw json data

    Returns:
        str: indented json dump string

    """
    return json.dumps(value, indent=2, ensure_ascii=False)


def detect_encoding(value):
    try:
        return json.detect_encoding(value)
    except AttributeError:
        return "utf-8"


def __stringify_request(request_data):
    """ stringfy HTTP request data

    Args:
        request_data (dict): HTTP request data in dict.

            {
                "url": "http://127.0.0.1:5000/api/get-token",
                "method": "POST",
                "headers": {
                    "User-Agent": "python-requests/2.20.0",
                    "Accept-Encoding": "gzip, deflate",
                    "Accept": "*/*",
                    "Connection": "keep-alive",
                    "user_agent": "iOS/10.3",
                    "device_sn": "TESTCASE_CREATE_XXX",
                    "os_platform": "ios",
                    "app_version": "2.8.6",
                    "Content-Type": "application/json",
                    "Content-Length": "52"
                },
                "body": b'{"sign": "cb9d60acd09080ea66c8e63a1c78c6459ea00168"}',
                "verify": false
            }

    """
    for key, value in request_data.items():

        if isinstance(value, (list, dict)):
            value = dumps_json(value)

        elif isinstance(value, bytes):
            try:
                encoding = detect_encoding(value)
                value = value.decode(encoding)
                if key == "body":
                    try:
                        # request body is in json format
                        value = json.loads(value)
                        value = dumps_json(value)
                    except json.JSONDecodeError:
                        pass
                value = escape(value)
            except UnicodeDecodeError:
                pass

        elif not isinstance(value, (str, bytes, int, float, Iterable)):
            # class instance, e.g. MultipartEncoder()
            value = repr(value)

        elif isinstance(value, RequestsCookieJar):
            value = value.get_dict()

        request_data[key] = value


def __stringify_response(response_data):
    """ stringfy HTTP response data

    Args:
        response_data (dict):

            {
                "status_code": 404,
                "headers": {
                    "Content-Type": "application/json",
                    "Content-Length": "30",
                    "Server": "Werkzeug/0.14.1 Python/3.7.0",
                    "Date": "Tue, 27 Nov 2018 06:19:27 GMT"
                },
                "encoding": "None",
                "content_type": "application/json",
                "ok": false,
                "url": "http://127.0.0.1:5000/api/users/9001",
                "reason": "NOT FOUND",
                "cookies": {},
                "body": {
                    "success": false,
                    "data": {}
                }
            }

    """
    for key, value in response_data.items():

        if isinstance(value, (list, dict)):
            value = dumps_json(value)

        elif isinstance(value, bytes):
            try:
                encoding = response_data.get("encoding")
                if not encoding or encoding == "None":
                    encoding = detect_encoding(value)

                if key == "body" and "image" in response_data["content_type"]:
                    # display image
                    value = "data:{};base64,{}".format(
                        response_data["content_type"],
                        b64encode(value).decode(encoding)
                    )
                else:
                    value = escape(value.decode(encoding))
            except UnicodeDecodeError:
                pass

        elif not isinstance(value, (str, bytes, int, float, Iterable)):
            # class instance, e.g. MultipartEncoder()
            value = repr(value)

        elif isinstance(value, RequestsCookieJar):
            value = value.get_dict()

        response_data[key] = value


def __get_total_response_time(step_datas: List[SessionData]):
    """ caculate total response time of all step_datas
    """
    try:
        response_time = 0
        for step_data in step_datas:
            response_time += step_data.stat.response_time_ms

        return "{:.2f}".format(response_time)

    except TypeError:
        # failure exists
        return "N/A"


def stringify_summary(testsuite_summary: TestSuiteSummary):
    """ stringify summary, in order to dump json file and generate html report.
    """
    for index, testcase_summary in enumerate(testsuite_summary.details):

        if not testcase_summary.name:
            testcase_summary.name = f"testcase {index}"

        step_datas = testcase_summary.step_datas
        for session_data in step_datas:
            req_resp_list = session_data.req_resp
            for req_resp in req_resp_list:
                __stringify_request(req_resp["request"])
                __stringify_response(req_resp["response"])

        testcase_summary.total_response_time = __get_total_response_time(step_datas)
