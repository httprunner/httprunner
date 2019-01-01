# encoding: utf-8

import io
import multiprocessing
import os
import sys

from httprunner.logger import color_print
from httprunner import loader


def parse_locustfile(file_path):
    """ parse testcase file and return locustfile path.
        if file_path is a Python file, assume it is a locustfile
        if file_path is a YAML/JSON file, convert it to locustfile
    """
    if not os.path.isfile(file_path):
        color_print("file path invalid, exit.", "RED")
        sys.exit(1)

    file_suffix = os.path.splitext(file_path)[1]
    if file_suffix == ".py":
        locustfile_path = file_path
    elif file_suffix in ['.yaml', '.yml', '.json']:
        locustfile_path = gen_locustfile(file_path)
    else:
        # '' or other suffix
        color_print("file type should be YAML/JSON/Python, exit.", "RED")
        sys.exit(1)

    return locustfile_path


def gen_locustfile(testcase_file_path):
    """ generate locustfile from template.
    """
    locustfile_path = 'locustfile.py'
    template_path = os.path.join(
        os.path.dirname(os.path.realpath(__file__)),
        "templates",
        "locustfile_template"
    )

    with io.open(template_path, encoding='utf-8') as template:
        with io.open(locustfile_path, 'w', encoding='utf-8') as locustfile:
            template_content = template.read()
            template_content = template_content.replace("$TESTCASE_FILE", testcase_file_path)
            locustfile.write(template_content)

    return locustfile_path


def start_locust_main():
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


def run_locusts_with_processes(sys_argv, processes_count):
    processes = []
    manager = multiprocessing.Manager()

    for _ in range(processes_count):
        p_slave = multiprocessing.Process(target=start_slave, args=(sys_argv,))
        p_slave.daemon = True
        p_slave.start()
        processes.append(p_slave)

    try:
        if "--slave" in sys_argv:
            [process.join() for process in processes]
        else:
            start_master(sys_argv)
    except KeyboardInterrupt:
        manager.shutdown()
