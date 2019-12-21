import sys

import sentry_sdk

sentry_sdk.init("https://cc6dd86fbe9f4e7fbd95248cfcff114d@sentry.io/1862849")


if __name__ == "__main__":
    from httprunner.cli import main
    sys.exit(main())
