# encoding: utf-8

"""
httprunner.compat
~~~~~~~~~~~~~~~~~

This module handles import compatibility issues between Python 2 and
Python 3.
"""

try:
    import simplejson as json
except ImportError:
    import json

import sys

# -------
# Pythons
# -------

# Syntax sugar.
_ver = sys.version_info

#: Python 2.x?
is_py2 = _ver[0] == 2

#: Python 3.x?
is_py3 = _ver[0] == 3


# ---------
# Specifics
# ---------

try:
    JSONDecodeError = json.JSONDecodeError
except AttributeError:
    JSONDecodeError = ValueError

if is_py2:
    builtin_str = str
    bytes = str
    str = unicode
    basestring = basestring
    numeric_types = (int, long, float)
    integer_types = (int, long)

    FileNotFoundError = IOError
    import StringIO as io

elif is_py3:
    builtin_str = str
    str = str
    bytes = bytes
    basestring = (str, bytes)
    numeric_types = (int, float)
    integer_types = (int,)

    FileNotFoundError = FileNotFoundError
    import io as io
