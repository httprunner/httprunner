import argparse
import os
import sys

if len(sys.argv) >= 2 and sys.argv[1] == "locusts":
    # monkey patch ssl at beginning to avoid RecursionError when running locust.
    try:
        from gevent import monkey
        monkey.patch_ssl()
        from locust.main import main
    except ImportError:
        msg = """
Locust is not installed, install first and try again.
install with pip:
$ pip install locustio
"""
        print(msg)
        sys.exit(1)

from loguru import logger

from httprunner import __description__, __version__
from httprunner.api import HttpRunner
from httprunner.ext.har2case import init_har2case_parser, main_har2case
from httprunner.ext.scaffold import init_parser_scaffold, main_scaffold
from httprunner.ext.locusts import init_parser_locusts, main_locusts
from httprunner.report import gen_html_report


def init_parser_run(subparsers):
    sub_parser_run = subparsers.add_parser(
        "run", help="Run HttpRunner testcases.")

    sub_parser_run.add_argument(
        'testfile_paths', nargs='*',
        help="Specify api/testcase/testsuite file paths to run.")
    sub_parser_run.add_argument(
        '--log-level', default='INFO',
        help="Specify logging level, default is INFO.")
    sub_parser_run.add_argument(
        '--log-file',
        help="Write logs to specified file path.")
    sub_parser_run.add_argument(
        '--dot-env-path',
        help="Specify .env file path, which is useful for keeping sensitive data.")
    sub_parser_run.add_argument(
        '--report-template',
        help="Specify report template path.")
    sub_parser_run.add_argument(
        '--report-dir',
        help="Specify report save directory.")
    sub_parser_run.add_argument(
        '--report-file',
        help="Specify report file path, this has higher priority than specifying report dir.")
    sub_parser_run.add_argument(
        '--save-tests', action='store_true', default=False,
        help="Save loaded/parsed/vars_out/summary json data to JSON files.")
    sub_parser_run.add_argument(
        '--failfast', action='store_true', default=False,
        help="Stop the test run on the first error or failure.")

    return sub_parser_run


def main_run(args):
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
            report_dir = args.report_dir or os.path.join(os.getcwd(), "reports")
            gen_html_report(
                summary,
                report_template=args.report_template,
                report_dir=report_dir,
                report_file=args.report_file
            )
            err_code |= (0 if summary and summary["success"] else 1)
    except Exception as ex:
        logger.error(f"!!!!!!!!!! exception stage: {runner.exception_stage} !!!!!!!!!!\n{str(ex)}")
        err_code = 1

    sys.exit(err_code)


def main():
    """ API test: parse command line options and run commands.
    """
    parser = argparse.ArgumentParser(description=__description__)
    parser.add_argument(
        '-V', '--version', dest='version', action='store_true',
        help="show version")

    subparsers = parser.add_subparsers(help='sub-command help')
    sub_parser_run = init_parser_run(subparsers)
    sub_parser_scaffold = init_parser_scaffold(subparsers)
    sub_parser_har2case = init_har2case_parser(subparsers)
    sub_parser_locusts = init_parser_locusts(subparsers)

    extra_args = []
    if len(sys.argv) >= 2 and sys.argv[1] == "locusts":
        args, extra_args = parser.parse_known_args()
    else:
        args = parser.parse_args()

    if args.version:
        print(f"{__version__}")
        sys.exit(0)

    if len(sys.argv) == 1:
        # hrun
        parser.print_help()
        sys.exit(0)

    elif sys.argv[1] == "run":
        # hrun run
        if len(sys.argv) == 2:
            sub_parser_run.print_help()
            sys.exit(0)

        main_run(args)

    elif sys.argv[1] == "startproject":
        # hrun startproject
        if len(sys.argv) == 2:
            sub_parser_scaffold.print_help()
            sys.exit(0)

        main_scaffold(args)

    elif sys.argv[1] == "har2case":
        # hrun har2case
        if len(sys.argv) == 2:
            sub_parser_har2case.print_help()
            sys.exit(0)

        main_har2case(args)

    elif sys.argv[1] == "locusts":
        # hrun locusts
        if len(sys.argv) == 2:
            sub_parser_locusts.print_help()
            sys.exit(0)

        main_locusts(args, extra_args)


def main_hrun_alias():
    """ command alias
        hrun = httprunner run
    """
    sys.argv.insert(1, "run")
    main()


if __name__ == '__main__':
    main()
