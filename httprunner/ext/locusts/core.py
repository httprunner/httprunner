import io
import multiprocessing
import os
import sys

from loguru import logger


def parse_locustfile(file_path):
    """ parse testcase file and return locustfile path.
        if file_path is a Python file, assume it is a locustfile
        if file_path is a YAML/JSON file, convert it to locustfile
    """
    if not os.path.isfile(file_path):
        logger.error("file path invalid, exit.")
        sys.exit(1)

    file_suffix = os.path.splitext(file_path)[1]
    if file_suffix == ".py":
        locustfile_path = file_path
    elif file_suffix in ['.yaml', '.yml', '.json']:
        locustfile_path = gen_locustfile(file_path)
    else:
        # '' or other suffix
        logger.error("file type should be YAML/JSON/Python, exit.")
        sys.exit(1)

    return locustfile_path


def gen_locustfile(testcase_file_path):
    """ generate locustfile from template.
    """
    locustfile_path = 'locustfile.py'
    template_path = os.path.join(
        os.path.dirname(os.path.realpath(__file__)),
        "locustfile_template.py"
    )

    with io.open(template_path, encoding='utf-8') as template:
        with io.open(locustfile_path, 'w', encoding='utf-8') as locustfile:
            template_content = template.read()
            template_content = template_content.replace("$TESTCASE_FILE", testcase_file_path)
            locustfile.write(template_content)

    return locustfile_path


def start_locust_main():
    logger.info(f"run command: {sys.argv}")
    from locust.main import main
    main()


def start_master(sys_argv):
    sys_argv.append("--master")
    sys.argv = sys_argv
    start_locust_main()


def start_slave(sys_argv):
    if "--slave" not in sys_argv:
        sys_argv.extend(["--slave"])

    sys.argv = sys_argv
    start_locust_main()


def init_slave_processes(slave_num):
    """ init specified number of locust slave processes."""
    processes = []

    for _ in range(slave_num):
        p_slave = multiprocessing.Process(target=start_slave, args=(sys.argv,))
        p_slave.daemon = True
        p_slave.start()
        processes.append(p_slave)

    return processes


def start_slaves(slave_num):
    logger.info(f"Start {slave_num} locust slaves ...")
    processes = init_slave_processes(slave_num)
    [process.join() for process in processes]


def quick_run_locusts(slave_num):
    """ quick start locust master and multiple slaves.

    Args:
        slave_num: locust slaves number
    """
    logger.info(f"Start locust master with {slave_num} slaves ...")

    processes = init_slave_processes(slave_num)
    processes.append(
        multiprocessing.Process(target=start_master, args=(sys.argv,))
    )
    [process.join() for process in processes]
