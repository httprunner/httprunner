import json
from flask import Flask
from flask import request, make_response

app = Flask(__name__)

""" storage all users' data
data structure:
    users_dict = {
        'uid1': {
            'uid': 'uid1',
            'name': 'name1',
            'password': 'pwd1'
        },
        'uid2': {
            'uid': 'uid2',
            'name': 'name2',
            'password': 'pwd2'
        }
    }
"""
users_dict = {}

@app.route('/api/user/clear')
def clear_users():
    users_dict.clear()
    return "ok"

@app.route('/api/user/add', methods=['POST'])
def add_user():
    user = request.get_json()
    users_dict[user["uid"]] = user
    return "ok"

@app.route('/api/user/<int:uid>')
def get_user(uid):
    user = users_dict.get(uid, {})
    response = make_response(json.dumps(user))
    response.headers["Content-Type"] = "application/json"
    return response

@app.route('/api/user/<int:uid>', methods=['DELETE'])
def delete_user(uid):
    user = users_dict.pop(uid, None)
    if user:
        return "ok"
    else:
        return "not_existed"
