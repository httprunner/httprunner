import json
import re
from collections import OrderedDict

import jsonpath
from loguru import logger

from httprunner import exceptions, utils

text_extractor_regexp_compile = re.compile(r".*\(.*\).*")


class ResponseObject(object):

    def __init__(self, resp_obj):
        """ initialize with a requests.Response object

        Args:
            resp_obj (instance): requests.Response instance

        """
        self.resp_obj = resp_obj

    def __getattr__(self, key):
        try:
            if key == "json":
                value = self.resp_obj.json()
            elif key == "cookies":
                value = self.resp_obj.cookies.get_dict()
            else:
                value = getattr(self.resp_obj, key)

            self.__dict__[key] = value
            return value
        except AttributeError:
            err_msg = f"ResponseObject does not have attribute: {key}"
            logger.error(err_msg)
            raise exceptions.ParamsError(err_msg)

    def _extract_field_with_jsonpath(self, field: str) -> list:
        """ extract field from response content with jsonpath expression.
        JSONPath Docs: https://goessner.net/articles/JsonPath/

        Args:
            field: jsonpath expression, e.g. $.code, $..items.*.id

        Returns:
            A list that extracted from json response example. 1) [200] 2) [1, 2]

        Raises:
            exceptions.ExtractFailure: If no content matched with jsonpath expression.

        Examples:
            For example, response body like below:
            {
                "code": 200,
                "data": {
                    "items": [{
                            "id": 1,
                            "name": "Bob"
                        },
                        {
                            "id": 2,
                            "name": "James"
                        }
                    ]
                },
                "message": "success"
            }

            >>> _extract_field_with_regex("$.code")
            [200]
            >>> _extract_field_with_regex("$..items.*.id")
            [1, 2]

        """
        try:
            json_body = self.json
            assert json_body

            result = jsonpath.jsonpath(json_body, field)
            assert result
            return result
        except (AssertionError, exceptions.JSONDecodeError):
            err_msg = f"Failed to extract data with jsonpath! => {field}\n"
            err_msg += f"response body: {self.text}\n"
            logger.error(err_msg)
            raise exceptions.ExtractFailure(err_msg)

    def _extract_field_with_regex(self, field):
        """ extract field from response content with regex.
            requests.Response body could be json or html text.

        Args:
            field (str): regex string that matched r".*\(.*\).*"

        Returns:
            str: matched content.

        Raises:
            exceptions.ExtractFailure: If no content matched with regex.

        Examples:
            >>> # self.text: "LB123abcRB789"
            >>> filed = "LB[\d]*(.*)RB[\d]*"
            >>> _extract_field_with_regex(field)
            abc

        """
        matched = re.search(field, self.text)
        if not matched:
            err_msg = f"Failed to extract data with regex! => {field}\n"
            err_msg += f"response body: {self.text}\n"
            logger.error(err_msg)
            raise exceptions.ExtractFailure(err_msg)

        return matched.group(1)

    def _extract_field_with_delimiter(self, field):
        """ response content could be json or html text.

        Args:
            field (str): string joined by delimiter.
            e.g.
                "status_code"
                "headers"
                "cookies"
                "content"
                "headers.content-type"
                "content.person.name.first_name"

        """
        # string.split(sep=None, maxsplit=1) -> list of strings
        # e.g. "content.person.name" => ["content", "person.name"]
        try:
            top_query, sub_query = field.split('.', 1)
        except ValueError:
            top_query = field
            sub_query = None

        # status_code
        if top_query in ["status_code", "encoding", "ok", "reason", "url"]:
            if sub_query:
                # status_code.XX
                err_msg = f"Failed to extract: {field}\n"
                logger.error(err_msg)
                raise exceptions.ParamsError(err_msg)

            return getattr(self, top_query)

        # cookies
        elif top_query == "cookies":
            cookies = self.cookies
            if not sub_query:
                # extract cookies
                return cookies

            try:
                return cookies[sub_query]
            except KeyError:
                err_msg = f"Failed to extract cookie! => {field}\n"
                err_msg += f"response cookies: {cookies}\n"
                logger.error(err_msg)
                raise exceptions.ExtractFailure(err_msg)

        # elapsed
        elif top_query == "elapsed":
            available_attributes = u"available attributes: days, seconds, microseconds, total_seconds"
            if not sub_query:
                err_msg = "elapsed is datetime.timedelta instance, attribute should also be specified!\n"
                err_msg += available_attributes
                logger.error(err_msg)
                raise exceptions.ParamsError(err_msg)
            elif sub_query in ["days", "seconds", "microseconds"]:
                return getattr(self.elapsed, sub_query)
            elif sub_query == "total_seconds":
                return self.elapsed.total_seconds()
            else:
                err_msg = f"{sub_query} is not valid datetime.timedelta attribute.\n"
                err_msg += available_attributes
                logger.error(err_msg)
                raise exceptions.ParamsError(err_msg)

        # headers
        elif top_query == "headers":
            headers = self.headers
            if not sub_query:
                # extract headers
                return headers

            try:
                return headers[sub_query]
            except KeyError:
                err_msg = f"Failed to extract header! => {field}\n"
                err_msg += f"response headers: {headers}\n"
                logger.error(err_msg)
                raise exceptions.ExtractFailure(err_msg)

        # response body
        elif top_query in ["body", "content", "text", "json"]:
            try:
                body = self.json
            except json.JSONDecodeError:
                body = self.text

            if not sub_query:
                # extract response body
                return body

            if isinstance(body, (dict, list)):
                # content = {"xxx": 123}, content.xxx
                return utils.query_json(body, sub_query)
            elif sub_query.isdigit():
                # content = "abcdefg", content.3 => d
                return utils.query_json(body, sub_query)
            else:
                # content = "<html>abcdefg</html>", content.xxx
                err_msg = f"Failed to extract attribute from response body! => {field}\n"
                err_msg += f"response body: {body}\n"
                logger.error(err_msg)
                raise exceptions.ExtractFailure(err_msg)

        # new set response attributes in teardown_hooks
        elif top_query in self.__dict__:
            attributes = self.__dict__[top_query]

            if not sub_query:
                # extract response attributes
                return attributes

            if isinstance(attributes, (dict, list)):
                # attributes = {"xxx": 123}, content.xxx
                return utils.query_json(attributes, sub_query)
            elif sub_query.isdigit():
                # attributes = "abcdefg", attributes.3 => d
                return utils.query_json(attributes, sub_query)
            else:
                # content = "attributes.new_attribute_not_exist"
                err_msg = f"Failed to extract cumstom set attribute from teardown hooks! => {field}\n"
                err_msg += f"response set attributes: {attributes}\n"
                logger.error(err_msg)
                raise exceptions.TeardownHooksFailure(err_msg)

        # others
        else:
            err_msg = f"Failed to extract attribute from response! => {field}\n"
            err_msg += "available response attributes: status_code, cookies, elapsed, headers, content, " \
                       "text, json, encoding, ok, reason, url.\n\n"
            err_msg += "If you want to set attribute in teardown_hooks, take the following example as reference:\n"
            err_msg += "response.new_attribute = 'new_attribute_value'\n"
            logger.error(err_msg)
            raise exceptions.ParamsError(err_msg)

    def extract_field(self, field):
        """ extract value from requests.Response.
        """
        if not isinstance(field, str):
            err_msg = f"Invalid extractor! => {field}\n"
            logger.error(err_msg)
            raise exceptions.ParamsError(err_msg)

        msg = f"extract: {field}"

        if field.startswith("$"):
            value = self._extract_field_with_jsonpath(field)
        elif text_extractor_regexp_compile.match(field):
            value = self._extract_field_with_regex(field)
        else:
            value = self._extract_field_with_delimiter(field)

        msg += f"\t=> {value}"
        logger.debug(msg)

        return value

    def extract_response(self, extractors):
        """ extract value from requests.Response and store in OrderedDict.

        Args:
            extractors (list):

                [
                    {"resp_status_code": "status_code"},
                    {"resp_headers_content_type": "headers.content-type"},
                    {"resp_content": "content"},
                    {"resp_content_person_first_name": "content.person.name.first_name"}
                ]

        Returns:
            OrderDict: variable binds ordered dict

        """
        if not extractors:
            return {}

        logger.debug("start to extract from response object.")
        extracted_variables_mapping = OrderedDict()
        extract_binds_order_dict = utils.ensure_mapping_format(extractors)

        for key, field in extract_binds_order_dict.items():
            extracted_variables_mapping[key] = self.extract_field(field)

        return extracted_variables_mapping
