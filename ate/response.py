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

    def extract_response(self, extract_binds, delimiter='.'):
        """ extract content from requests.Response
        @param (dict) extract_binds
            {
                "resp_status_code": "status_code",
                "resp_headers_content_type": "headers.content-type",
                "resp_content": "content",
                "resp_content_person_first_name": "content.person.name.first_name"
            }
        """
        extract_binds_dict = {}

        for key, value in extract_binds.items():
            if not isinstance(value, utils.string_type):
                raise exception.ParamsError("invalid extract_binds!")

            try:
                value += "."
                # string.split(sep=None, maxsplit=-1) -> list of strings
                # e.g. "content.person.name" => ["content", "person.name"]
                top_query, sub_query = value.split(delimiter, 1)

                if top_query in ["body", "content", "text"]:
                    json_content = self.parsed_body()
                else:
                    json_content = getattr(self.resp_obj, top_query)

                if sub_query:
                    # e.g. key: resp_headers_content_type, sub_query = "content-type"
                    answer = utils.query_json(json_content, sub_query)
                    extract_binds_dict[key] = answer
                else:
                    # e.g. key: resp_status_code, resp_content
                    extract_binds_dict[key] = json_content

            except AttributeError:
                raise exception.ParamsError("invalid extract_binds!")

        return extract_binds_dict

    def validate(self, validators, variables_mapping):
        """ Bind named validators to value within the context.
        @param (dict) validators
            {
                "resp_status_code": {"comparator": "eq", "expected": 201},
                "resp_body_success": {"comparator": "eq", "expected": True}
            }
        @param (dict) variables_mapping
            {
                "resp_status_code": 200,
                "resp_body_success": True
            }
        @return (dict) content differences
            {
                "resp_status_code": {
                    "comparator": "eq", "expected": 201, "value": 200
                }
            }
        """
        diff_content_dict = {}

        for validator_key, validator_dict in validators.items():

            try:
                value = variables_mapping[validator_key]
                validator_dict["value"] = value
            except KeyError:
                raise exception.ParamsError("invalid validator %s" % validator_key)

            difference_exist = utils.compare(
                value,
                validator_dict["expected"],
                validator_dict["comparator"]
            )

            if difference_exist:
                diff_content_dict[validator_key] = validator_dict

        self.success = False if diff_content_dict else True
        return diff_content_dict
