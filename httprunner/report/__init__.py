"""
HttpRunner report

- summarize: aggregate test stat data to summary
- stringify: stringify summary, in order to dump json file and generate html report.
- html: render html report
"""

from httprunner.report.summarize import get_platform, aggregate_stat, get_summary
from httprunner.report.stringify import stringify_summary
from httprunner.report.html import HtmlTestResult, gen_html_report

__all__ = [
    "get_platform",
    "aggregate_stat",
    "get_summary",
    "stringify_summary",
    "HtmlTestResult",
    "gen_html_report"
]
