import os
import argparse
import logging

import PyUnitReport

from ate import __version__
from ate.task import create_task

def main():
    """ parse command line options and run commands.
    """
    parser = argparse.ArgumentParser(
        description='Api Test Engine.')
    parser.add_argument(
        '-V', '--version', dest='version', action='store_true',
        help="show version")
    parser.add_argument(
        'testset_paths', nargs='*',
        help="testset file path")
    parser.add_argument(
        '--log-level', default='INFO',
        help="Specify logging level, default is INFO.")
    parser.add_argument(
        '--report-name',
        help="Specify report name, default is generated time.")

    args = parser.parse_args()

    if args.version:
        print(__version__)
        exit(0)

    log_level = getattr(logging, args.log_level.upper())
    logging.basicConfig(level=log_level)

    report_name = args.report_name
    if report_name and len(args.testset_paths) > 1:
        report_name = None
        logging.warning("More than one testset paths specified, \
                        report name is ignored, use generated time instead.")

    for testset_path in args.testset_paths:

        testset_path = testset_path.strip('/')
        task_suite = create_task(testset_path)

        output_folder_name = os.path.basename(os.path.splitext(testset_path)[0])
        kwargs = {
            "output": output_folder_name,
            "report_name": report_name,
            "failfast": True
        }
        PyUnitReport.HTMLTestRunner(**kwargs).run(task_suite)
