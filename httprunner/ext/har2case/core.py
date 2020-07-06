import base64
import json
import os
import sys
import urllib.parse as urlparse
from typing import Text

from httprunner.compat import ensure_path_sep
from loguru import logger
from sentry_sdk import capture_exception

from httprunner.ext.har2case import utils
from httprunner.make import make_testcase, format_pytest_with_black

try:
    from json.decoder import JSONDecodeError
except ImportError:
    JSONDecodeError = ValueError


def ensure_file_path(path: Text) -> Text:

    if not path or not path.endswith(".har"):
        logger.error("HAR file not specified.")
        sys.exit(1)

    path = ensure_path_sep(path)
    if not os.path.isfile(path):
        logger.error(f"HAR file not exists: {path}")
        sys.exit(1)

    if not os.path.isabs(path):
        path = os.path.join(os.getcwd(), path)

    return path


class HarParser(object):
    def __init__(self, har_file_path, filter_str=None, exclude_str=None):
        self.har_file_path = ensure_file_path(har_file_path)
        self.filter_str = filter_str
        self.exclude_str = exclude_str or ""

    def __make_request_url(self, teststep_dict, entry_json):
        """ parse HAR entry request url and queryString, and make teststep url and params

        Args:
            entry_json (dict):
                {
                    "request": {
                        "url": "https://httprunner.top/home?v=1&w=2",
                        "queryString": [
                            {"name": "v", "value": "1"},
                            {"name": "w", "value": "2"}
                        ],
                    },
                    "response": {}
                }

        Returns:
            {
                "name: "/home",
                "request": {
                    url: "https://httprunner.top/home",
                    params: {"v": "1", "w": "2"}
                }
            }

        """
        request_params = utils.convert_list_to_dict(
            entry_json["request"].get("queryString", [])
        )

        url = entry_json["request"].get("url")
        if not url:
            logger.exception("url missed in request.")
            sys.exit(1)

        parsed_object = urlparse.urlparse(url)
        if request_params:
            parsed_object = parsed_object._replace(query="")
            teststep_dict["request"]["url"] = parsed_object.geturl()
            teststep_dict["request"]["params"] = request_params
        else:
            teststep_dict["request"]["url"] = url

        teststep_dict["name"] = parsed_object.path

    def __make_request_method(self, teststep_dict, entry_json):
        """ parse HAR entry request method, and make teststep method.
        """
        method = entry_json["request"].get("method")
        if not method:
            logger.exception("method missed in request.")
            sys.exit(1)

        teststep_dict["request"]["method"] = method

    def __make_request_cookies(self, teststep_dict, entry_json):
        cookies = {}
        for cookie in entry_json["request"].get("cookies", []):
            cookies[cookie["name"]] = cookie["value"]

        if cookies:
            teststep_dict["request"]["cookies"] = cookies

    def __make_request_headers(self, teststep_dict, entry_json):
        """ parse HAR entry request headers, and make teststep headers.
            header in IGNORE_REQUEST_HEADERS will be ignored.

        Args:
            entry_json (dict):
                {
                    "request": {
                        "headers": [
                            {"name": "Host", "value": "httprunner.top"},
                            {"name": "Content-Type", "value": "application/json"},
                            {"name": "User-Agent", "value": "iOS/10.3"}
                        ],
                    },
                    "response": {}
                }

        Returns:
            {
                "request": {
                    headers: {"Content-Type": "application/json"}
            }

        """
        teststep_headers = {}
        for header in entry_json["request"].get("headers", []):
            if header["name"] == "cookie" or header["name"].startswith(":"):
                continue

            teststep_headers[header["name"]] = header["value"]

        if teststep_headers:
            teststep_dict["request"]["headers"] = teststep_headers

    def _make_request_data(self, teststep_dict, entry_json):
        """ parse HAR entry request data, and make teststep request data

        Args:
            entry_json (dict):
                {
                    "request": {
                        "method": "POST",
                        "postData": {
                            "mimeType": "application/x-www-form-urlencoded; charset=utf-8",
                            "params": [
                                {"name": "a", "value": 1},
                                {"name": "b", "value": "2"}
                            }
                        },
                    },
                    "response": {...}
                }


        Returns:
            {
                "request": {
                    "method": "POST",
                    "data": {"v": "1", "w": "2"}
                }
            }

        """
        method = entry_json["request"].get("method")
        if method in ["POST", "PUT", "PATCH"]:
            postData = entry_json["request"].get("postData", {})
            mimeType = postData.get("mimeType")

            # Note that text and params fields are mutually exclusive.
            if "text" in postData:
                post_data = postData.get("text")
            else:
                params = postData.get("params", [])
                post_data = utils.convert_list_to_dict(params)

            request_data_key = "data"
            if not mimeType:
                pass
            elif mimeType.startswith("application/json"):
                try:
                    post_data = json.loads(post_data)
                    request_data_key = "json"
                except JSONDecodeError:
                    pass
            elif mimeType.startswith("application/x-www-form-urlencoded"):
                post_data = utils.convert_x_www_form_urlencoded_to_dict(post_data)
            else:
                # TODO: make compatible with more mimeType
                pass

            teststep_dict["request"][request_data_key] = post_data

    def _make_validate(self, teststep_dict, entry_json):
        """ parse HAR entry response and make teststep validate.

        Args:
            entry_json (dict):
                {
                    "request": {},
                    "response": {
                        "status": 200,
                        "headers": [
                            {
                                "name": "Content-Type",
                                "value": "application/json; charset=utf-8"
                            },
                        ],
                        "content": {
                            "size": 71,
                            "mimeType": "application/json; charset=utf-8",
                            "text": "eyJJc1N1Y2Nlc3MiOnRydWUsIkNvZGUiOjIwMCwiTWVzc2FnZSI6bnVsbCwiVmFsdWUiOnsiQmxuUmVzdWx0Ijp0cnVlfX0=",
                            "encoding": "base64"
                        }
                    }
                }

        Returns:
            {
                "validate": [
                    {"eq": ["status_code", 200]}
                ]
            }

        """
        teststep_dict["validate"].append(
            {"eq": ["status_code", entry_json["response"].get("status")]}
        )

        resp_content_dict = entry_json["response"].get("content")

        headers_mapping = utils.convert_list_to_dict(
            entry_json["response"].get("headers", [])
        )
        if "Content-Type" in headers_mapping:
            teststep_dict["validate"].append(
                {"eq": ["headers.Content-Type", headers_mapping["Content-Type"]]}
            )

        text = resp_content_dict.get("text")
        if not text:
            return

        mime_type = resp_content_dict.get("mimeType")
        if mime_type and mime_type.startswith("application/json"):

            encoding = resp_content_dict.get("encoding")
            if encoding and encoding == "base64":
                content = base64.b64decode(text)
                try:
                    content = content.decode("utf-8")
                except UnicodeDecodeError:
                    logger.warning(f"failed to decode base64 content with utf-8 !")
                    return
            else:
                content = text

            try:
                resp_content_json = json.loads(content)
            except JSONDecodeError:
                logger.warning(f"response content can not be loaded as json: {content}")
                return

            if not isinstance(resp_content_json, dict):
                # e.g. ['a', 'b']
                return

            for key, value in resp_content_json.items():
                if isinstance(value, (dict, list)):
                    continue

                teststep_dict["validate"].append({"eq": ["body.{}".format(key), value]})

    def _prepare_teststep(self, entry_json):
        """ extract info from entry dict and make teststep

        Args:
            entry_json (dict):
                {
                    "request": {
                        "method": "POST",
                        "url": "https://httprunner.top/api/v1/Account/Login",
                        "headers": [],
                        "queryString": [],
                        "postData": {},
                    },
                    "response": {
                        "status": 200,
                        "headers": [],
                        "content": {}
                    }
                }

        """
        teststep_dict = {"name": "", "request": {}, "validate": []}

        self.__make_request_url(teststep_dict, entry_json)
        self.__make_request_method(teststep_dict, entry_json)
        self.__make_request_cookies(teststep_dict, entry_json)
        self.__make_request_headers(teststep_dict, entry_json)
        self._make_request_data(teststep_dict, entry_json)
        self._make_validate(teststep_dict, entry_json)

        return teststep_dict

    def _prepare_config(self):
        """ prepare config block.
        """
        return {"name": "testcase description", "variables": {}, "verify": False}

    def _prepare_teststeps(self):
        """ make teststep list.
            teststeps list are parsed from HAR log entries list.

        """

        def is_exclude(url, exclude_str):
            exclude_str_list = exclude_str.split("|")
            for exclude_str in exclude_str_list:
                if exclude_str and exclude_str in url:
                    return True

            return False

        teststeps = []
        log_entries = utils.load_har_log_entries(self.har_file_path)
        for entry_json in log_entries:
            url = entry_json["request"].get("url")
            if self.filter_str and self.filter_str not in url:
                continue

            if is_exclude(url, self.exclude_str):
                continue

            teststeps.append(self._prepare_teststep(entry_json))

        return teststeps

    def _make_testcase(self):
        """ Extract info from HAR file and prepare for testcase
        """
        logger.info("Extract info from HAR file and prepare for testcase.")

        config = self._prepare_config()
        teststeps = self._prepare_teststeps()

        testcase = {"config": config, "teststeps": teststeps}
        return testcase

    def gen_testcase(self, file_type="pytest"):
        logger.info(f"Start to generate testcase from {self.har_file_path}")
        harfile = os.path.splitext(self.har_file_path)[0]

        try:
            testcase = self._make_testcase()
        except Exception as ex:
            capture_exception(ex)
            raise

        if file_type == "JSON":
            output_testcase_file = f"{harfile}.json"
            utils.dump_json(testcase, output_testcase_file)
        elif file_type == "YAML":
            output_testcase_file = f"{harfile}.yml"
            utils.dump_yaml(testcase, output_testcase_file)
        else:
            # default to generate pytest file
            testcase["config"]["path"] = self.har_file_path
            output_testcase_file = make_testcase(testcase)
            format_pytest_with_black(output_testcase_file)

        logger.info(f"generated testcase: {output_testcase_file}")
