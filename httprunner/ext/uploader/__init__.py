""" upload test extension.

If you want to use this extension, you should install the following dependencies first.

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

For compatibility, you can also write upload test script in old way:

    - test:
        name: upload file
        variables:
            file: "data/file_to_upload"
            field1: "value1"
            field2: "value2"
            m_encoder: ${multipart_encoder(file=$file, field1=$field1, field2=$field2)}
        request:
            url: http://httpbin.org/upload
            method: POST
            headers:
                Content-Type: ${multipart_content_type($m_encoder)}
                Cookie: session=AAA-BBB-CCC
            data: $m_encoder
        validate:
            - eq: ["status_code", 200]

"""

import os
import sys
from typing import Text, NoReturn

from loguru import logger

from httprunner.parser import parse_variables_mapping
from httprunner.schema import TStep, FunctionsMapping

try:
    import filetype
    from requests_toolbelt import MultipartEncoder

    UPLOAD_READY = True
except ModuleNotFoundError:
    UPLOAD_READY = False


def prepare_upload_step(step: TStep, functions: FunctionsMapping) -> "NoReturn":
    """ preprocess for upload test
        replace `upload` info with MultipartEncoder

    Args:
        step: teststep
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
        functions: functions mapping

    """
    if not step.request.upload:
        return

    if not UPLOAD_READY:
        msg = """
uploader extension dependencies uninstalled, install first and try again.
install with pip:
$ pip install requests_toolbelt filetype
"""
        logger.error(msg)
        sys.exit(1)

    params_list = []
    for key, value in step.request.upload.items():
        step.variables[key] = value
        params_list.append(f"{key}=${key}")

    params_str = ", ".join(params_list)
    step.variables["m_encoder"] = "${multipart_encoder(" + params_str + ")}"

    # parse variables
    step.variables = parse_variables_mapping(step.variables, functions)

    step.request.headers["Content-Type"] = "${multipart_content_type($m_encoder)}"

    step.request.data = "$m_encoder"


def multipart_encoder(**kwargs):
    """ initialize MultipartEncoder with uploading fields.

    Returns:
        MultipartEncoder: initialized MultipartEncoder object

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
            # value is absolute file path
            _file_path = value
            is_exists_file = os.path.isfile(value)
        else:
            # value is not absolute file path, check if it is relative file path
            from httprunner.loader import load_project_meta

            project_meta = load_project_meta(os.getcwd())

            _file_path = os.path.join(project_meta.PWD, value)
            is_exists_file = os.path.isfile(_file_path)

        if is_exists_file:
            # value is file path to upload
            filename = os.path.basename(_file_path)
            mime_type = get_filetype(_file_path)
            # TODO: fix ResourceWarning for unclosed file
            file_handler = open(_file_path, "rb")
            fields_dict[key] = (filename, file_handler, mime_type)
        else:
            fields_dict[key] = value

    return MultipartEncoder(fields=fields_dict)


def multipart_content_type(m_encoder) -> Text:
    """ prepare Content-Type for request headers

    Args:
        m_encoder: MultipartEncoder object

    Returns:
        content type

    """
    return m_encoder.content_type
