import hashlib
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


def gen_random_string(str_len):
    return ''.join(
        random.choice(string.ascii_letters + string.digits) for _ in range(str_len))

def gen_md5(*args):
    args = [handle_req_data(item) for item in args]
    return hashlib.md5("".join(args).encode('utf-8')).hexdigest()

def handle_req_data(data):

    if PYTHON_VERSION == 3 and isinstance(data, bytes):
        # In Python3, convert bytes to str
        data = data.decode('utf-8')

    if not data:
        return data

    if isinstance(data, str):
        # check if data in str can be converted to dict
        try:
            data = json.loads(data)
        except ValueError:
            pass

    if isinstance(data, dict):
        # sort data in dict with keys, then convert to str
        data = json.dumps(data, sort_keys=True)

    return data

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
