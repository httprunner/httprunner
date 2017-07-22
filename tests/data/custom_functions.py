import hashlib
import hmac
import json
import random
import string
import time

try:
    string_type = basestring
    PYTHON_VERSION = 2
    import urllib
except NameError:
    string_type = str
    PYTHON_VERSION = 3
    import urllib.parse as urllib

SECRET_KEY = "DebugTalk"

def gen_random_string(str_len):
    random_char_list = []
    for _ in range(str_len):
        random_char = random.choice(string.ascii_letters + string.digits)
        random_char_list.append(random_char)

    random_string = ''.join(random_char_list)
    return random_string

gen_random_string_lambda = lambda str_len: ''.join(
    random.choice(string.ascii_letters + string.digits) for _ in range(str_len))

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

def gen_urlencode_str(**kargs):
    urlencoded_str = ""
    quote_times = int(kargs.pop("quote_times", 1))

    for key, value in kargs.items():
        urlencoded_str += key
        urlencoded_str += "="
        if value == "undefined":
            urlencoded_str += "undefined"
        else:
            if isinstance(value, (dict, list)):
                value = json.dumps(value)
            elif isinstance(value, (int, float)):
                value = str(value)

            value_str = value.encode('utf-8')
            for _ in range(quote_times):
                value_str = urllib.quote_plus(value_str)
            urlencoded_str += value_str

        urlencoded_str += "&"

    return urlencoded_str.strip("&")

def get_timestamp():
    return int(time.time() * 1000)
