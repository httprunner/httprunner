import os


def is_test_path(path):
    """ check if path is valid json/yaml file path or a existed directory.

    Args:
        path (str/list/tuple): file path/directory or file path list.

    Returns:
        bool: True if path is valid file path or path list, otherwise False.

    """
    if not isinstance(path, (str, list, tuple)):
        return False

    elif isinstance(path, (list, tuple)):
        for p in path:
            if not is_test_path(p):
                return False

        return True

    else:
        # path is string
        if not os.path.exists(path):
            return False

        # path exists
        if os.path.isfile(path):
            # path is a file
            file_suffix = os.path.splitext(path)[1].lower()
            if file_suffix not in [".json", ".yaml", ".yml"]:
                # path is not json/yaml file
                return False
            else:
                return True
        elif os.path.isdir(path):
            # path is a directory
            return True
        else:
            # path is neither a folder nor a file, maybe a symbol link or something else
            return False
