# encoding: utf-8

import argparse
import multiprocessing
import os
import sys
import unittest

from httprunner import logger
from httprunner.__about__ import __description__, __version__
from httprunner.compat import is_py2
from httprunner.api import HttpRunner
from httprunner.utils import (create_scaffold, get_python2_retire_msg,
                              prettify_json_file, validate_json_file)


def main_hrun():
    """ API test: parse command line options and run commands.
    """
    parser = argparse.ArgumentParser(description=__description__)
    parser.add_argument(
        '-V', '--version', dest='version', action='store_true',
        help="show version")
    parser.add_argument(
        'testset_paths', nargs='*',
        help="testset file path")
    parser.add_argument(
        '--no-html-report', action='store_true', default=False,
        help="do not generate html report.")
    parser.add_argument(
        '--html-report-name',
        help="specify html report name, only effective when generating html report.")
    parser.add_argument(
        '--html-report-template',
        help="specify html report template path.")
    parser.add_argument(
        '--log-level', default='INFO',
        help="Specify logging level, default is INFO.")
    parser.add_argument(
        '--log-file',
        help="Write logs to specified file path.")
    parser.add_argument(
        '--dot-env-path',
        help="Specify .env file path, which is useful for keeping sensitive data.")
    parser.add_argument(
        '--failfast', action='store_true', default=False,
        help="Stop the test run on the first error or failure.")
    parser.add_argument(
        '--startproject',
        help="Specify new project name.")
    parser.add_argument(
        '--validate', nargs='*',
        help="Validate JSON testset format.")
    parser.add_argument(
        '--prettify', nargs='*',
        help="Prettify JSON testset format.")

    args = parser.parse_args()
    logger.setup_logger(args.log_level, args.log_file)

    if is_py2:
        logger.log_warning(get_python2_retire_msg())

    if args.version:
        logger.color_print("{}".format(__version__), "GREEN")
        exit(0)

    if args.validate:
        validate_json_file(args.validate)
        exit(0)
    if args.prettify:
        prettify_json_file(args.prettify)
        exit(0)

    project_name = args.startproject
    if project_name:
        project_path = os.path.join(os.getcwd(), project_name)
        create_scaffold(project_path)
        exit(0)

    runner = HttpRunner(failfast=args.failfast, dot_env_path=args.dot_env_path).run(args.testset_paths)

    if not args.no_html_report:
        runner.gen_html_report(
            html_report_name=args.html_report_name,
            html_report_template=args.html_report_template
        )

    summary = runner.summary
    return 0 if summary["success"] else 1

def main_locust():
    """ Performance test with locust: parse command line options and run commands.
    """
    logger.setup_logger("INFO")

    try:
        from httprunner import locusts
    except ImportError:
        msg = "Locust is not installed, install first and try again.\n"
        msg += "install command: pip install locustio"
        logger.log_warning(msg)
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
        logger.log_error("Testcase file is not specified, exit.")
        sys.exit(1)

    testcase_file_path = sys.argv[testcase_index]
    sys.argv[testcase_index] = locusts.parse_locustfile(testcase_file_path)

    if "--processes" in sys.argv:
        """ locusts -f locustfile.py --processes 4
        """
        if "--no-web" in sys.argv:
            logger.log_error("conflict parameter args: --processes & --no-web. \nexit.")
            sys.exit(1)

        processes_index = sys.argv.index('--processes')

        processes_count_index = processes_index + 1

        if processes_count_index >= len(sys.argv):
            """ do not specify processes count explicitly
                locusts -f locustfile.py --processes
            """
            processes_count = multiprocessing.cpu_count()
            logger.log_warning("processes count not specified, use {} by default.".format(processes_count))
        else:
            try:
                """ locusts -f locustfile.py --processes 4 """
                processes_count = int(sys.argv[processes_count_index])
                sys.argv.pop(processes_count_index)
            except ValueError:
                """ locusts -f locustfile.py --processes -P 8888 """
                processes_count = multiprocessing.cpu_count()
                logger.log_warning("processes count not specified, use {} by default.".format(processes_count))

        sys.argv.pop(processes_index)
        locusts.run_locusts_with_processes(sys.argv, processes_count)
    else:
        locusts.main()
