import multiprocessing
import sys

from loguru import logger

from httprunner import __version__
from httprunner.ext.locusts.core import start_locust_main, parse_locustfile, quick_run_locusts, start_master, \
    start_slaves

CPU_COUNT = multiprocessing.cpu_count()


def init_parser_locusts(subparsers):
    sub_parser_locusts = subparsers.add_parser(
        "locusts", help="Run load test with locust.")
    sub_parser_locusts.add_argument(
        '--locust-help', action='store_true', default=False,
        help="Show locust help.")
    sub_parser_locusts.add_argument('test_file', nargs='?',
                                    help="Specify YAML/JSON testcase file.")
    sub_parser_locusts.add_argument(
        "--master", action='store_true', default=False, help="Start locust master.")
    sub_parser_locusts.add_argument(
        "--slaves", type=int, help="Specify locust slave number.")
    sub_parser_locusts.add_argument(
        "--quickstart", action='store_true', default=False,
        help=f"Start locust master with {CPU_COUNT} slaves.")
    return sub_parser_locusts


def main_locusts(args, extra_args):
    """ Performance test with locust: parse command line options and run commands.
    """
    logger.info(f"HttpRunner version: {__version__}")
    sys.argv = ["locust", *extra_args]

    if args.locust_help:
        sys.argv = ["locust", "-h"]
        start_locust_main()

    def get_arg_index(*target_args):
        for arg in target_args:
            if arg not in sys.argv:
                continue

            return sys.argv.index(arg) + 1

        return None

    # set logging level
    loglevel_index = get_arg_index("-L", "--loglevel")
    if loglevel_index and loglevel_index < len(sys.argv):
        loglevel = sys.argv[loglevel_index]
        loglevel = loglevel.upper()
    else:
        # default
        loglevel = "INFO"

    logger.remove()
    logger.add(sys.stdout, level=loglevel)

    if not args.test_file:
        logger.error("Testcase file is not specified, exit.")
        sys.exit(1)

    # convert httprunner yaml/json case to locustfile.py
    locustfile_path = parse_locustfile(args.test_file)
    sys.argv.extend(["-f", locustfile_path])

    manager = multiprocessing.Manager()
    try:
        if args.quickstart:
            quick_run_locusts(CPU_COUNT)
        elif args.master:
            start_master(sys.argv)
        elif args.slaves:
            start_slaves(args.slaves)
        else:
            quick_run_locusts(CPU_COUNT)

    except KeyboardInterrupt:
        manager.shutdown()
