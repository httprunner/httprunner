import argparse
import os
import sys

import sentry_sdk

from httprunner import __description__, __version__, exceptions
from httprunner.api import HttpRunner
from httprunner.compat import is_py2
from httprunner.loader import load_cases
from httprunner.logger import color_print, log_error
from httprunner.report import gen_html_report
from httprunner.utils import (create_scaffold, get_python2_retire_msg,
                              prettify_json_file, init_sentry_sdk)

init_sentry_sdk()


def main():
    """ API test: parse command line options and run commands.
    """
    if is_py2:
        color_print(get_python2_retire_msg(), "YELLOW")

    parser = argparse.ArgumentParser(description=__description__)
    parser.add_argument(
        '-V', '--version', dest='version', action='store_true',
        help="show version")
    parser.add_argument(
        'testfile_paths', nargs='*',
        help="Specify api/testcase/testsuite file paths to run.")
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
        '--report-template',
        help="Specify report template path.")
    parser.add_argument(
        '--report-dir',
        help="Specify report save directory.")
    parser.add_argument(
        '--report-file',
        help="Specify report file path, this has higher priority than specifying report dir.")
    parser.add_argument(
        '--save-tests', action='store_true', default=False,
        help="Save loaded/parsed/vars_out/summary json data to JSON files.")
    parser.add_argument(
        '--failfast', action='store_true', default=False,
        help="Stop the test run on the first error or failure.")
    parser.add_argument(
        '--startproject',
        help="Specify new project name.")
    parser.add_argument(
        '--validate', nargs='*',
        help="Validate YAML/JSON api/testcase/testsuite format.")
    parser.add_argument(
        '--prettify', nargs='*',
        help="Prettify JSON testcase format.")

    args = parser.parse_args()

    if len(sys.argv) == 1:
        # no argument passed
        parser.print_help()
        sys.exit(0)

    if args.version:
        color_print("{}".format(__version__), "GREEN")
        sys.exit(0)

    if args.validate:
        for validate_path in args.validate:
            try:
                color_print("validate test file: {}".format(validate_path), "GREEN")
                load_cases(validate_path, args.dot_env_path)
            except exceptions.MyBaseError as ex:
                log_error(str(ex))
                continue

        color_print("done!", "BLUE")
        sys.exit(0)

    if args.prettify:
        prettify_json_file(args.prettify)
        sys.exit(0)

    project_name = args.startproject
    if project_name:
        create_scaffold(project_name)
        sys.exit(0)

    runner = HttpRunner(
        failfast=args.failfast,
        save_tests=args.save_tests,
        log_level=args.log_level,
        log_file=args.log_file
    )

    err_code = 0
    try:
        for path in args.testfile_paths:
            summary = runner.run(path, dot_env_path=args.dot_env_path)
            report_dir = args.report_dir or os.path.join(runner.project_working_directory, "reports")
            gen_html_report(
                summary,
                report_template=args.report_template,
                report_dir=report_dir,
                report_file=args.report_file
            )
            err_code |= (0 if summary and summary["success"] else 1)
    except Exception as ex:
        color_print("!!!!!!!!!! exception stage: {} !!!!!!!!!!".format(runner.exception_stage), "YELLOW")
        color_print(str(ex), "RED")
        sentry_sdk.capture_exception(ex)
        err_code = 1

    sys.exit(err_code)


if __name__ == '__main__':
    main()
