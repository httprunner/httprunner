import json
import yaml
import os.path
from ate.exception import ParamsError

def load_yaml_file(yaml_file):
    with open(yaml_file, 'r+') as stream:
        return yaml.load(stream)

def load_json_file(json_file):
    with open(json_file) as data_file:
        return json.load(data_file)

def load_testcases(testcase_file_path):
    file_suffix = os.path.splitext(testcase_file_path)[1]
    if file_suffix == '.json':
        return load_json_file(testcase_file_path)
    elif file_suffix in ['.yaml', '.yml']:
        return load_yaml_file(testcase_file_path)
    else:
        # '' or other suffix
        raise ParamsError("Bad testcase file name!")

def parse_response_object(resp_obj):
    try:
        resp_content = resp_obj.json()
    except ValueError:
        resp_content = resp_obj.text

    return {
        'status_code': resp_obj.status_code,
        'headers': resp_obj.headers,
        'content': resp_content
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

    return diff_content
