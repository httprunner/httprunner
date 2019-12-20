from base64 import b64encode
from collections import Iterable

from jinja2 import escape
from requests.cookies import RequestsCookieJar

from httprunner.compat import basestring, bytes, json, numeric_types, JSONDecodeError


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
                    except JSONDecodeError:
                        pass
                value = escape(value)
            except UnicodeDecodeError:
                pass

        elif not isinstance(value, (basestring, numeric_types, Iterable)):
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

        elif not isinstance(value, (basestring, numeric_types, Iterable)):
            # class instance, e.g. MultipartEncoder()
            value = repr(value)

        elif isinstance(value, RequestsCookieJar):
            value = value.get_dict()

        response_data[key] = value


def __expand_meta_datas(meta_datas, meta_datas_expanded):
    """ expand meta_datas to one level

    Args:
        meta_datas (dict/list): maybe in nested format

    Returns:
        list: expanded list in one level

    Examples:
        >>> meta_datas = [
                [
                    dict1,
                    dict2
                ],
                dict3
            ]
        >>> meta_datas_expanded = []
        >>> __expand_meta_datas(meta_datas, meta_datas_expanded)
        >>> print(meta_datas_expanded)
            [dict1, dict2, dict3]

    """
    if isinstance(meta_datas, dict):
        meta_datas_expanded.append(meta_datas)
    elif isinstance(meta_datas, list):
        for meta_data in meta_datas:
            __expand_meta_datas(meta_data, meta_datas_expanded)


def __get_total_response_time(meta_datas_expanded):
    """ caculate total response time of all meta_datas
    """
    try:
        response_time = 0
        for meta_data in meta_datas_expanded:
            response_time += meta_data["stat"]["response_time_ms"]

        return "{:.2f}".format(response_time)

    except TypeError:
        # failure exists
        return "N/A"


def __stringify_meta_datas(meta_datas):

    if isinstance(meta_datas, list):
        for _meta_data in meta_datas:
            __stringify_meta_datas(_meta_data)
    elif isinstance(meta_datas, dict):
        data_list = meta_datas["data"]
        for data in data_list:
            __stringify_request(data["request"])
            __stringify_response(data["response"])


def stringify_summary(summary):
    """ stringify summary, in order to dump json file and generate html report.
    """
    for index, suite_summary in enumerate(summary["details"]):

        if not suite_summary.get("name"):
            suite_summary["name"] = "testcase {}".format(index)

        for record in suite_summary.get("records"):
            meta_datas = record['meta_datas']
            __stringify_meta_datas(meta_datas)
            meta_datas_expanded = []
            __expand_meta_datas(meta_datas, meta_datas_expanded)
            record["meta_datas_expanded"] = meta_datas_expanded
            record["response_time"] = __get_total_response_time(meta_datas_expanded)
