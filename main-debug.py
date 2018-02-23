import sys

from httprunner.cli import main_hrun, main_locust

cmd = sys.argv.pop(1)

if cmd in ["hrun", "httprunner", "ate"]:
    main_hrun()
elif cmd in ["locust", "locusts"]:
    main_locust()
