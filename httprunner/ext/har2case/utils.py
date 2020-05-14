import io
import json
import logging
import sys
from json.decoder import JSONDecodeError
from urllib.parse import unquote

import yaml


def load_har_log_entries(file_path):
    """ load HAR file and return log entries list

    Args:
        file_path (str)

    Returns:
        list: entries
            [
                {
                    "request": {},
                    "response": {}
                },
                {
                    "request": {},
                    "response": {}
                }
            ]

    """
    with io.open(file_path, "r+", encoding="utf-8-sig") as f:
        try:
            content_json = json.loads(f.read())
            return content_json["log"]["entries"]
        except (KeyError, TypeError, JSONDecodeError):
            logging.error("HAR file content error: {}".format(file_path))
            sys.exit(1)


def x_www_form_urlencoded(post_data):
    """ convert origin dict to x-www-form-urlencoded

    Args:
        post_data (dict):
            {"a": 1, "b":2}

    Returns:
        str:
            a=1&b=2

    """
    if isinstance(post_data, dict):
        return "&".join(
            [u"{}={}".format(key, value) for key, value in post_data.items()]
        )
    else:
        return post_data


def convert_x_www_form_urlencoded_to_dict(post_data):
    """ convert x_www_form_urlencoded data to dict

    Args:
        post_data (str): a=1&b=2

    Returns:
        dict: {"a":1, "b":2}

    """
    if isinstance(post_data, str):
        converted_dict = {}
        for k_v in post_data.split("&"):
            try:
                key, value = k_v.split("=")
            except ValueError:
                raise Exception(
                    "Invalid x_www_form_urlencoded data format: {}".format(post_data)
                )
            converted_dict[key] = unquote(value)
        return converted_dict
    else:
        return post_data


def convert_list_to_dict(origin_list):
    """ convert HAR data list to mapping

    Args:
        origin_list (list)
            [
                {"name": "v", "value": "1"},
                {"name": "w", "value": "2"}
            ]

    Returns:
        dict:
            {"v": "1", "w": "2"}

    """
    return {item["name"]: item.get("value") for item in origin_list}


def dump_yaml(testcase, yaml_file):
    """ dump HAR entries to yaml testcase
    """
    logging.info("dump testcase to YAML format.")

    with io.open(yaml_file, "w", encoding="utf-8") as outfile:
        yaml.dump(
            testcase, outfile, allow_unicode=True, default_flow_style=False, indent=4
        )

    logging.info("Generate YAML testcase successfully: {}".format(yaml_file))


def dump_json(testcase, json_file):
    """ dump HAR entries to json testcase
    """
    logging.info("dump testcase to JSON format.")

    with io.open(json_file, "w", encoding="utf-8") as outfile:
        my_json_str = json.dumps(testcase, ensure_ascii=False, indent=4)
        if isinstance(my_json_str, bytes):
            my_json_str = my_json_str.decode("utf-8")

        outfile.write(my_json_str)

    logging.info("Generate JSON testcase successfully: {}".format(json_file))
