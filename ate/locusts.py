import multiprocessing
import sys

from locust.main import main


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
