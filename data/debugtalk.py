import hashlib
import hmac
import random
import string
import time

SECRET_KEY = "DebugTalk"

def gen_random_string(str_len):
    random_char_list = []
    for _ in range(str_len):
        random_char = random.choice(string.ascii_letters + string.digits)
        random_char_list.append(random_char)

    random_string = ''.join(random_char_list)
    return random_string

def get_sign(*args):
    content = ''.join(args).encode('ascii')
    sign_key = SECRET_KEY.encode('ascii')
    sign = hmac.new(sign_key, content, hashlib.sha1).hexdigest()
    return sign

def gen_user_id():
    return int(time.time() * 1000)

def get_user_id():
    return [
        {"user_id": 1001},
        {"user_id": 1002},
        {"user_id": 1003},
        {"user_id": 1004}
    ]

def get_account(num):
    accounts = []
    for index in range(1, num+1):
        accounts.append(
            {"username": "user%s" % index, "password": str(index) * 6},
        )

    return accounts

def get_os_platform():
    return [
        {"os_platform": "ios"},
        {"os_platform": "android"}
    ]
