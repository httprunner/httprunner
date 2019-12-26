import re
from collections import OrderedDict

import jsonpath

from httprunner import exceptions, logger, utils
from httprunner.compat import basestring, is_py2

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
            err_msg = "ResponseObject does not have attribute: {}".format(key)
            logger.log_error(err_msg)
            raise exceptions.ParamsError(err_msg)

    def _extract_field_with_jsonpath(self, field):
        """
        JSONPath Docs: https://goessner.net/articles/JsonPath/
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

        :param field:  Jsonpath expression, e.g. 1)$.code   2) $..items.*.id
        :return:       A list that extracted from json repsonse example.    1) [200]   2) [1, 2]
        """
        result = jsonpath.jsonpath(self.parsed_body(), field)
        if result:
            return result
        else:
            raise exceptions.ExtractFailure("\tjsonpath {} get nothing\n".format(field))
            
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
            err_msg = u"Failed to extract data with regex! => {}\n".format(field)
            err_msg += u"response body: {}\n".format(self.text)
            logger.log_error(err_msg)
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
                err_msg = u"Failed to extract: {}\n".format(field)
                logger.log_error(err_msg)
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
        elif top_query in ["body", "content", "text", "json"]:
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
            err_msg += u"available response attributes: status_code, cookies, elapsed, headers, content, " \
                       u"text, json, encoding, ok, reason, url.\n\n"
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

        if field.startswith("$"):
            value = self._extract_field_with_jsonpath(field)
        elif text_extractor_regexp_compile.match(field):
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

        logger.log_debug("start to extract from response object.")
        extracted_variables_mapping = OrderedDict()
        extract_binds_order_dict = utils.ensure_mapping_format(extractors)

        for key, field in extract_binds_order_dict.items():
            extracted_variables_mapping[key] = self.extract_field(field)

        return extracted_variables_mapping
