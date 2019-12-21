import os
import sys

from httprunner import exceptions, logger

project_working_directory = None


def locate_file(start_path, file_name):
    """ locate filename and return absolute file path.
        searching will be recursive upward until current working directory.

    Args:
        file_name (str): target locate file name
        start_path (str): start locating path, maybe file path or directory path

    Returns:
        str: located file path. None if file not found.

    Raises:
        exceptions.FileNotFound: If failed to locate file.

    """
    if os.path.isfile(start_path):
        start_dir_path = os.path.dirname(start_path)
    elif os.path.isdir(start_path):
        start_dir_path = start_path
    else:
        raise exceptions.FileNotFound("invalid path: {}".format(start_path))

    file_path = os.path.join(start_dir_path, file_name)
    if os.path.isfile(file_path):
        return os.path.abspath(file_path)

    # current working directory
    if os.path.abspath(start_dir_path) in [os.getcwd(), os.path.abspath(os.sep)]:
        raise exceptions.FileNotFound("{} not found in {}".format(file_name, start_path))

    # locate recursive upward
    return locate_file(os.path.dirname(start_dir_path), file_name)


def locate_debugtalk_py(start_path):
    """ locate debugtalk.py file

    Args:
        start_path (str): start locating path,
            maybe testcase file path or directory path

    Returns:
        str: debugtalk.py file path, None if not found

    """
    try:
        # locate debugtalk.py file.
        debugtalk_path = locate_file(start_path, "debugtalk.py")
    except exceptions.FileNotFound:
        debugtalk_path = None

    return debugtalk_path


def init_project_working_directory(test_path):
    """ this should be called at startup
        init_project_working_directory <- load_project_data <- load_cases <- run

    Args:
        test_path: specified testfile path

    Returns:
        (str, str): debugtalk.py path, project_working_directory

    """

    def prepare_path(path):
        if not os.path.exists(path):
            err_msg = "path not exist: {}".format(path)
            logger.log_error(err_msg)
            raise exceptions.FileNotFound(err_msg)

        if not os.path.isabs(path):
            path = os.path.join(os.getcwd(), path)

        return path

    test_path = prepare_path(test_path)

    # locate debugtalk.py file
    debugtalk_path = locate_debugtalk_py(test_path)

    global project_working_directory
    if debugtalk_path:
        # The folder contains debugtalk.py will be treated as PWD.
        project_working_directory = os.path.dirname(debugtalk_path)
    else:
        # debugtalk.py not found, use os.getcwd() as PWD.
        project_working_directory = os.getcwd()

    # add PWD to sys.path
    sys.path.insert(0, project_working_directory)

    return debugtalk_path, project_working_directory


def get_project_working_directory():
    global project_working_directory
    if project_working_directory is None:
        raise exceptions.MyBaseFailure("loader.load_cases() has not been called!")

    return project_working_directory
