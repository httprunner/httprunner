""" Convert HAR (HTTP Archive) to YAML/JSON testcase for HttpRunner.

Usage:
    # convert to JSON format testcase
    $ hrun har2case demo.har

    # convert to YAML format testcase
    $ hrun har2case demo.har -2y

"""
import os
import sys

from loguru import logger

from httprunner.ext.har2case.core import HarParser


def init_har2case_parser(subparsers):
    """ HAR converter: parse command line options and run commands.
    """
    parser = subparsers.add_parser(
        "har2case",
        help="Convert HAR(HTTP Archive) to YAML/JSON testcases for HttpRunner.",
    )
    parser.add_argument("har_source_file", nargs="?", help="Specify HAR source file")
    parser.add_argument(
        "-2y",
        "--to-yml",
        "--to-yaml",
        dest="to_yaml",
        action="store_true",
        help="Convert to YAML format, if not specified, convert to JSON format by default.",
    )
    parser.add_argument(
        "--filter",
        help="Specify filter keyword, only url include filter string will be converted.",
    )
    parser.add_argument(
        "--exclude",
        help="Specify exclude keyword, url that includes exclude string will be ignored, "
        "multiple keywords can be joined with '|'",
    )

    return parser


def main_har2case(args):
    har_source_file = args.har_source_file
    if not har_source_file or not har_source_file.endswith(".har"):
        logger.error("HAR file not specified.")
        sys.exit(1)

    if not os.path.isfile(har_source_file):
        logger.error(f"HAR file not exists: {har_source_file}")
        sys.exit(1)

    output_file_type = "YML" if args.to_yaml else "JSON"
    HarParser(har_source_file, args.filter, args.exclude).gen_testcase(output_file_type)

    return 0
