import codecs
import multiprocessing
import os
import sys

from httprunner.testcase import load_test_file
from locust.main import main


def parse_locustfile(file_path):
    """ parse testcase file and return locustfile path.
        if file_path is a Python file, assume it is a locustfile
        if file_path is a YAML/JSON file, convert it to locustfile
    """
    if not os.path.isfile(file_path):
        print("file path invalid, exit.")
        sys.exit(1)

    file_suffix = os.path.splitext(file_path)[1]
    if file_suffix == ".py":
        locustfile_path = file_path
    elif file_suffix in ['.yaml', '.yml', '.json']:
        locustfile_path = gen_locustfile(file_path)
    else:
        # '' or other suffix
        print("file type should be YAML/JSON/Python, exit.")
        sys.exit(1)

    return locustfile_path

def gen_locustfile(testcase_file_path):
    """ generate locustfile from template.
    """
    locustfile_path = 'locustfile.py'
    template_path = os.path.join(
        os.path.dirname(os.path.realpath(__file__)),
        'locustfile_template'
    )
    testset = load_test_file(testcase_file_path)
    host = testset.get("config", {}).get("request", {}).get("base_url", "")

    with codecs.open(template_path, encoding='utf-8') as template:
        with codecs.open(locustfile_path, 'w', encoding='utf-8') as locustfile:
            template_content = template.read()
            template_content = template_content.replace("$HOST", host)
            template_content = template_content.replace("$TESTCASE_FILE", testcase_file_path)
            locustfile.write(template_content)

    return locustfile_path

def start_master(sys_argv):
    sys_argv.append("--master")
    sys.argv = sys_argv
    main()

def start_slave(sys_argv):
    sys_argv.extend(["--slave"])
    sys.argv = sys_argv
    main()

def run_locusts_at_full_speed(sys_argv):
    sys_argv.pop(sys_argv.index("--full-speed"))
    slaves_num = multiprocessing.cpu_count()

    processes = []
    for _ in range(slaves_num):
        p_slave = multiprocessing.Process(target=start_slave, args=(sys_argv,))
        p_slave.daemon = True
        p_slave.start()
        processes.append(p_slave)

    try:
        start_master(sys_argv)
    except KeyboardInterrupt:
        sys.exit(0)
