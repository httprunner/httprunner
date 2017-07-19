import os
import argparse
import logging

import PyUnitReport

from ate.task import create_task

def main():
    """ parse command line options and run commands.
    """
    parser = argparse.ArgumentParser(
        description='Api Test Engine.')
    parser.add_argument(
        '--testcase-path', default='testcases',
        help="testcase file path")
    parser.add_argument(
        '--log-level', default='INFO',
        help="Specify logging level, default is INFO.")
    parser.add_argument(
        '--report-name',
        help="Specify report name, default is generated time.")

    args = parser.parse_args()

    log_level = getattr(logging, args.log_level.upper())
    logging.basicConfig(level=log_level)

    testcase_path = args.testcase_path.rstrip('/')
    task_suite = create_task(testcase_path)

    output_folder_name = os.path.basename(os.path.splitext(testcase_path)[0])
    kwargs = {
        "output": output_folder_name,
        "report_name": args.report_name
    }
    PyUnitReport.HTMLTestRunner(**kwargs).run(task_suite)
