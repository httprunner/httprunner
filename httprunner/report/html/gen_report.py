import io
import os

from jinja2 import Template
from loguru import logger

from httprunner.exceptions import SummaryEmpty
from httprunner.v3.schema import TestSuiteSummary


def gen_html_report(testsuite_summary: TestSuiteSummary, report_template=None, report_dir=None, report_file=None):
    """ render html report with specified report name and template

    Args:
        testsuite_summary (dict): testsuite result summary data
        report_template (str): specify html report template path, template should be in Jinja2 format.
        report_dir (str): specify html report save directory
        report_file (str): specify html report file path, this has higher priority than specifying report dir.

    """
    if testsuite_summary.stat.total == 0:
        logger.error(f"test result testsuite_summary is empty ! {testsuite_summary}")
        raise SummaryEmpty

    if not report_template:
        report_template = os.path.join(
            os.path.abspath(os.path.dirname(__file__)),
            "template.html"
        )
        logger.debug("No html report template specified, use default.")
    else:
        logger.info(f"render with html report template: {report_template}")

    logger.info("Start to render Html report ...")

    if report_file:
        report_dir = os.path.dirname(report_file)
        report_file_name = os.path.basename(report_file)
    else:
        report_dir = report_dir or os.path.join(os.getcwd(), "reports")
        # fix #826: Windows does not support file name include ":"
        report_file_name = "{}.html".format(testsuite_summary.time.start_at_iso_format.replace(":", "").replace("-", ""))

    if not os.path.isdir(report_dir):
        os.makedirs(report_dir)

    report_path = os.path.join(report_dir, report_file_name)
    with io.open(report_template, "r", encoding='utf-8') as fp_r:
        template_content = fp_r.read()
        with io.open(report_path, 'w', encoding='utf-8') as fp_w:
            rendered_content = Template(
                template_content,
                extensions=["jinja2.ext.loopcontrols"]
            ).render(testsuite_summary.dict())
            fp_w.write(rendered_content)

    logger.info(f"Generated Html report: {report_path}")

    return report_path

