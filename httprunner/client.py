# encoding: utf-8

import time

import requests
import urllib3
from requests import Request, Response
from requests.exceptions import (InvalidSchema, InvalidURL, MissingSchema,
                                 RequestException)

from httprunner import logger, response
from httprunner.utils import lower_dict_keys, omit_long_data

urllib3.disable_warnings(urllib3.exceptions.InsecureRequestWarning)


def get_req_resp_record(resp_obj):
    """ get request and response info from Response() object.
    """
    def log_print(req_resp_dict, r_type):
        msg = "\n================== {} details ==================\n".format(r_type)
        for key, value in req_resp_dict[r_type].items():
            msg += "{:<16} : {}\n".format(key, repr(value))
        logger.log_debug(msg)

    req_resp_dict = {
        "request": {},
        "response": {}
    }

    # record actual request info
    req_resp_dict["request"]["url"] = resp_obj.request.url
    req_resp_dict["request"]["method"] = resp_obj.request.method
    req_resp_dict["request"]["headers"] = dict(resp_obj.request.headers)

    request_body = resp_obj.request.body
    if request_body:
        request_content_type = lower_dict_keys(
            req_resp_dict["request"]["headers"]
        ).get("content-type")
        if request_content_type and "multipart/form-data" in request_content_type:
            # upload file type
            req_resp_dict["request"]["body"] = "upload file stream (OMITTED)"
        else:
            req_resp_dict["request"]["body"] = request_body

    # log request details in debug mode
    log_print(req_resp_dict, "request")

    # record response info
    req_resp_dict["response"]["ok"] = resp_obj.ok
    req_resp_dict["response"]["url"] = resp_obj.url
    req_resp_dict["response"]["status_code"] = resp_obj.status_code
    req_resp_dict["response"]["reason"] = resp_obj.reason
    req_resp_dict["response"]["cookies"] = resp_obj.cookies or {}
    req_resp_dict["response"]["encoding"] = resp_obj.encoding
    resp_headers = dict(resp_obj.headers)
    req_resp_dict["response"]["headers"] = resp_headers

    lower_resp_headers = lower_dict_keys(resp_headers)
    content_type = lower_resp_headers.get("content-type", "")
    req_resp_dict["response"]["content_type"] = content_type

    if "image" in content_type:
        # response is image type, record bytes content only
        req_resp_dict["response"]["body"] = resp_obj.content
    else:
        try:
            # try to record json data
            if isinstance(resp_obj, response.ResponseObject):
                req_resp_dict["response"]["body"] = resp_obj.json
            else:
                req_resp_dict["response"]["body"] = resp_obj.json()
        except ValueError:
            # only record at most 512 text charactors
            resp_text = resp_obj.text
            req_resp_dict["response"]["body"] = omit_long_data(resp_text)

    # log response details in debug mode
    log_print(req_resp_dict, "response")

    return req_resp_dict


class ApiResponse(Response):

    def raise_for_status(self):
        if hasattr(self, 'error') and self.error:
            raise self.error
        Response.raise_for_status(self)


