import io
import os
import platform
import time
import unittest
from base64 import b64encode
from collections import Iterable
from datetime import datetime

from jinja2 import Template, escape
from requests.cookies import RequestsCookieJar

from httprunner import __version__, logger
from httprunner.compat import basestring, bytes, json, numeric_types


def get_platform():
    return {
        "httprunner_version": __version__,
        "python_version": "{} {}".format(
            platform.python_implementation(),
            platform.python_version()
        ),
        "platform": platform.platform()
    }


def get_summary(result):
    """ get summary from test result

    Args:
        result (instance): HtmlTestResult() instance

    Returns:
        dict: summary extracted from result.

            {
                "success": True,
                "stat": {},
                "time": {},
                "records": []
            }

    """
    summary = {
        "success": result.wasSuccessful(),
        "stat": {
            'total': result.testsRun,
            'failures': len(result.failures),
            'errors': len(result.errors),
            'skipped': len(result.skipped),
            'expectedFailures': len(result.expectedFailures),
            'unexpectedSuccesses': len(result.unexpectedSuccesses)
        }
    }
    summary["stat"]["successes"] = summary["stat"]["total"] \
                                   - summary["stat"]["failures"] \
                                   - summary["stat"]["errors"] \
                                   - summary["stat"]["skipped"] \
                                   - summary["stat"]["expectedFailures"] \
                                   - summary["stat"]["unexpectedSuccesses"]

    summary["time"] = {
        'start_at': result.start_at,
        'duration': result.duration
    }
    summary["records"] = result.records

    return summary


def aggregate_stat(origin_stat, new_stat):
    """ aggregate new_stat to origin_stat.

    Args:
        origin_stat (dict): origin stat dict, will be updated with new_stat dict.
        new_stat (dict): new stat dict.

    """
    for key in new_stat:
        if key not in origin_stat:
            origin_stat[key] = new_stat[key]
        elif key == "start_at":
            # start datetime
            origin_stat["start_at"] = min(origin_stat["start_at"], new_stat["start_at"])
        elif key == "duration":
            # duration = max_end_time - min_start_time
            max_end_time = max(origin_stat["start_at"] + origin_stat["duration"],
                               new_stat["start_at"] + new_stat["duration"])
            min_start_time = min(origin_stat["start_at"], new_stat["start_at"])
            origin_stat["duration"] = max_end_time - min_start_time
        else:
            origin_stat[key] += new_stat[key]


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
                "json": {
                    "sign": "cb9d60acd09080ea66c8e63a1c78c6459ea00168"
                },
                "verify": false
            }

    """
    for key, value in request_data.items():

        if isinstance(value, list):
            value = json.dumps(value, indent=2, ensure_ascii=False)

        elif isinstance(value, bytes):
            try:
                encoding = "utf-8"
                value = escape(value.decode(encoding))
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
                "json": {
                    "success": false,
                    "data": {}
                }
            }

    """
    for key, value in response_data.items():

        if isinstance(value, list):
            value = json.dumps(value, indent=2, ensure_ascii=False)

        elif isinstance(value, bytes):
            try:
                encoding = response_data.get("encoding")
                if not encoding or encoding == "None":
                    encoding = "utf-8"

                if key == "content" and "image" in response_data["content_type"]:
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


def render_html_report(summary, report_template=None, report_dir=None, report_file=None):
    """ render html report with specified report name and template

    Args:
        report_template (str): specify html report template path
        report_dir (str): specify html report save directory

    """
    if not report_template:
        report_template = os.path.join(
            os.path.abspath(os.path.dirname(__file__)),
            "static",
            "report_template.html"
        )
        logger.log_debug("No html report template specified, use default.")
    else:
        logger.log_info("render with html report template: {}".format(report_template))

    logger.log_info("Start to render Html report ...")

    report_dir = report_dir or os.path.join(os.getcwd(), "reports")
    if not os.path.isdir(report_dir):
        os.makedirs(report_dir)

    start_at_timestamp = int(summary["time"]["start_at"])
    summary["time"]["start_datetime"] = datetime.fromtimestamp(start_at_timestamp).strftime('%Y-%m-%d %H:%M:%S')

    if report_file:
        report_path = os.path.join(report_dir, report_file)
    else:
        report_path = os.path.join(report_dir, "{}.html".format(start_at_timestamp))

    with io.open(report_template, "r", encoding='utf-8') as fp_r:
        template_content = fp_r.read()
        with io.open(report_path, 'w', encoding='utf-8') as fp_w:
            rendered_content = Template(
                template_content,
                extensions=["jinja2.ext.loopcontrols"]
            ).render(summary)
            fp_w.write(rendered_content)

    logger.log_info("Generated Html report: {}".format(report_path))

    return report_path


class HtmlTestResult(unittest.TextTestResult):
    """ A html result class that can generate formatted html results.
        Used by TextTestRunner.
    """
    def __init__(self, stream, descriptions, verbosity):
        super(HtmlTestResult, self).__init__(stream, descriptions, verbosity)
        self.records = []

    def _record_test(self, test, status, attachment=''):
        data = {
            'name': test.shortDescription(),
            'status': status,
            'attachment': attachment,
            "meta_datas": test.meta_datas
        }
        self.records.append(data)

    def startTestRun(self):
        self.start_at = time.time()

    def startTest(self, test):
        """ add start test time """
        super(HtmlTestResult, self).startTest(test)
        logger.color_print(test.shortDescription(), "yellow")

    def addSuccess(self, test):
        super(HtmlTestResult, self).addSuccess(test)
        self._record_test(test, 'success')
        print("")

    def addError(self, test, err):
        super(HtmlTestResult, self).addError(test, err)
        self._record_test(test, 'error', self._exc_info_to_string(err, test))
        print("")

    def addFailure(self, test, err):
        super(HtmlTestResult, self).addFailure(test, err)
        self._record_test(test, 'failure', self._exc_info_to_string(err, test))
        print("")

    def addSkip(self, test, reason):
        super(HtmlTestResult, self).addSkip(test, reason)
        self._record_test(test, 'skipped', reason)
        print("")

    def addExpectedFailure(self, test, err):
        super(HtmlTestResult, self).addExpectedFailure(test, err)
        self._record_test(test, 'ExpectedFailure', self._exc_info_to_string(err, test))
        print("")

    def addUnexpectedSuccess(self, test):
        super(HtmlTestResult, self).addUnexpectedSuccess(test)
        self._record_test(test, 'UnexpectedSuccess')
        print("")

    @property
    def duration(self):
        return time.time() - self.start_at
