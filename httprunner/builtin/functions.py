"""
Built-in functions used in YAML/JSON testcases.
"""

import datetime
import os
import random
import string
import time

import filetype
from requests_toolbelt import MultipartEncoder

from httprunner.compat import builtin_str, integer_types
from httprunner.exceptions import ParamsError

PWD = os.getcwd()


def gen_random_string(str_len):
    """ generate random string with specified length
    """
    return ''.join(
        random.choice(string.ascii_letters + string.digits) for _ in range(str_len))


def get_timestamp(str_len=13):
    """ get timestamp string, length can only between 0 and 16
    """
    if isinstance(str_len, integer_types) and 0 < str_len < 17:
        return builtin_str(time.time()).replace(".", "")[:str_len]

    raise ParamsError("timestamp length can only between 0 and 16.")


def get_current_date(fmt="%Y-%m-%d"):
    """ get current date, default format is %Y-%m-%d
    """
    return datetime.datetime.now().strftime(fmt)


def sleep(n_secs):
    """ sleep n seconds
    """
    time.sleep(n_secs)


"""
upload files with requests-toolbelt
e.g.

    - test:
        name: upload file
        variables:
            file_path: "data/test.env"
            multipart_encoder: ${multipart_encoder(file=$file_path)}
        request:
            url: /post
            method: POST
            headers:
                Content-Type: ${multipart_content_type($multipart_encoder)}
            data: $multipart_encoder
        validate:
            - eq: ["status_code", 200]
            - startswith: ["content.files.file", "UserName=test"]
"""


def multipart_encoder(**kwargs):
    """ initialize MultipartEncoder with uploading fields.
    """

    def get_filetype(file_path):
        file_type = filetype.guess(file_path)
        if file_type:
            return file_type.mime
        else:
            return "text/html"

    fields_dict = {}
    for key, value in kwargs.items():

        if os.path.isabs(value):
            _file_path = value
            is_file = True
        else:
            global PWD
            _file_path = os.path.join(PWD, value)
            is_file = os.path.isfile(_file_path)

        if is_file:
            filename = os.path.basename(_file_path)
            with open(_file_path, 'rb') as f:
                mime_type = get_filetype(_file_path)
                fields_dict[key] = (filename, f.read(), mime_type)
        else:
            fields_dict[key] = value

    return MultipartEncoder(fields=fields_dict)


def multipart_content_type(multipart_encoder):
    """ prepare Content-Type for request headers
    """
    return multipart_encoder.content_type
