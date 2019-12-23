__version__ = "2.4.6"
__description__ = "One-stop solution for HTTP(S) testing."

__all__ = ["__version__", "__description__"]

import sentry_sdk

sentry_sdk.init(
    dsn="https://cc6dd86fbe9f4e7fbd95248cfcff114d@sentry.io/1862849",
    release="httprunner@{}".format(__version__)
)
