import copy
import inspect
from typing import Text

from httprunner.models import TConfig, TConfigThrift, TConfigDB, ProtoType, VariablesMapping


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

    def service_name(self, service_name: Text) -> "ConfigThrift":
        self.__config.thrift.service_name = service_name
        return self

    def method(self, method: Text) -> "ConfigThrift":
        self.__config.thrift.method = method
        return self

    def ip(self, service_name_: Text) -> "ConfigThrift":
        self.__config.thrift.service_name = service_name_
        return self

    def port(self, port: int) -> "ConfigThrift":
        self.__config.thrift.port = port
        return self

    def timeout(self, timeout: int) -> "ConfigThrift":
        self.__config.thrift.timeout = timeout
        return self

    def proto_type(self, proto_type: ProtoType) -> "ConfigThrift":
        self.__config.thrift.proto_type = proto_type
        return self

    def trans_type(self, trans_type: ProtoType) -> "ConfigThrift":
        self.__config.thrift.trans_type = trans_type
        return self

    def struct(self) -> TConfig:
        return self.__config


class ConfigDB(object):
    def __init__(self, config: TConfig):
        self.__config = config
        self.__config.db = TConfigDB()

    def psm(self, psm):
        self.__config.db.psm = psm
        return self

    def user(self, user):
        self.__config.db.user = user
        return self

    def password(self, password):
        self.__config.db.password = password
        return self

    def ip(self, ip):
        self.__config.db.ip = ip
        return self

    def port(self, port: int):
        self.__config.db.port = port
        return self

    def database(self, database: Text):
        self.__config.db.database = database
        return self

    def struct(self) -> TConfig:
        return self.__config


class Config(object):
    def __init__(self, name: Text) -> None:
        caller_frame = inspect.stack()[1]
        self.__name: Text = name
        self.__base_url: Text = ""
        self.__variables: VariablesMapping = {}
        self.__config = TConfig(name=name, path=caller_frame.filename)

    @property
    def name(self) -> Text:
        return self.__config.name

    @property
    def path(self) -> Text:
        return self.__config.path

    def variables(self, **variables) -> "Config":
        self.__variables.update(variables)
        return self

    def base_url(self, base_url: Text) -> "Config":
        self.__base_url = base_url
        return self

    def verify(self, verify: bool) -> "Config":
        self.__config.verify = verify
        return self

    def export(self, *export_var_name: Text) -> "Config":
        self.__config.export.extend(export_var_name)
        self.__config.export = list(set(self.__config.export))
        return self

    def struct(self) -> TConfig:
        self.__init()
        return self.__config

    def thrift(self) -> ConfigThrift:
        self.__init()
        return ConfigThrift(self.__config)

    def db(self) -> ConfigDB:
        self.__init()
        return ConfigDB(self.__config)

    def __init(self) -> None:
        self.__config.name = self.__name
        self.__config.base_url = self.__base_url
        self.__config.variables = copy.copy(self.__variables)
