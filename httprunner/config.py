import inspect
from typing import Text

from httprunner.models import TConfig, TConfigThrift


class ConfigThrift(object):
    def __init__(self, config: TConfig) -> None:
        self.__config = config
        self.__config.thrift = TConfigThrift()

    def psm(self, psm: Text) -> "ConfigThrift":
        self.__config.thrift.psm = psm
        return self

    def env(self, env: Text) -> "ConfigThrift":
        self.__config.thrift.env = env
        return self

    def cluster(self, cluster: Text) -> "ConfigThrift":
        self.__config.thrift.cluster = cluster
        return self

    def target(self, target: Text) -> "ConfigThrift":
        self.__config.thrift.target = target
        return self

    def struct(self) -> TConfig:
        return self.__config


class Config(object):
    def __init__(self, name: Text) -> None:
        caller_frame = inspect.stack()[1]
        self.__config = TConfig(name=name, path=caller_frame.filename)

    @property
    def name(self) -> Text:
        return self.__config.name

    @property
    def path(self) -> Text:
        return self.__config.path

    def variables(self, **variables) -> "Config":
        self.__config.variables.update(variables)
        return self

    def base_url(self, base_url: Text) -> "Config":
        self.__config.base_url = base_url
        return self

    def verify(self, verify: bool) -> "Config":
        self.__config.verify = verify
        return self

    def export(self, *export_var_name: Text) -> "Config":
        self.__config.export.extend(export_var_name)
        self.__config.export = list(set(self.__config.export))
        return self

    def struct(self) -> TConfig:
        return self.__config

    def thrift(self) -> ConfigThrift:
        return ConfigThrift(self.__config)
