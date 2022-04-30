import logging
import os
import sys

from colorama import Fore, init
from colorlog import ColoredFormatter

init(autoreset=True)

LOG_LEVEL = "INFO"
LOG_FILE_PATH = ""

log_colors_config = {
    "DEBUG": "cyan",
    "INFO": "green",
    "WARNING": "yellow",
    "ERROR": "red",
    "CRITICAL": "red",
}
loggers = {}


def setup_logger(log_level, log_file=None):
    global LOG_LEVEL
    LOG_LEVEL = log_level

    if log_file:
        global LOG_FILE_PATH
        LOG_FILE_PATH = log_file


def get_logger(name=None):
    """setup logger with ColoredFormatter."""
    name = name or "httprunner"
    logger_key = "".join([name, LOG_LEVEL, LOG_FILE_PATH])
    if logger_key in loggers:
        return loggers[logger_key]

    _logger = logging.getLogger(name)

    log_level = LOG_LEVEL
    level = getattr(logging, log_level.upper(), None)
    if not level:
        color_print("Invalid log level: %s" % log_level, "RED")
        sys.exit(1)

    # hide traceback when log level is INFO/WARNING/ERROR/CRITICAL
    if level >= logging.INFO:
        sys.tracebacklimit = 0

    _logger.setLevel(level)
    if LOG_FILE_PATH:
        log_dir = os.path.dirname(LOG_FILE_PATH)
        if not os.path.isdir(log_dir):
            os.makedirs(log_dir)
        handler = logging.FileHandler(LOG_FILE_PATH, encoding="utf-8")
    else:
        handler = logging.StreamHandler(sys.stdout)

    formatter = ColoredFormatter(
        "%(log_color)s%(bg_white)s%(levelname)-8s%(reset)s %(message)s",
        datefmt=None,
        reset=True,
        log_colors=log_colors_config,
    )
    handler.setFormatter(formatter)
    _logger.addHandler(handler)

    loggers[logger_key] = _logger
    return _logger


def coloring(text, color="WHITE"):
    fore_color = getattr(Fore, color.upper())
    return fore_color + text


def color_print(msg, color="WHITE"):
    fore_color = getattr(Fore, color.upper())
    print(fore_color + msg)


def log_with_color(level):
    """log with color by different level"""

    def wrapper(text):
        color = log_colors_config[level.upper()]
        _logger = get_logger()
        getattr(_logger, level.lower())(coloring(text, color))

    return wrapper


log_debug = log_with_color("debug")
log_info = log_with_color("info")
log_warning = log_with_color("warning")
log_error = log_with_color("error")
log_critical = log_with_color("critical")
