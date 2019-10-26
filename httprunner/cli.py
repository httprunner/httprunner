import argparse
import sys

from httprunner import __description__, __version__
from httprunner.api import HttpRunner
from httprunner.compat import is_py2
from httprunner.logger import color_print
from httprunner.utils import (create_scaffold, get_python2_retire_msg,
                              prettify_json_file)
from httprunner.validator import validate_json_file


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
        'testcase_paths', nargs='*',
        help="testcase file path")
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
        help="specify report template path.")
    parser.add_argument(
        '--report-dir',
        help="specify report save directory.")
    parser.add_argument(
        '--report-file',
        help="specify report file name.")
    parser.add_argument(
        '--failfast', action='store_true', default=False,
        help="Stop the test run on the first error or failure.")
    parser.add_argument(
        '--save-tests', action='store_true', default=False,
        help="Save loaded tests and parsed tests to JSON file.")
    parser.add_argument(
        '--startproject',
        help="Specify new project name.")
    parser.add_argument(
        '--validate', nargs='*',
        help="Validate JSON testcase format.")
    parser.add_argument(
        '--prettify', nargs='*',
        help="Prettify JSON testcase format.")

    args = parser.parse_args()

    if len(sys.argv) == 1:
        # no argument passed
        parser.print_help()
        return 0

    if args.version:
        color_print("{}".format(__version__), "GREEN")
        return 0

    if args.validate:
        validate_json_file(args.validate)
        return 0
    if args.prettify:
        prettify_json_file(args.prettify)
        return 0

    project_name = args.startproject
    if project_name:
        create_scaffold(project_name)
        return 0

    runner = HttpRunner(
        failfast=args.failfast,
        save_tests=args.save_tests,
        report_template=args.report_template,
        report_dir=args.report_dir,
        log_level=args.log_level,
        log_file=args.log_file,
        report_file=args.report_file
    )

    try:
        for path in args.testcase_paths:
            runner.run(path, dot_env_path=args.dot_env_path)
    except Exception:
        color_print("!!!!!!!!!! exception stage: {} !!!!!!!!!!".format(runner.exception_stage), "YELLOW")
        raise

    if runner.summary and runner.summary["success"]:
        return 0
    else:
        return 1


if __name__ == '__main__':
    sys.exit(main())
