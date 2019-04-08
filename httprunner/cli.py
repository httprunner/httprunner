# encoding: utf-8

def main_hrun():
    """ API test: parse command line options and run commands.
    """
    import argparse
    from httprunner import logger
    from httprunner.__about__ import __description__, __version__
    from httprunner.api import HttpRunner
    from httprunner.compat import is_py2
    from httprunner.validator import validate_json_file
    from httprunner.utils import (create_scaffold, get_python2_retire_msg,
                                prettify_json_file)

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
        create_scaffold(project_name)
        exit(0)

    runner = HttpRunner(
        failfast=args.failfast,
        save_tests=args.save_tests,
        report_template=args.report_template,
        report_dir=args.report_dir
    )
    try:
        for path in args.testcase_paths:
            runner.run(path, dot_env_path=args.dot_env_path)
    except Exception:
        logger.log_error("!!!!!!!!!! exception stage: {} !!!!!!!!!!".format(runner.exception_stage))
        raise

    return 0


def main_locust():
    """ Performance test with locust: parse command line options and run commands.
    """
    # monkey patch ssl at beginning to avoid RecursionError when running locust.
    from gevent import monkey; monkey.patch_ssl()

    import multiprocessing
    import sys
    from httprunner import logger

    try:
        from httprunner import locusts
    except ImportError:
        msg = "Locust is not installed, install first and try again.\n"
        msg += "install command: pip install locustio"
        print(msg)
        exit(1)

    sys.argv[0] = 'locust'
    if len(sys.argv) == 1:
        sys.argv.extend(["-h"])

    if sys.argv[1] in ["-h", "--help", "-V", "--version"]:
        locusts.start_locust_main()
        sys.exit(0)

    # set logging level
    if "-L" in sys.argv:
        loglevel_index = sys.argv.index('-L') + 1
    elif "--loglevel" in sys.argv:
        loglevel_index = sys.argv.index('--loglevel') + 1
    else:
        loglevel_index = None

    if loglevel_index and loglevel_index < len(sys.argv):
        loglevel = sys.argv[loglevel_index]
    else:
        # default
        loglevel = "WARNING"

    logger.setup_logger(loglevel)

    # get testcase file path
    try:
        if "-f" in sys.argv:
            testcase_index = sys.argv.index('-f') + 1
        elif "--locustfile" in sys.argv:
            testcase_index = sys.argv.index('--locustfile') + 1
        else:
            testcase_index = None

        assert testcase_index and testcase_index < len(sys.argv)
    except AssertionError:
        print("Testcase file is not specified, exit.")
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
        locusts.start_locust_main()
