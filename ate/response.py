from ate import utils, exception


class ResponseObject(object):

    def __init__(self, resp_obj):
        """ initialize with a requests.Response object
        @param (requests.Response instance) resp_obj
        """
        self.resp_obj = resp_obj
        self.success = True

    def parsed_body(self):
        try:
            return self.resp_obj.json()
        except ValueError:
            return self.resp_obj.text

    def parsed_dict(self):
        return {
            'status_code': self.resp_obj.status_code,
            'headers': self.resp_obj.headers,
            'body': self.parsed_body()
        }

    def extract_field(self, field, delimiter='.'):
        """ extract field from requests.Response
        @param (str) field of requests.Response object, and may be joined by delimiter
            "status_code"
            "content"
            "headers.content-type"
            "content.person.name.first_name"
        """
        try:
            field += "."
            # string.split(sep=None, maxsplit=-1) -> list of strings
            # e.g. "content.person.name" => ["content", "person.name"]
            top_query, sub_query = field.split(delimiter, 1)

            if top_query in ["body", "content", "text"]:
                json_content = self.parsed_body()
            else:
                json_content = getattr(self.resp_obj, top_query)

            if sub_query:
                # e.g. key: resp_headers_content_type, sub_query = "content-type"
                return utils.query_json(json_content, sub_query)
            else:
                # e.g. key: resp_status_code, resp_content
                return json_content

        except AttributeError:
            raise exception.ParamsError("invalid extract_binds!")

    def extract_response(self, extract_binds):
        """ extract content from requests.Response
        @param (dict) extract_binds
            {
                "resp_status_code": "status_code",
                "resp_headers_content_type": "headers.content-type",
                "resp_content": "content",
                "resp_content_person_first_name": "content.person.name.first_name"
            }
        """
        extracted_variables_mapping = {}

        for key, field in extract_binds.items():
            if not isinstance(field, utils.string_type):
                raise exception.ParamsError("invalid extract_binds!")

            extracted_variables_mapping[key] = self.extract_field(field)

        return extracted_variables_mapping

    def validate(self, validators, variables_mapping):
        """ Bind named validators to value within the context.
        @param (list) validators
            [
                {"check": "status_code", "comparator": "eq", "expected": 201},
                {"check": "resp_body_success", "comparator": "eq", "expected": True}
            ]
        @param (dict) variables_mapping
            {
                "resp_body_success": True
            }
        @return (list) content differences
            [
                {
                    "check": "status_code",
                    "comparator": "eq", "expected": 201, "value": 200
                }
            ]
        """
        diff_content_list = []

        for validator_dict in validators:

            if "expected" not in validator_dict or "check" not in validator_dict:
                raise exception.ParamsError("expected not specified in validator")

            validator_key = validator_dict["check"]
            try:
                validator_dict["value"] = variables_mapping[validator_key]
            except KeyError:
                validator_dict["value"] = self.extract_field(validator_key)

            match_expected = utils.match_expected(
                validator_dict["value"],
                validator_dict["expected"],
                validator_dict.get("comparator", "eq")
            )

            if not match_expected:
                diff_content_list.append(validator_dict)

        self.success = False if diff_content_list else True
        return diff_content_list
