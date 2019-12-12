""" upload test plugin.

If you want to use this plugin, you should install the following dependencies first.

- requests_toolbelt
- filetype

Then you can write upload test script as below:

    - test:
        name: upload file
        request:
            url: http://httpbin.org/upload
            method: POST
            headers:
                Cookie: session=AAA-BBB-CCC
            upload:
                file: "data/file_to_upload"
                field1: "value1"
                field2: "value2"
        validate:
            - eq: ["status_code", 200]

"""

import os
import sys

try:
    import filetype
    from requests_toolbelt import MultipartEncoder
except ImportError:
    msg = """
uploader plugin dependencies uninstalled, install first and try again.
install with pip:
$ pip install requests_toolbelt filetype
"""
    print(msg)
    sys.exit(0)

from httprunner.exceptions import ParamsError


def prepare_upload_test(test_dict):
    """ preprocess for upload test
        replace `upload` info with MultipartEncoder

    Args:
        test_dict (dict):

            {
                "variables": {},
                "request": {
                    "url": "http://httpbin.org/upload",
                    "method": "POST",
                    "headers": {
                        "Cookie": "session=AAA-BBB-CCC"
                    },
                    "upload": {
                        "file": "data/file_to_upload"
                        "md5": "123"
                    }
                }
            }


    Returns:
        (dict, dict):
            - variables: prepared variables for upload test
            - request: prepared request for upload test

    """
    upload_json = test_dict["request"].pop("upload", {})
    if not upload_json:
        raise ParamsError("invalid upload info: {}".format(upload_json))

    params_list = []
    for key, value in upload_json.items():

        if os.path.isabs(value) and os.path.isfile(value):
            # value is absolute file path, do nothing
            pass
        else:
            # value is not absolute file path, check if it is relative file path
            from httprunner.loader import get_pwd
            _file_path = os.path.join(get_pwd(), value)
            if os.path.isfile(_file_path):
                # value is relative file path, convert it to be absolute path
                value = _file_path

        test_dict["variables"][key] = value
        params_list.append("{}=${}".format(key, key))

    params_str = ", ".join(params_list)
    test_dict["variables"]["m_encoder"] = "${multipart_encoder(" + params_str + ")}"
    test_dict["request"]["headers"]["Content-Type"] = "${multipart_content_type($m_encoder)}"
    test_dict["request"]["data"] = "$m_encoder"


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

        if os.path.isfile(value):
            # value is file path to upload,
            # it has been ensured to be absolute in prepare_upload_test
            filename = os.path.basename(value)
            with open(value, 'rb') as f:
                mime_type = get_filetype(value)
                fields_dict[key] = (filename, f.read(), mime_type)
        else:
            fields_dict[key] = value

    return MultipartEncoder(fields=fields_dict)


def multipart_content_type(m_encoder):
    """ prepare Content-Type for request headers
    """
    return m_encoder.content_type
