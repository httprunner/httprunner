# -*- coding: utf-8 -*-
from __future__ import absolute_import

import enum
import json

import thriftpy2
from loguru import logger
from thriftpy2.protocol import (
    TBinaryProtocolFactory,
    TCompactProtocolFactory,
    TCyBinaryProtocolFactory,
    TJSONProtocolFactory,
)
from thriftpy2.rpc import make_client
from thriftpy2.transport import (
    TBufferedTransportFactory,
    TCyBufferedTransportFactory,
    TCyFramedTransportFactory,
    TFramedTransportFactory,
)

from httprunner.thrift.data_convertor import json2thrift, thrift2dict


class ProtoType(enum.Enum):
    Binary = 1
    CyBinary = 2
    Compact = 3
    Json = 4


class TransType(enum.Enum):
    Buffered = 1
    CyBuffered = 2
    Framed = 3
    CyFramed = 4


class RequestFormat(enum.Enum):
    json = 1
    binary = 2


def get_proto_factory(proto_type):
    if proto_type == ProtoType.Binary:
        return TBinaryProtocolFactory()
    if proto_type == ProtoType.CyBinary:
        return TCyBinaryProtocolFactory()
    if proto_type == ProtoType.Compact:
        return TCompactProtocolFactory()
    if proto_type == ProtoType.Json:
        return TJSONProtocolFactory()


def get_trans_factory(trans_type):
    if trans_type == TransType.Buffered:
        return TBufferedTransportFactory()
    if trans_type == TransType.CyBuffered:
        return TCyBufferedTransportFactory()
    if trans_type == TransType.Framed:
        return TFramedTransportFactory()
    if trans_type == TransType.CyFramed:
        return TCyFramedTransportFactory()


class ThriftClient(object):
    def __init__(
        self,
        thrift_file,
        service_name,
        ip,
        port,
        include_dirs=None,
        timeout=3000,
        proto_type=ProtoType.CyBinary,
        trans_type=TransType.CyBuffered,
    ):
        self.thrift_file = thrift_file
        self.include_dirs = include_dirs
        self.service_name = service_name
        self.ip = ip
        self.port = port
        self.timeout = timeout
        self.proto_type = proto_type
        self.trans_type = trans_type
        try:
            logger.debug(
                "init thrift module: thrift_file=%s, module_name=%s",
                thrift_file,
                str(self.service_name) + "_thrift",
            )
            self.thrift_module = thriftpy2.load(
                self.thrift_file,
                module_name=str(self.service_name) + "_thrift",
                include_dirs=self.include_dirs,
            )
            self.thrift_service_obj = getattr(self.thrift_module, self.service_name)
            logger.debug(
                "init thrift client: service_name=%s, ip=%s, port=%s",
                self.thrift_service_obj,
                ip,
                port,
            )
            self.client = make_client(
                self.thrift_service_obj,
                self.ip,
                int(self.port),
                timeout=self.timeout,
                proto_factory=get_proto_factory(self.proto_type),
                trans_factory=get_trans_factory(self.trans_type),
            )
        except Exception as e:
            self.thrift_module = None
            self.thrift_service_obj = None
            self.client = None
            logger.exception("init thrift module and client failed: {}".format(e))
        finally:
            thriftpy2.parser.parser.thrift_stack = []

    def get_client(self):
        return self.client

    def send_request(self, request_data, request_method=""):
        thrift_req_cls = getattr(
            self.thrift_service_obj, request_method + "_args"
        ).thrift_spec[1][2]
        request_obj = json2thrift(json.dumps(request_data), thrift_req_cls)
        logger.debug(
            "send thrift request: request_method=%s, request_obj=%s",
            request_method,
            request_obj,
        )
        response_obj = getattr(self.client, request_method)(request_obj)
        logger.debug("thrift response = %s", response_obj)
        return thrift2dict(response_obj)

    def __del__(self):
        self.client.close()
