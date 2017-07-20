import hashlib
import json
from functools import wraps

from flask import Flask, make_response, request
from ate import utils

app = Flask(__name__)

""" storage all users' data
data structure:
    users_dict = {
        'uid1': {
            'name': 'name1',
            'password': 'pwd1'
        },
        'uid2': {
            'name': 'name2',
            'password': 'pwd2'
        }
    }
"""
users_dict = {}

AUTHENTICATION = False
TOKEN = "debugtalk"

def validate_request(func):

    @wraps(func)
    def wrapper(*args, **kwds):
        if not AUTHENTICATION:
            return func(*args, **kwds)

        try:
            req_headers = request.headers
            req_authorization = req_headers['Authorization']
            random_str = req_headers['Random']
            data = utils.handle_req_data(request.data)
            authorization = utils.gen_md5(TOKEN, data, random_str)
            assert authorization == req_authorization
            return func(*args, **kwds)
        except (KeyError, AssertionError):
            result = {
                'success': False,
                'msg': "Authorization failed!"
            }
            response = make_response(json.dumps(result), 403)
            response.headers["Content-Type"] = "application/json"
            return response

    return wrapper


@app.route('/')
@validate_request
def index():
    return "Hello World!"

@app.route('/customize-response', methods=['POST'])
@validate_request
def get_customized_response():
    expected_resp_json = request.get_json()
    status_code = expected_resp_json.get('status_code', 200)
    headers_dict = expected_resp_json.get('headers', {})
    body = expected_resp_json.get('body', {})
    response = make_response(json.dumps(body), status_code)

    for header_key, header_value in headers_dict.items():
        response.headers[header_key] = header_value

    return response

@app.route('/api/token')
@validate_request
def get_token():
    result = {
        'success': True,
        'token': utils.gen_random_string(8)
    }
    response = make_response(json.dumps(result))
    response.headers["Content-Type"] = "application/json"
    return response

@app.route('/api/users')
@validate_request
def get_users():
    users_list = [user for uid, user in users_dict.items()]
    users = {
        'success': True,
        'count': len(users_list),
        'items': users_list
    }
    response = make_response(json.dumps(users))
    response.headers["Content-Type"] = "application/json"
    return response

@app.route('/api/users', methods=['DELETE'])
@validate_request
def clear_users():
    users_dict.clear()
    result = {
        'success': True
    }
    response = make_response(json.dumps(result))
    response.headers["Content-Type"] = "application/json"
    return response

@app.route('/api/users/<int:uid>', methods=['POST'])
@validate_request
def create_user(uid):
    user = request.get_json()
    if uid not in users_dict:
        result = {
            'success': True,
            'msg': "user created successfully."
        }
        status_code = 201
        users_dict[uid] = user
    else:
        result = {
            'success': False,
            'msg': "user already existed."
        }
        status_code = 500

    response = make_response(json.dumps(result), status_code)
    response.headers["Content-Type"] = "application/json"
    return response

@app.route('/api/users/<int:uid>')
@validate_request
def get_user(uid):
    user = users_dict.get(uid, {})
    if user:
        result = {
            'success': True,
            'data': user
        }
        status_code = 200
    else:
        result = {
            'success': False,
            'data': user
        }
        status_code = 404

    response = make_response(json.dumps(result), status_code)
    response.headers["Content-Type"] = "application/json"
    return response

@app.route('/api/users/<int:uid>', methods=['PUT'])
@validate_request
def update_user(uid):
    user = users_dict.get(uid, {})
    if user:
        user = request.get_json()
        success = True
        status_code = 200
    else:
        success = False
        status_code = 404

    result = {
        'success': success,
        'data': user
    }
    response = make_response(json.dumps(result), status_code)
    response.headers["Content-Type"] = "application/json"
    return response

@app.route('/api/users/<int:uid>', methods=['DELETE'])
@validate_request
def delete_user(uid):
    user = users_dict.pop(uid, {})
    if user:
        success = True
        status_code = 200
    else:
        success = False
        status_code = 404

    result = {
        'success': success,
        'data': user
    }
    response = make_response(json.dumps(result), status_code)
    response.headers["Content-Type"] = "application/json"
    return response
