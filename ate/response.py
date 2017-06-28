from ate import utils, exception


def parse_response_body(resp_obj):
    try:
        return resp_obj.json()
    except ValueError:
        return resp_obj.text

def parse_response_object(resp_obj):
    return {
        'status_code': resp_obj.status_code,
        'headers': resp_obj.headers,
        'body': parse_response_body(resp_obj)
    }

def diff_response(resp_obj, expected_resp_json):
    diff_content = {}
    resp_info = parse_response_object(resp_obj)

    expected_status_code = expected_resp_json.get('status_code', 200)
    if resp_info['status_code'] != int(expected_status_code):
        diff_content['status_code'] = {
            'value': resp_info['status_code'],
            'expected': expected_status_code
        }

    expected_headers = expected_resp_json.get('headers', {})
    headers_diff = utils.diff_json(resp_info['headers'], expected_headers)
    if headers_diff:
        diff_content['headers'] = headers_diff

    expected_body = expected_resp_json.get('body', None)

    if expected_body is None:
        body_diff = {}
    elif type(expected_body) != type(resp_info['body']):
        body_diff = {
            'value': resp_info['body'],
            'expected': expected_body
        }
    elif isinstance(expected_body, str):
        if expected_body != resp_info['body']:
            body_diff = {
                'value': resp_info['body'],
                'expected': expected_body
            }
    elif isinstance(expected_body, dict):
        body_diff = utils.diff_json(resp_info['body'], expected_body)

    if body_diff:
        diff_content['body'] = body_diff

    return diff_content

def extract_response(resp_obj, context, delimiter='.'):
    """ extract content from requests.Response, and bind extracted value to context.extractors
    @param (requests.Response instance) resp_obj
    @param (ate.context.Context instance) context
        context.extractors:
        {
            "resp_status_code": "status_code",
            "resp_headers_content_type": "headers.content-type",
            "resp_content": "content",
            "resp_content_person_first_name": "content.person.name.first_name"
        }
    """
    for key, value in context.extractors.items():
        try:
            if isinstance(value, str):
                value += "."
                # string.split(sep=None, maxsplit=-1) -> list of strings
                # e.g. "content.person.name" => ["content", "person.name"]
                top_query, sub_query = value.split(delimiter, 1)

                if top_query in ["body", "content", "text"]:
                    json_content = parse_response_body(resp_obj)
                else:
                    json_content = getattr(resp_obj, top_query)

                if sub_query:
                    # e.g. key: resp_headers_content_type, sub_query = "content-type"
                    answer = utils.query_json(json_content, sub_query)
                    context.extractors[key] = answer
                else:
                    # e.g. key: resp_status_code, resp_content
                    context.extractors[key] = json_content

            else:
                raise NotImplementedError("TODO: support template.")

        except AttributeError:
            raise exception.ParamsError("invalid extract_binds!")
