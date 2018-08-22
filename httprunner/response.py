# encoding: utf-8

import json
import re

from httprunner import exceptions, logger, utils
from httprunner.compat import OrderedDict, basestring, is_py2
from requests.models import PreparedRequest
from requests.structures import CaseInsensitiveDict

text_extractor_regexp_compile = re.compile(r".*\(.*\).*")


class ResponseObject(object):

    def __init__(self, resp_obj):
        """ initialize with a requests.Response object
        @param (requests.Response instance) resp_obj
        """
        self.resp_obj = resp_obj

    def __getattr__(self, key):
        try:
            if key == "json":
                value = self.resp_obj.json()
            else:
                value =  getattr(self.resp_obj, key)

            self.__dict__[key] = value
            return value
        except AttributeError:
            err_msg = "ResponseObject does not have attribute: {}".format(key)
            logger.log_error(err_msg)
            raise exceptions.ParamsError(err_msg)

    def _extract_field_with_regex(self, field):
        """ extract field from response content with regex.
            requests.Response body could be json or html text.
        @param (str) field should only be regex string that matched r".*\(.*\).*"
        e.g.
            self.text: "LB123abcRB789"
            field: "LB[\d]*(.*)RB[\d]*"
            return: abc
        """
        matched = re.search(field, self.text)
        if not matched:
            err_msg = u"Failed to extract data with regex! => {}\n".format(field)
            err_msg += u"response body: {}\n".format(self.text)
            logger.log_error(err_msg)
            raise exceptions.ExtractFailure(err_msg)

        return matched.group(1)

    def _extract_field_with_delimiter(self, field):
        """ response content could be json or html text.
        @param (str) field should be string joined by delimiter.
        e.g.
            "status_code"
            "headers"
            "cookies"
            "content"
            "headers.content-type"
            "content.person.name.first_name"
        """
        # string.split(sep=None, maxsplit=-1) -> list of strings
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
                err_msg = u"Failed to extract: {}\n".format(field)
                logger.log_error(err_msg)
                raise exceptions.ParamsError(err_msg)

            return getattr(self, top_query)

        # cookies
        elif top_query == "cookies":
            cookies = self.cookies.get_dict()
            if not sub_query:
                # extract cookies
                return cookies

            try:
                return cookies[sub_query]
            except KeyError:
                err_msg = u"Failed to extract cookie! => {}\n".format(field)
                err_msg += u"response cookies: {}\n".format(cookies)
                logger.log_error(err_msg)
                raise exceptions.ExtractFailure(err_msg)

        # elapsed
        elif top_query == "elapsed":
            available_attributes = u"available attributes: days, seconds, microseconds, total_seconds"
            if not sub_query:
                err_msg = u"elapsed is datetime.timedelta instance, attribute should also be specified!\n"
                err_msg += available_attributes
                logger.log_error(err_msg)
                raise exceptions.ParamsError(err_msg)
            elif sub_query in ["days", "seconds", "microseconds"]:
                return getattr(self.elapsed, sub_query)
            elif sub_query == "total_seconds":
                return self.elapsed.total_seconds()
            else:
                err_msg = "{} is not valid datetime.timedelta attribute.\n".format(sub_query)
                err_msg += available_attributes
                logger.log_error(err_msg)
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
                err_msg = u"Failed to extract header! => {}\n".format(field)
                err_msg += u"response headers: {}\n".format(headers)
                logger.log_error(err_msg)
                raise exceptions.ExtractFailure(err_msg)

        # response body
        elif top_query in ["content", "text", "json"]:
            try:
                body = self.json
            except exceptions.JSONDecodeError:
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
                err_msg = u"Failed to extract attribute from response body! => {}\n".format(field)
                err_msg += u"response body: {}\n".format(body)
                logger.log_error(err_msg)
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
                err_msg = u"Failed to extract cumstom set attribute from teardown hooks! => {}\n".format(field)
                err_msg += u"response set attributes: {}\n".format(attributes)
                logger.log_error(err_msg)
                raise exceptions.TeardownHooksFailure(err_msg)

        # others
        else:
            err_msg = u"Failed to extract attribute from response! => {}\n".format(field)
            err_msg += u"available response attributes: status_code, cookies, elapsed, headers, content, text, json, encoding, ok, reason, url.\n\n"
            err_msg += u"If you want to set attribute in teardown_hooks, take the following example as reference:\n"
            err_msg += u"response.new_attribute = 'new_attribute_value'\n"
            logger.log_error(err_msg)
            raise exceptions.ParamsError(err_msg)

    def extract_field(self, field):
        """ extract value from requests.Response.
        """
        if not isinstance(field, basestring):
            err_msg = u"Invalid extractor! => {}\n".format(field)
            logger.log_error(err_msg)
            raise exceptions.ParamsError(err_msg)

        msg = "extract: {}".format(field)

        if text_extractor_regexp_compile.match(field):
            value = self._extract_field_with_regex(field)
        else:
            value = self._extract_field_with_delimiter(field)

        if is_py2 and isinstance(value, unicode):
            value = value.encode("utf-8")

        msg += "\t=> {}".format(value)
        logger.log_debug(msg)

        return value

    def extract_response(self, extractors):
        """ extract value from requests.Response and store in OrderedDict.
        @param (list) extractors
            [
                {"resp_status_code": "status_code"},
                {"resp_headers_content_type": "headers.content-type"},
                {"resp_content": "content"},
                {"resp_content_person_first_name": "content.person.name.first_name"}
            ]
        @return (OrderDict) variable binds ordered dict
        """
        if not extractors:
            return {}

        logger.log_info("start to extract from response object.")
        extracted_variables_mapping = OrderedDict()
        extract_binds_order_dict = utils.convert_mappinglist_to_orderdict(extractors)

        for key, field in extract_binds_order_dict.items():
            extracted_variables_mapping[key] = self.extract_field(field)

        return extracted_variables_mapping
