import json
import os
import random
import string
import time

from tests.api_server import HTTPBIN_SERVER, SECRET_KEY, gen_md5, get_sign

BASE_URL = "http://127.0.0.1:5000"


def sum_status_code(status_code, expect_sum):
    """ sum status code digits
        e.g. 400 => 4, 201 => 3
    """
    sum_value = 0
    for digit in str(status_code):
        sum_value += int(digit)

    assert sum_value == expect_sum

def is_status_code_200(status_code):
    return status_code == 200

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

def setup_hook_httpntlmauth(request):
    if "httpntlmauth" in request:
        from requests_ntlm import HttpNtlmAuth
        auth_account = request.pop("httpntlmauth")
        request["auth"] = HttpNtlmAuth(
            auth_account["username"], auth_account["password"])

def alter_response(response):
    response.status_code = 500
    response.headers["Content-Type"] = "html/text"
    response.json["headers"]["Host"] = "127.0.0.1:8888"
    response.new_attribute = "new_attribute_value"
    response.new_attribute_dict = {
        "key": 123
    }

def alter_response_error(response):
    # NameError
    not_defined_variable
