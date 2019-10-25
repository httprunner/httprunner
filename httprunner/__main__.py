import os
import sys

from httprunner.cli import main_hrun, main_locust

if __name__ == "__main__":
    """ debugging mode
    """
    if len(sys.argv) == 0:
        sys.exit(1)

    sys.path.insert(0, os.getcwd())
    cmd = sys.argv.pop(1)

    if cmd in ["hrun", "httprunner", "ate"]:
        main_hrun()
    elif cmd in ["locust", "locusts"]:
        main_locust()
    else:
        from httprunner.logger import color_print
        color_print("Miss debugging type.", "RED")
        example = "\n".join([
            "e.g.",
            "python -m httprunner hrun /path/to/testcase_file",
            "python -m httprunner locusts -f /path/to/testcase_file"
        ])
        color_print(example, "yellow")
