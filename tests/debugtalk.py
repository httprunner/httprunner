import hashlib
import hmac
import json
import os
import random
import string
import time

try:
    import urllib
except NameError:
    import urllib.parse as urllib

SECRET_KEY = "DebugTalk"
BASE_URL = "http://127.0.0.1:5000"

def get_sign(*args):
    content = ''.join(args).encode('ascii')
    sign_key = SECRET_KEY.encode('ascii')
    sign = hmac.new(sign_key, content, hashlib.sha1).hexdigest()
    return sign

get_sign_lambda = lambda *args: hmac.new(
    'DebugTalk'.encode('ascii'),
    ''.join(args).encode('ascii'),
    hashlib.sha1).hexdigest()

def gen_md5(*args):
    return hashlib.md5("".join(args).encode('utf-8')).hexdigest()

def sum_status_code(status_code, expect_sum):
    """ sum status code digits
        e.g. 400 => 4, 201 => 3
    """
    sum_value = 0
    for digit in str(status_code):
        sum_value += int(digit)

    assert sum_value == expect_sum

os.environ["TEST_ENV"] = "PRODUCTION"

def skip_test_in_production_env():
    """ skip this test in production environment
    """
    return os.environ["TEST_ENV"] == "PRODUCTION"

def gen_app_version():
    return [
        {"app_version": "2.8.5"},
        {"app_version": "2.8.6"}
    ]

def get_account():
    return [
        {"username": "user1", "password": "111111"},
        {"username": "user2", "password": "222222"}
    ]

SECRET_KEY = "DebugTalk"

def gen_random_string(str_len):
    random_char_list = []
    for _ in range(str_len):
        random_char = random.choice(string.ascii_letters + string.digits)
        random_char_list.append(random_char)

    random_string = ''.join(random_char_list)
    return random_string

def setup_hook_add_kwargs(request):
    request["key"] = "value"

def setup_hook_remove_kwargs(request):
    request.pop("key")

def teardown_hook_sleep_N_secs(response, n_secs):
    """ sleep n seconds after request
    """
    if response.status_code == 200:
        time.sleep(0.1)
    else:
        time.sleep(n_secs)

def hook_print(msg):
    print(msg)

def modify_headers_os_platform(request, os_platform):
    request["headers"]["os_platform"] = os_platform