class HttpSession(requests.Session):
    """
    Class for performing HTTP requests and holding (session-) cookies between requests (in order
    to be able to log in and out of websites). Each request is logged so that HttpRunner can
    display statistics.

    This is a slightly extended version of `python-request <http://python-requests.org>`_'s
    :py:class:`requests.Session` class and mostly this class works exactly the same.
    """
    def __init__(self):
        super(HttpSession, self).__init__()
        self.init_meta_data()

    def init_meta_data(self):
        """ initialize meta_data, it will store detail data of request and response
        """
        self.meta_data = {
            "name": "",
            "data": [
                {
                    "request": {
                        "url": "N/A",
                        "method": "N/A",
                        "headers": {}
                    },
                    "response": {
                        "status_code": "N/A",
                        "headers": {},
                        "encoding": None,
                        "content_type": ""
                    }
                }
            ],
            "stat": {
                "content_size": "N/A",
                "response_time_ms": "N/A",
                "elapsed_ms": "N/A",
            }
        }

    def update_last_req_resp_record(self, resp_obj):
        """
        update request and response info from Response() object.
        """
        self.meta_data["data"].pop()
        self.meta_data["data"].append(get_req_resp_record(resp_obj))

    def request(self, method, url, name=None, **kwargs):
        """
        Constructs and sends a :py:class:`requests.Request`.
        Returns :py:class:`requests.Response` object.

        :param method:
            method for the new :class:`Request` object.
        :param url:
            URL for the new :class:`Request` object.
        :param name: (optional)
            Placeholder, make compatible with Locust's HttpSession
        :param params: (optional)
            Dictionary or bytes to be sent in the query string for the :class:`Request`.
        :param data: (optional)
            Dictionary or bytes to send in the body of the :class:`Request`.
        :param headers: (optional)
            Dictionary of HTTP Headers to send with the :class:`Request`.
        :param cookies: (optional)
            Dict or CookieJar object to send with the :class:`Request`.
        :param files: (optional)
            Dictionary of ``'filename': file-like-objects`` for multipart encoding upload.
        :param auth: (optional)
            Auth tuple or callable to enable Basic/Digest/Custom HTTP Auth.
        :param timeout: (optional)
            How long to wait for the server to send data before giving up, as a float, or \
            a (`connect timeout, read timeout <user/advanced.html#timeouts>`_) tuple.
            :type timeout: float or tuple
        :param allow_redirects: (optional)
            Set to True by default.
        :type allow_redirects: bool
        :param proxies: (optional)
            Dictionary mapping protocol to the URL of the proxy.
        :param stream: (optional)
            whether to immediately download the response content. Defaults to ``False``.
        :param verify: (optional)
            if ``True``, the SSL cert will be verified. A CA_BUNDLE path can also be provided.
        :param cert: (optional)
            if String, path to ssl client cert file (.pem). If Tuple, ('cert', 'key') pair.
        """
        self.init_meta_data()

        # record test name
        self.meta_data["name"] = name

        # record original request info
        self.meta_data["data"][0]["request"]["method"] = method
        self.meta_data["data"][0]["request"]["url"] = url
        kwargs.setdefault("timeout", 120)
        self.meta_data["data"][0]["request"].update(kwargs)

        start_timestamp = time.time()
        response = self._send_request_safe_mode(method, url, **kwargs)
        response_time_ms = round((time.time() - start_timestamp) * 1000, 2)

        # get the length of the content, but if the argument stream is set to True, we take
        # the size from the content-length header, in order to not trigger fetching of the body
        if kwargs.get("stream", False):
            content_size = int(dict(response.headers).get("content-length") or 0)
        else:
            content_size = len(response.content or "")

        # record the consumed time
        self.meta_data["stat"] = {
            "response_time_ms": response_time_ms,
            "elapsed_ms": response.elapsed.microseconds / 1000.0,
            "content_size": content_size
        }

        # record request and response histories, include 30X redirection
        response_list = response.history + [response]
        self.meta_data["data"] = [
            get_req_resp_record(resp_obj)
            for resp_obj in response_list
        ]

        try:
            response.raise_for_status()
        except RequestException as e:
            logger.log_error(u"{exception}".format(exception=str(e)))
        else:
            logger.log_info(
                """status_code: {}, response_time(ms): {} ms, response_length: {} bytes\n""".format(
                    response.status_code,
                    response_time_ms,
                    content_size
                )
            )

        return response

    def _send_request_safe_mode(self, method, url, **kwargs):
        """
        Send a HTTP request, and catch any exception that might occur due to connection problems.
        Safe mode has been removed from requests 1.x.
        """
        try:
            msg = "processed request:\n"
            msg += "> {method} {url}\n".format(method=method, url=url)
            msg += "> kwargs: {kwargs}".format(kwargs=kwargs)
            logger.log_debug(msg)
            return requests.Session.request(self, method, url, **kwargs)
        except (MissingSchema, InvalidSchema, InvalidURL):
            raise
        except RequestException as ex:
            resp = ApiResponse()
            resp.error = ex
            resp.status_code = 0  # with this status_code, content returns None
            resp.request = Request(method, url).prepare()
            return resp
