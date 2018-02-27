import argparse
import multiprocessing
import os
import sys
import unittest

from httprunner import logger
from httprunner.__about__ import __version__
from httprunner.task import HttpRunner
from httprunner.utils import (create_scaffold, load_dot_env_file, print_output,
                              string_type)


def main_hrun():
    """ API test: parse command line options and run commands.
    """
    parser = argparse.ArgumentParser(
        description='HTTP test runner, not just about api test and load test.')
    parser.add_argument(
        '-V', '--version', dest='version', action='store_true',
        help="show version")
    parser.add_argument(
        'testset_paths', nargs='*',
        help="testset file path")
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
        '--dot-env-path',
        help="Specify .env file path, which is useful for keeping production credentials.")
    parser.add_argument(
        '--failfast', action='store_true', default=False,
        help="Stop the test run on the first error or failure.")
    parser.add_argument(
        '--startproject',
        help="Specify new project name.")

    args = parser.parse_args()
    logger.setup_logger(args.log_level)

    if args.version:
        logger.color_print("{}".format(__version__), "GREEN")
        exit(0)

    dot_env_path = args.dot_env_path or os.path.join(os.getcwd(), ".env")
    if dot_env_path:
        load_dot_env_file(dot_env_path)

    project_name = args.startproject
    if project_name:
        project_path = os.path.join(os.getcwd(), project_name)
        create_scaffold(project_path)
        exit(0)

    result = HttpRunner(args.testset_paths, failfast=args.failfast).run(
        html_report_name=args.html_report_name,
        html_report_template=args.html_report_template
    )

    print_output(result["output"])
    return 0 if result["success"] else 1

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

    if "--cpu-cores" in sys.argv:
        """ locusts -f locustfile.py --cpu-cores 4
        """
        if "--no-web" in sys.argv:
            logger.log_error("conflict parameter args: --cpu-cores & --no-web. \nexit.")
            sys.exit(1)

        cpu_cores_index = sys.argv.index('--cpu-cores')

        cpu_cores_num_index = cpu_cores_index + 1

        if cpu_cores_num_index >= len(sys.argv):
            """ do not specify cpu cores explicitly
                locusts -f locustfile.py --cpu-cores
            """
            cpu_cores_num_value = multiprocessing.cpu_count()
            logger.log_warning("cpu cores number not specified, use {} by default.".format(cpu_cores_num_value))
        else:
            try:
                """ locusts -f locustfile.py --cpu-cores 4 """
                cpu_cores_num_value = int(sys.argv[cpu_cores_num_index])
                sys.argv.pop(cpu_cores_num_index)
            except ValueError:
                """ locusts -f locustfile.py --cpu-cores -P 8888 """
                cpu_cores_num_value = multiprocessing.cpu_count()
                logger.log_warning("cpu cores number not specified, use {} by default.".format(cpu_cores_num_value))

        sys.argv.pop(cpu_cores_index)
        locusts.run_locusts_on_cpu_cores(sys.argv, cpu_cores_num_value)
    else:
        locusts.main()
