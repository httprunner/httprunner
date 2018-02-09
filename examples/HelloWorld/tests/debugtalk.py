import hashlib
import hmac
import random
import string

SECRET_KEY = "DebugTalk"
default_request = {
    "base_url": "http://127.0.0.1:5000",
    "headers": {
        "Content-Type": "application/json",
        "device_sn": "$device_sn"
    }
}

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
