import io
import os
from datetime import datetime

from jinja2 import Template

from httprunner import logger
from httprunner.exceptions import SummaryEmpty


def gen_html_report(summary, report_template=None, report_dir=None, report_file=None):
    """ render html report with specified report name and template

    Args:
        summary (dict): test result summary data
        report_template (str): specify html report template path, template should be in Jinja2 format.
        report_dir (str): specify html report save directory
        report_file (str): specify html report file path, this has higher priority than specifying report dir.

    """
    if not summary["time"] or summary["stat"]["testcases"]["total"] == 0:
        logger.log_error("test result summary is empty ! {}".format(summary))
        raise SummaryEmpty

    if not report_template:
        report_template = os.path.join(
            os.path.abspath(os.path.dirname(__file__)),
            "template.html"
        )
        logger.log_debug("No html report template specified, use default.")
    else:
        logger.log_info("render with html report template: {}".format(report_template))

    logger.log_info("Start to render Html report ...")

    start_at_timestamp = int(summary["time"]["start_at"])
    summary["time"]["start_datetime"] = datetime.fromtimestamp(start_at_timestamp).strftime('%Y-%m-%d %H:%M:%S')

    if report_file:
        report_dir = os.path.dirname(report_file)
        report_file_name = os.path.basename(report_file)
    else:
        report_dir = report_dir or os.path.join(os.getcwd(), "reports")
        report_file_name = "{}.html".format(start_at_timestamp)

    if not os.path.isdir(report_dir):
        os.makedirs(report_dir)

    report_path = os.path.join(report_dir, report_file_name)
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

