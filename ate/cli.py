import argparse
import logging
import os
import sys
from collections import OrderedDict

from ate import __version__, exception
from ate.task import TaskSuite
from ate.utils import create_scaffold

import PyUnitReport


def main_ate():
    """ API test: parse command line options and run commands.
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
    parser.add_argument(
        '--failfast', action='store_true', default=False,
        help="Stop the test run on the first error or failure.")
    parser.add_argument(
        '--startproject',
        help="Specify new project name.")

    args = parser.parse_args()

    if args.version:
        print("HttpRunner version: {}".format(__version__))
        exit(0)

    log_level = getattr(logging, args.log_level.upper())
    logging.basicConfig(level=log_level)

    project_name = args.startproject
    if project_name:
        project_path = os.path.join(os.getcwd(), project_name)
        create_scaffold(project_path)
        exit(0)

    report_name = args.report_name
    if report_name and len(args.testset_paths) > 1:
        report_name = None
        logging.warning("More than one testset paths specified, \
                        report name is ignored, use generated time instead.")

    results = {}
    success = True

    for testset_path in set(args.testset_paths):

        testset_path = testset_path.rstrip('/')

        try:
            task_suite = TaskSuite(testset_path)
        except exception.TestcaseNotFound:
            success = False
            continue

        output_folder_name = os.path.basename(os.path.splitext(testset_path)[0])
        kwargs = {
            "output": output_folder_name,
            "report_name": report_name,
            "failfast": args.failfast
        }
        result = PyUnitReport.HTMLTestRunner(**kwargs).run(task_suite)
        results[testset_path] = OrderedDict({
            "total": result.testsRun,
            "successes": len(result.successes),
            "failures": len(result.failures),
            "errors": len(result.errors),
            "skipped": len(result.skipped)
        })

        if len(result.successes) != result.testsRun:
            success = False

        for task in task_suite.tasks:
            task.print_output()

    return 0 if success is True else 1

def main_locust():
    """ Performance test with locust: parse command line options and run commands.
    """
    try:
        from ate import locusts
    except ImportError:
        print("Locust is not installed, exit.")
        exit(1)

    sys.argv[0] = 'locust'
    if len(sys.argv) == 1:
        sys.argv.extend(["-h"])

    if sys.argv[1] in ["-h", "--help", "-V", "--version"]:
        locusts.main()
        sys.exit(0)

    try:
        testcase_index = sys.argv.index('-f') + 1
        assert testcase_index < len(sys.argv)
    except (ValueError, AssertionError):
        print("Testcase file is not specified, exit.")
        sys.exit(1)

    testcase_file_path = sys.argv[testcase_index]
    sys.argv[testcase_index] = locusts.parse_locustfile(testcase_file_path)

    if "--full-speed" in sys.argv:
        locusts.run_locusts_at_full_speed(sys.argv)
    else:
        locusts.main()
