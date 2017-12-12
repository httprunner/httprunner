import logging
import re
from collections import OrderedDict

from httprunner import exception, utils
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
            err_msg = u"Extractor error: failed to extract data with regex!\n"
            err_msg += u"response body: {}\n".format(self.resp_text)
            err_msg += u"regex: {}\n".format(field)
            logging.error(err_msg)
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
                    err_msg = u"Extractor error: failed to extract data with regex!\n"
                    err_msg += u"response: {}\n".format(self.parsed_dict())
                    err_msg += u"regex: {}\n".format(field)
                    logging.error(err_msg)
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
            logging.error(err_msg)
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

    def parse_validator(self, validator, variables_mapping):
        """ parse validator, validator maybe in two format
        @param (dict) validator
            format1: this is kept for compatiblity with the previous versions.
                {"check": "status_code", "comparator": "eq", "expect": 201}
                {"check": "resp_body_success", "comparator": "eq", "expect": True}
            format2: recommended new version
                {'eq': ['status_code', 201]}
                {'eq': ['resp_body_success', True]}
        @param (dict) variables_mapping
            {
                "resp_body_success": True
            }
        @return validator info
            check_item, check_value, expect_value, comparator
        """
        if not isinstance(validator, dict):
            raise exception.ParamsError("invalid validator: {}".format(validator))

        if "check" in validator and len(validator) > 1:
            # format1
            check_item = validator.get("check")

            if "expect" in validator:
                expect_value = validator.get("expect")
            elif "expected" in validator:
                expect_value = validator.get("expected")
            else:
                raise exception.ParamsError("invalid validator: {}".format(validator))

            comparator = validator.get("comparator", "eq")

        elif len(validator) == 1:
            # format2
            comparator = list(validator.keys())[0]
            compare_values = validator[comparator]

            if not isinstance(compare_values, list) or len(compare_values) != 2:
                raise exception.ParamsError("invalid validator: {}".format(validator))

            check_item, expect_value = compare_values

        else:
            raise exception.ParamsError("invalid validator: {}".format(validator))

        if check_item in variables_mapping:
            check_value = variables_mapping[check_item]
        else:
            try:
                check_value = self.extract_field(check_item)
            except exception.ParseResponseError:
                raise exception.ParseResponseError("failed to extract check item in response!")

        return check_item, check_value, expect_value, comparator

    def validate(self, validators, variables_mapping):
        """ check validators with the context variable mapping.
        """
        for validator in validators:
            check_item, check_value, expect_value, comparator = self.parse_validator(validator, variables_mapping)

            utils.match_expected(
                check_value,
                expect_value,
                comparator,
                check_item
            )

        return True
