import re
from collections import OrderedDict

from httprunner import exception, logger, testcase, utils
from requests.structures import CaseInsensitiveDict

text_extractor_regexp_compile = re.compile(r".*\(.*\).*")


class ResponseObject(object):

    def __init__(self, resp_obj):
        """ initialize with a requests.Response object
        @param (requests.Response instance) resp_obj
        """
        self.resp_obj = resp_obj
        self.resp_text = resp_obj.text
        self.resp_body = self.parsed_body()

    def parsed_body(self):
        try:
            return self.resp_obj.json()
        except ValueError:
            return self.resp_text

    def parsed_dict(self):
        return {
            'status_code': self.resp_obj.status_code,
            'headers': self.resp_obj.headers,
            'body': self.resp_body
        }

    def _extract_field_with_regex(self, field):
        """ extract field from response content with regex.
            requests.Response body could be json or html text.
        @param (str) field should only be regex string that matched r".*\(.*\).*"
        e.g.
            self.resp_text: "LB123abcRB789"
            field: "LB[\d]*(.*)RB[\d]*"
            return: abc
        """
        matched = re.search(field, self.resp_text)
        if not matched:
            err_msg = u"Failed to extract data with regex!\n"
            err_msg += u"response body: {}\n".format(self.resp_text)
            err_msg += u"regex: {}\n".format(field)
            logger.log_error(err_msg)
            raise exception.ParamsError(err_msg)

        return matched.group(1)

    def _extract_field_with_delimiter(self, field):
        """ response content could be json or html text.
        @param (str) field should be string joined by delimiter.
        e.g.
            "status_code"
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

            if top_query in ["body", "content", "text"]:
                top_query_content = self.parsed_body()
            else:
                top_query_content = getattr(self.resp_obj, top_query)

            if sub_query:
                if not isinstance(top_query_content, (dict, CaseInsensitiveDict, list)):
                    err_msg = u"Failed to extract data with regex!\n"
                    err_msg += u"response: {}\n".format(self.parsed_dict())
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
            err_msg += u"response: {}\n".format(self.parsed_dict())
            err_msg += u"extract field: {}\n".format(field)
            logger.log_error(err_msg)
            raise exception.ParamsError(err_msg)

    def extract_field(self, field):
        """ extract value from requests.Response.
        """
        if text_extractor_regexp_compile.match(field):
            return self._extract_field_with_regex(field)
        else:
            return self._extract_field_with_delimiter(field)

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
        extracted_variables_mapping = OrderedDict()
        extract_binds_order_dict = utils.convert_to_order_dict(extractors)

        for key, field in extract_binds_order_dict.items():
            if not isinstance(field, utils.string_type):
                raise exception.ParamsError("invalid extractors in testcase!")

            extracted_variables_mapping[key] = self.extract_field(field)

        return extracted_variables_mapping
