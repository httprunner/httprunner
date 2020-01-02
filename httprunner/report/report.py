import json
import platform
import time
import uuid

import requests

from httprunner import __version__


def prepare_event_kwargs(event_name, params):
    """ prepare report event kwargs"""

    kwargs = {
        "headers": {
            'content-type': 'application/json'
        },
        "json": {
            "user": {
                "user_unique_id": str(uuid.getnode())
            },
            "header": {
                "app_id": 173519,
                "os_name": platform.system(),
                "os_version": platform.release(),
                "app_version": __version__  # HttpRunner version
            },
            "events": [
                {
                    "event": event_name,
                    "params": json.dumps(params),
                    "time": int(time.time())
                }
            ],
            "verbose": 1
        }
    }
    return kwargs


def report_event(event_name, success=True):
    params = {
        "success": 1 if success else 0
    }
    kwargs = prepare_event_kwargs(event_name, params)
    resp = requests.post("http://mcs.snssdk.com/v1/json", **kwargs)
    print("resp---", resp.json())


if __name__ == '__main__':
    report_event("loader")
