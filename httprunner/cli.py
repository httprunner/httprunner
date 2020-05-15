import argparse
import os
import sys

import pytest

from httprunner import __description__, __version__, exceptions
from httprunner.ext.har2case import init_har2case_parser, main_har2case
from httprunner.ext.make import init_make_parser, main_make, convert_testcase_path
from httprunner.ext.scaffold import init_parser_scaffold, main_scaffold


def init_parser_run(subparsers):
    sub_parser_run = subparsers.add_parser(
        "run", help="Make HttpRunner testcases and run with pytest."
    )
    return sub_parser_run


def main_run(extra_args):
    tests_path_list = []
    for index, item in enumerate(extra_args):
        if not os.path.exists(item):
            # item is not file/folder path
            continue
        elif os.path.isfile(item):
            # replace YAML/JSON file path with generated python file
            extra_args[index] = convert_testcase_path(item)

        tests_path_list.append(item)

    if len(tests_path_list) == 0:
        # has not specified any testcase path
        raise exceptions.ParamsError("Missed testcase path")

    main_make(tests_path_list)

    if "-s" not in extra_args:
        extra_args.insert(0, "-s")
    pytest.main(extra_args)


def main():
    """ API test: parse command line options and run commands.
    """
    parser = argparse.ArgumentParser(description=__description__)
    parser.add_argument(
        "-V", "--version", dest="version", action="store_true", help="show version"
    )

    subparsers = parser.add_subparsers(help="sub-command help")
    sub_parser_run = init_parser_run(subparsers)
    sub_parser_scaffold = init_parser_scaffold(subparsers)
    sub_parser_har2case = init_har2case_parser(subparsers)
    sub_parser_make = init_make_parser(subparsers)

    if len(sys.argv) == 1:
        # httprunner
        parser.print_help()
        sys.exit(0)
    elif len(sys.argv) == 2:
        # print help for sub-commands
        if sys.argv[1] in ["-V", "--version"]:
            # httprunner -V
            print(f"{__version__}")
        elif sys.argv[1] in ["-h", "--help"]:
            # httprunner -h
            parser.print_help()
        elif sys.argv[1] == "startproject":
            # httprunner startproject
            sub_parser_scaffold.print_help()
        elif sys.argv[1] == "har2case":
            # httprunner har2case
            sub_parser_har2case.print_help()
        elif sys.argv[1] == "run":
            # httprunner run
            pytest.main(["-h"])
        elif sys.argv[1] == "make":
            # httprunner make
            sub_parser_make.print_help()
        sys.exit(0)
    elif (
        len(sys.argv) == 3 and sys.argv[1] == "run" and sys.argv[2] in ["-h", "--help"]
    ):
        # httprunner run -h
        pytest.main(["-h"])
        sys.exit(0)

    extra_args = []
    if len(sys.argv) >= 2 and sys.argv[1] in ["run", "locusts"]:
        args, extra_args = parser.parse_known_args()
    else:
        args = parser.parse_args()

    if args.version:
        print(f"{__version__}")
        sys.exit(0)

    if sys.argv[1] == "run":
        main_run(extra_args)
    elif sys.argv[1] == "startproject":
        main_scaffold(args)
    elif sys.argv[1] == "har2case":
        main_har2case(args)
    elif sys.argv[1] == "make":
        main_make(args.testcase_path)


def main_hrun_alias():
    """ command alias
        hrun = httprunner run
    """
    if len(sys.argv) == 2:
        if sys.argv[1] in ["-V", "--version"]:
            # hrun -V
            sys.argv = ["httprunner", "-V"]
        elif sys.argv[1] in ["-h", "--help"]:
            pytest.main(["-h"])
            sys.exit(0)
        else:
            # hrun /path/to/testcase
            sys.argv.insert(1, "run")
    else:
        sys.argv.insert(1, "run")

    main()


def main_make_alias():
    """ command alias
        hmake = httprunner make
    """
    sys.argv.insert(1, "make")
    main()


def main_har2case_alias():
    """ command alias
        har2case = httprunner har2case
    """
    sys.argv.insert(1, "har2case")
    main()


if __name__ == "__main__":
    main()
