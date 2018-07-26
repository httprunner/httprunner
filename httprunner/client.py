# encoding: utf-8

import re
import time

import requests
import urllib3
from httprunner import logger
from httprunner.exceptions import ParamsError
from requests import Request, Response
from requests.exceptions import (InvalidSchema, InvalidURL, MissingSchema,
                                 RequestException)

urllib3.disable_warnings(urllib3.exceptions.InsecureRequestWarning)

absolute_http_url_regexp = re.compile(r"^https?://", re.I)


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
    :py:class:`requests.Session` class and mostly this class works exactly the same. However
    the methods for making requests (get, post, delete, put, head, options, patch, request)
    can now take a *url* argument that's only the path part of the URL, in which case the host
    part of the URL will be prepended with the HttpSession.base_url which is normally inherited
    from a HttpRunner class' host property.
    """
    def __init__(self, base_url=None, *args, **kwargs):
        super(HttpSession, self).__init__(*args, **kwargs)
        self.base_url = base_url if base_url else ""
        self.init_meta_data()

    def _build_url(self, path):
        """ prepend url with hostname unless it's already an absolute URL """
        if absolute_http_url_regexp.match(path):
            return path
        elif self.base_url:
            return "{}/{}".format(self.base_url.rstrip("/"), path.lstrip("/"))
        else:
            raise ParamsError("base url missed!")

    def init_meta_data(self):
        """ initialize meta_data, it will store detail data of request and response
        """
        self.meta_data = {
            "request": {
                "url": "N/A",
                "method": "N/A",
                "headers": {},
                "start_timestamp": None
            },
            "response": {
                "status_code": "N/A",
                "headers": {},
                "content_size": "N/A",
                "response_time_ms": "N/A",
                "elapsed_ms": "N/A",
                "encoding": None,
                "content": None,
                "content_type": ""
            }
        }

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
        def log_print(request_response):
            msg = "\n================== {} details ==================\n".format(request_response)
            for key, value in self.meta_data[request_response].items():
                msg += "{:<16} : {}\n".format(key, repr(value))
            logger.log_debug(msg)

        # record original request info
        self.meta_data["request"]["method"] = method
        self.meta_data["request"]["url"] = url
        self.meta_data["request"].update(kwargs)
        self.meta_data["request"]["start_timestamp"] = time.time()

        # prepend url with hostname unless it's already an absolute URL
        url = self._build_url(url)

        kwargs.setdefault("timeout", 120)
        response = self._send_request_safe_mode(method, url, **kwargs)

        # record the consumed time
        self.meta_data["response"]["response_time_ms"] = \
            round((time.time() - self.meta_data["request"]["start_timestamp"]) * 1000, 2)
        self.meta_data["response"]["elapsed_ms"] = response.elapsed.microseconds / 1000.0

        # record actual request info
        self.meta_data["request"]["url"] = (response.history and response.history[0] or response).request.url
        self.meta_data["request"]["headers"] = dict(response.request.headers)
        self.meta_data["request"]["body"] = response.request.body

        # log request details in debug mode
        log_print("request")

        # record response info
        self.meta_data["response"]["ok"] = response.ok
        self.meta_data["response"]["url"] = response.url
        self.meta_data["response"]["status_code"] = response.status_code
        self.meta_data["response"]["reason"] = response.reason
        self.meta_data["response"]["headers"] = dict(response.headers)
        self.meta_data["response"]["cookies"] = response.cookies or {}
        self.meta_data["response"]["encoding"] = response.encoding
        self.meta_data["response"]["content"] = response.content
        self.meta_data["response"]["text"] = response.text
        self.meta_data["response"]["content_type"] = response.headers.get("Content-Type", "")

        try:
            self.meta_data["response"]["json"] = response.json()
        except ValueError:
            self.meta_data["response"]["json"] = None

        # get the length of the content, but if the argument stream is set to True, we take
        # the size from the content-length header, in order to not trigger fetching of the body
        if kwargs.get("stream", False):
            self.meta_data["response"]["content_size"] = int(self.meta_data["response"]["headers"].get("content-length") or 0)
        else:
            self.meta_data["response"]["content_size"] = len(response.content or "")

        # log response details in debug mode
        log_print("response")

        try:
            response.raise_for_status()
        except RequestException as e:
            logger.log_error(u"{exception}".format(exception=str(e)))
        else:
            logger.log_info(
                """status_code: {}, response_time(ms): {} ms, response_length: {} bytes""".format(
                    self.meta_data["response"]["status_code"],
                    self.meta_data["response"]["response_time_ms"],
                    self.meta_data["response"]["content_size"]
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
