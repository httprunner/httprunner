# encoding: utf-8

import io
import os
import platform
import time
import unittest
from collections import Iterable, OrderedDict
from datetime import datetime

from httprunner import logger
from httprunner.__about__ import __version__
from httprunner.compat import basestring, bytes, json, numeric_types
from jinja2 import Template
from requests.structures import CaseInsensitiveDict


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
    """
    summary = {
        "success": result.wasSuccessful(),
        "stat": {
            'testsRun': result.testsRun,
            'failures': len(result.failures),
            'errors': len(result.errors),
            'skipped': len(result.skipped),
            'expectedFailures': len(result.expectedFailures),
            'unexpectedSuccesses': len(result.unexpectedSuccesses)
        },
        "platform": get_platform()
    }
    summary["stat"]["successes"] = summary["stat"]["testsRun"] \
        - summary["stat"]["failures"] \
        - summary["stat"]["errors"] \
        - summary["stat"]["skipped"] \
        - summary["stat"]["expectedFailures"] \
        - summary["stat"]["unexpectedSuccesses"]

    if getattr(result, "records", None):
        summary["time"] = {
            'start_at': datetime.fromtimestamp(result.start_at),
            'duration': result.duration
        }
        summary["records"] = result.records

    return summary

def make_json_serializable(raw_json):
    serializable_json = {}
    for key, value in raw_json.items():
        if isinstance(value, bytes):
            value = value.decode("utf-8")
        elif isinstance(value, CaseInsensitiveDict):
            value = dict(value)
        elif not isinstance(value, (basestring, numeric_types, Iterable)):
            # class instance, e.g. MultipartEncoder()
            value = repr(value)

        serializable_json[key] = value

    keyorder = ["url", "method", "request_headers", "request_body", "request_time",
        "status_code", "response_headers", "response_body",
        "content_size", "response_time(ms)", "elapsed(ms)"]
    serializable_ordered_json = OrderedDict(
        sorted(serializable_json.items(), key=lambda x:keyorder.index(x[0])))

    return json.dumps(serializable_ordered_json, indent=2, ensure_ascii=False)


class HtmlTestResult(unittest.TextTestResult):
    """A html result class that can generate formatted html results.

    Used by TextTestRunner.
    """
    def __init__(self, stream, descriptions, verbosity):
        super(HtmlTestResult, self).__init__(stream, descriptions, verbosity)
        self.records = []
        self.default_report_template_path = os.path.join(
            os.path.abspath(os.path.dirname(__file__)),
            "templates",
            "default_report_template.html"
        )
        self.report_path = None

    def _record_test(self, test, status, attachment=''):
        self.records.append({
            'name': test.shortDescription(),
            'status': status,
            'response_time_ms': test.meta_data.get("response_time(ms)", 0),
            'attachment': attachment,
            "meta_data": make_json_serializable(test.meta_data)
        })

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

    @property
    def summary(self):
        return get_summary(self)

    def render_html_report(self, html_report_name=None, html_report_template=None):
        """ render html report with specified report name and template
            if html_report_name is not specified, use current datetime
            if html_report_template is not specified, use default report template
        """
        if not html_report_template:
            html_report_template = self.default_report_template_path
            logger.log_debug("No html report template specified, use default.")
        else:
            logger.log_info("render with html report template: {}".format(html_report_template))

        summary = self.summary
        logger.log_info("Start to render Html report ...")
        logger.log_debug("render data: {}".format(summary))

        report_dir_path = os.path.join(os.getcwd(), "reports")
        start_datetime = summary["time"]["start_at"].strftime('%Y-%m-%d-%H-%M-%S')
        if html_report_name:
            summary["html_report_name"] = html_report_name
            report_dir_path = os.path.join(report_dir_path, html_report_name)
            html_report_name += "-{}.html".format(start_datetime)
        else:
            summary["html_report_name"] = ""
            html_report_name = "{}.html".format(start_datetime)

        if not os.path.isdir(report_dir_path):
            os.makedirs(report_dir_path)

        with io.open(html_report_template, "r", encoding='utf-8') as fp_r:
            template_content = fp_r.read()
            report_path = os.path.join(report_dir_path, html_report_name)
            with io.open(report_path, 'w', encoding='utf-8') as fp_w:
                rendered_content = Template(template_content).render(summary)
                fp_w.write(rendered_content)

        logger.log_info("Generated Html report: {}".format(report_path))

        return report_path
