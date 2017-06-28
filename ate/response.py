from ate import utils


def parse_response_object(resp_obj):
    try:
        resp_body = resp_obj.json()
    except ValueError:
        resp_body = resp_obj.text

    return {
        'status_code': resp_obj.status_code,
        'headers': resp_obj.headers,
        'body': resp_body
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
