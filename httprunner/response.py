# encoding: utf-8

import json
import re

from httprunner import exception, logger, testcase, utils
from httprunner.compat import OrderedDict, basestring
from requests.structures import CaseInsensitiveDict
from requests.models import PreparedRequest

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
            raise exception.ParamsError(err_msg)

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
            err_msg = u"Failed to extract data with regex!\n"
            err_msg += u"response content: {}\n".format(self.content)
            err_msg += u"regex: {}\n".format(field)
            logger.log_error(err_msg)
            raise exception.ParamsError(err_msg)

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
        try:
            # string.split(sep=None, maxsplit=-1) -> list of strings
            # e.g. "content.person.name" => ["content", "person.name"]
            try:
                top_query, sub_query = field.split('.', 1)
            except ValueError:
                top_query = field
                sub_query = None

            if top_query == "cookies":
                cookies = self.cookies
                try:
                    return cookies[sub_query]
                except KeyError:
                    err_msg = u"Failed to extract attribute from cookies!\n"
                    err_msg += u"cookies: {}\n".format(cookies)
                    err_msg += u"attribute: {}".format(sub_query)
                    logger.log_error(err_msg)
                    raise exception.ParamsError(err_msg)
            elif top_query == "elapsed":
                if sub_query in ["days", "seconds", "microseconds"]:
                    return getattr(self.elapsed, sub_query)
                elif sub_query == "total_seconds":
                    return self.elapsed.total_seconds()
                else:
                    err_msg = "{}: {} is not valid timedelta attribute.\n".format(field, sub_query)
                    err_msg += "elapsed only support attributes: days, seconds, microseconds, total_seconds.\n"
                    logger.log_error(err_msg)
                    raise exception.ParamsError(err_msg)

            try:
                top_query_content = getattr(self, top_query)
            except AttributeError:
                err_msg = u"Failed to extract attribute from response object: resp_obj.{}".format(top_query)
                logger.log_error(err_msg)
                raise exception.ParamsError(err_msg)

            if sub_query:
                if not isinstance(top_query_content, (dict, CaseInsensitiveDict, list)):
                    try:
                        # TODO: remove compatibility for content, text
                        if isinstance(top_query_content, bytes):
                            top_query_content = top_query_content.decode("utf-8")

                        if isinstance(top_query_content, PreparedRequest):
                            top_query_content = top_query_content.__dict__
                        else:
                            top_query_content = json.loads(top_query_content)
                    except json.decoder.JSONDecodeError:
                        err_msg = u"Failed to extract data with delimiter!\n"
                        err_msg += u"response content: {}\n".format(self.content)
                        err_msg += u"regex: {}\n".format(field)
                        logger.log_error(err_msg)
                        raise exception.ParamsError(err_msg)

                # e.g. key: resp_headers_content_type, sub_query = "content-type"
                return utils.query_json(top_query_content, sub_query)
            else:
                # e.g. key: resp_status_code, resp_content
                return top_query_content

        except AttributeError:
            err_msg = u"Failed to extract value from response!\n"
            err_msg += u"response content: {}\n".format(self.content)
            err_msg += u"extract field: {}\n".format(field)
            logger.log_error(err_msg)
            raise exception.ParamsError(err_msg)

    def extract_field(self, field):
        """ extract value from requests.Response.
        """
        msg = "extract field: {}".format(field)

        try:
            if text_extractor_regexp_compile.match(field):
                value = self._extract_field_with_regex(field)
            else:
                value = self._extract_field_with_delimiter(field)

            msg += "\t=> {}".format(value)
            logger.log_debug(msg)

        # TODO: unify ParseResponseError type
        except (exception.ParseResponseError, TypeError):
            logger.log_error("failed to extract field: {}".format(field))
            raise

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
        extract_binds_order_dict = utils.convert_to_order_dict(extractors)

        for key, field in extract_binds_order_dict.items():
            if not isinstance(field, basestring):
                raise exception.ParamsError("invalid extractors in testcase!")

            extracted_variables_mapping[key] = self.extract_field(field)

        return extracted_variables_mapping
