import configparser
import json

cf_parser = configparser.ConfigParser(allow_no_value=True)
cf_parser.read("pyproject.toml")
__version__ = json.loads(cf_parser["tool.poetry"]["version"])
__description__ = json.loads(cf_parser["tool.poetry"]["description"])

__all__ = ["__version__", "__description__"]
