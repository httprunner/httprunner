import sys

from httprunner.cli import main_hrun, main_locust
from httprunner.logger import color_print

cmd = sys.argv.pop(1)

if cmd in ["hrun", "httprunner", "ate"]:
    main_hrun()
elif cmd in ["locust", "locusts"]:
    main_locust()
else:
    color_print("Miss debugging type.", "RED")
    example = "\n".join([
        "e.g.",
        "python main-debug.py hrun /path/to/testset_file",
        "python main-debug.py locusts -f /path/to/testset_file"
    ])
    color_print(example, "yellow")
