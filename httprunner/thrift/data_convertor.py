# -*- coding: utf-8 -*-

from __future__ import division

import json
import traceback
import re
import logging
import base64

from thrift.Thrift import TType

try:
    from _json import encode_basestring_ascii as c_encode_basestring_ascii
except ImportError:
    c_encode_basestring_ascii = None

text_characters = "".join(map(chr, range(32, 127))) + "\n\r\t\b"
_null_trans = str.maketrans("", "")
ESCAPE = re.compile(r'[\x00-\x1f\\"\b\f\n\r\t]')
ESCAPE_ASCII = re.compile(r'([\\"]|[^\ -~])')
HAS_UTF8 = re.compile(r"[\x80-\xff]")
ESCAPE_DCT = {
    "\\": "\\\\",
    '"': '\\"',
    "\b": "\\b",
    "\f": "\\f",
    "\n": "\\n",
    "\r": "\\r",
    "\t": "\\t",
}
for i in range(0x20):
    ESCAPE_DCT.setdefault(chr(i), "\\u{0:04x}".format(i))
    # ESCAPE_DCT.setdefault(chr(i), '\\u%04x' % (i,))

INFINITY = float("inf")
FLOAT_REPR = repr


def istext(s_input):
    """
    既然我们要判断这串内容是不是可以做为Json的value,那为什么不放下试试呢？
    :param s_input:
    :return:
    """
    return not isinstance(s_input, bytes)


def unicode_2_utf8_keep_native(para):
    # if type(para) is str:
    #     return ''.join(filter(lambda x: not str.isalpha(x), para))
    if type(para) is str:
        return para

    if type(para) is list:
        for i in range(len(para)):
            para[i] = unicode_2_utf8_keep_native(para[i])
        return para
    elif type(para) is dict:
        newpara = {}
        for (key, value) in para.items():
            key = unicode_2_utf8_keep_native(key)
            value = unicode_2_utf8_keep_native(value)
            newpara[key] = value
        return newpara
    elif type(para) is tuple:
        return tuple(unicode_2_utf8_keep_native(list(para)))
    elif type(para) is str:
        return para.encode("utf-8")
    else:
        logging.debug("type========", type(para))
        # if issubclass(type(para), dict):
        if isinstance(para, dict):
            logging.debug("type ************in dict: %s" % (type(para)))
            return unicode_2_utf8_keep_native(dict(para))
        else:
            return para


def encode_basestring(s):
    """Return a JSON representation of a Python string"""

    def replace(match):
        return ESCAPE_DCT[match.group(0)]

    return '"' + ESCAPE.sub(replace, s) + '"'


def py_encode_basestring_ascii(s):
    """Return an ASCII-only JSON representation of a Python string"""
    if isinstance(s, str) and HAS_UTF8.search(s) is not None:
        s = s.decode("utf-8")

    def replace(match):
        s = match.group(0)
        try:
            return ESCAPE_DCT[s]
        except KeyError:
            n = ord(s)
            if n < 0x10000:
                return "\\u{0:04x}".format(n)
                # return '\\u%04x' % (n,)
            else:
                # surrogate pair
                n -= 0x10000
                s1 = 0xD800 | ((n >> 10) & 0x3FF)
                s2 = 0xDC00 | (n & 0x3FF)
                return "\\u{0:04x}\\u{1:04x}".format(s1, s2)
                # return '\\u%04x\\u%04x' % (s1, s2)

    return '"' + str(ESCAPE_ASCII.sub(replace, s)) + '"'


encode_basestring_ascii = c_encode_basestring_ascii or py_encode_basestring_ascii


class ThriftJSONDecoder(json.JSONDecoder):
    def __init__(self, *args, **kwargs):
        self._thrift_class = kwargs.pop("thrift_class")
        super(ThriftJSONDecoder, self).__init__(*args, **kwargs)

    def decode(self, json_str):
        if isinstance(json_str, dict):
            dct = json_str
        else:
            dct = super(ThriftJSONDecoder, self).decode(json_str)
        return self._convert(
            dct,
            TType.STRUCT,
            # (self._thrift_class, self._thrift_class.thrift_spec))
            self._thrift_class,
        )

    def _convert(self, val, ttype, ttype_info):
        if ttype == TType.STRUCT:
            if val is None:
                ret = None
            else:
                # (thrift_class, thrift_spec) = ttype_info
                thrift_class = ttype_info
                thrift_spec = ttype_info.thrift_spec
                ret = thrift_class()
                for tag, field in thrift_spec.items():
                    if field is None:
                        continue
                    # {1: (15, 'ad_ids', 10, False), 255: (12, 'Base', <class 'base.Base'>, False)}
                    # {1: (15, 'models', (12, <class 'adcommon.Ad'>), False), 255: (12, 'BaseResp', <class 'base.BaseResp'>, False)}
                    if len(field) <= 3:
                        (field_ttype, field_name, dummy) = field
                        field_ttype_info = None
                    else:
                        (field_ttype, field_name, field_ttype_info, dummy) = field

                    if val is None or field_name not in val:
                        continue
                    converted_val = self._convert(
                        val[field_name], field_ttype, field_ttype_info
                    )
                    setattr(ret, field_name, converted_val)
        elif ttype == TType.LIST:
            if type(ttype_info) != tuple:  # 说明是基础类型了, 无法在细分
                (element_ttype, element_ttype_info) = (ttype_info, None)
            else:
                (element_ttype, element_ttype_info) = ttype_info
            if val is not None:
                ret = [self._convert(x, element_ttype, element_ttype_info) for x in val]
            else:
                ret = None

        elif ttype == TType.SET:
            if type(ttype_info) != tuple:  # 说明是基础类型了, 无法在细分
                (element_ttype, element_ttype_info) = (ttype_info, None)
            else:
                (element_ttype, element_ttype_info) = ttype_info
            if val is not None:
                ret = set(
                    [self._convert(x, element_ttype, element_ttype_info) for x in val]
                )
            else:
                ret = None

        elif ttype == TType.MAP:
            # key处理
            if type(ttype_info[0]) == tuple:
                key_ttype, key_ttype_info = ttype_info[0]
            else:
                key_ttype, key_ttype_info = ttype_info[0], None

            # value处理
            if type(ttype_info[1]) != tuple:  # 说明value为基础类型, 已不可在细分
                val_ttype = ttype_info[1]
                val_ttype_info = None
            else:
                val_ttype, val_ttype_info = ttype_info[1]

            if val is not None:
                ret = dict(
                    [
                        (
                            self._convert(k, key_ttype, key_ttype_info),
                            self._convert(v, val_ttype, val_ttype_info),
                        )
                        for (k, v) in val.items()
                    ]
                )
            else:
                ret = None
        elif ttype == TType.STRING:
            if isinstance(val, str):
                ret = val.encode("utf8")
            elif val is None:
                ret = None
            else:
                ret = str(val)
            # 判断string字段是否是base64编码后的string, 如果是则此处需要对该string字段进行b64decode, 还原成原本的字符串
            # todo : 留待实现

        elif ttype == TType.DOUBLE:
            if val is not None:
                ret = float(val)
            else:
                ret = None
        elif ttype == TType.I64:
            if val is not None:
                ret = int(val)
            else:
                ret = None
        elif ttype == TType.I32 or ttype == TType.I16 or ttype == TType.BYTE:
            if val is not None:
                ret = int(val)
            else:
                ret = None
        elif ttype == TType.BOOL:
            if val is not None:
                ret = bool(val)
            else:
                ret = None
        else:
            raise TypeError("Unrecognized thrift field type: %s" % ttype)
        return ret


def json2thrift(json_str, thrift_class):
    logging.debug(json_str)
    return json.loads(
        json_str, cls=ThriftJSONDecoder, thrift_class=thrift_class, strict=False
    )


def dumper(obj):
    try:
        return json.dumps(obj, default=lambda o: o.__dict__, sort_keys=True, indent=2)
    except:
        return obj.__dict__


class MyJSONEncoder(json.JSONEncoder):
    def __init__(
        self,
        skipkeys=False,
        ensure_ascii=True,
        check_circular=True,
        allow_nan=True,
        indent=None,
        separators=None,
        encoding="utf-8",
        default=None,
        sort_keys=False,
        **kw
    ):
        super(MyJSONEncoder, self).__init__(
            skipkeys=skipkeys,
            ensure_ascii=ensure_ascii,
            check_circular=check_circular,
            allow_nan=allow_nan,
            indent=indent,
            separators=separators,
            encoding=encoding,
            default=default,
            sort_keys=sort_keys,
        )
        self.skip_nonutf8_value = kw.get(
            "skip_nonutf8_value", False
        )  # 默认不skip忽略非utf-8编码的字段

    def encode(self, o):
        """Return a JSON string representation of a Python data structure.
         JSONEncoder().encode({"foo": ["bar", "baz"]})
        '{"foo": ["bar", "baz"]}'

        """
        # This is for extremely simple cases and benchmarks.

        if isinstance(o, str):

            if isinstance(o, str):
                _encoding = self.encoding
                if _encoding is not None and not (_encoding == "utf-8"):
                    o = o.decode(_encoding)
            if self.ensure_ascii:
                return encode_basestring_ascii(o)
            else:
                return encode_basestring(o)
            # This doesn't pass the iterator directly to ''.join() because the
            # exceptions aren't as detailed.  The list call should be roughly
            # equivalent to the PySequence_Fast that ''.join() would do.
        chunks = self.iterencode(o, _one_shot=True)
        if not isinstance(chunks, (list, tuple)):
            chunks = list(chunks)
        # add by braver
        # todo: fix 'utf8' codec can't decode byte 0x91 in position 3: invalid start byte"
        if self.skip_nonutf8_value:  # 缺省为false
            tmp_chunks = []
            for chunk in chunks:
                try:
                    tmp_chunks.append(unicode_2_utf8_keep_native(chunk))
                except Exception as err:
                    logging.debug(traceback.format_exc())
            return "".join(tmp_chunks)

        # 保留老的逻辑, /usr/lib/python2.7/package/json/__init__.py dumps接口
        return "".join(chunks)


class ThriftJSONEncoder(json.JSONEncoder):
    """
    add by braver
    """

    def __init__(
        self,
        skipkeys=False,
        ensure_ascii=True,
        check_circular=True,
        allow_nan=True,
        indent=None,
        separators=None,
        default=None,
        sort_keys=False,
        **kw
    ):

        super(ThriftJSONEncoder, self).__init__(
            skipkeys=skipkeys,
            ensure_ascii=ensure_ascii,
            check_circular=check_circular,
            allow_nan=allow_nan,
            indent=indent,
            separators=separators,
            default=default,
            sort_keys=sort_keys,
        )
        self.skip_nonutf8_value = kw.get(
            "skip_nonutf8_value", False
        )  # 默认不skip忽略非utf-8编码的字段

    def encode(self, o):
        """Return a JSON string representation of a Python data structure.
         JSONEncoder().encode({"foo": ["bar", "baz"]})
        '{"foo": ["bar", "baz"]}'

        """
        # This is for extremely simple cases and benchmarks.

        if isinstance(o, str):
            if isinstance(o, str):
                _encoding = self.encoding
                if _encoding is not None and not (_encoding == "utf-8"):
                    o = o.decode(_encoding)
            if self.ensure_ascii:
                return encode_basestring_ascii(o)
            else:
                return encode_basestring(o)
            # This doesn't pass the iterator directly to ''.join() because the
            # exceptions aren't as detailed.  The list call should be roughly
            # equivalent to the PySequence_Fast that ''.join() would do.
        chunks = self.iterencode(o, _one_shot=True)
        if not isinstance(chunks, (list, tuple)):
            chunks = list(chunks)
        # add by braver
        # todo: fix 'utf8' codec can't decode byte 0x91 in position 3: invalid start byte"
        if self.skip_nonutf8_value:  # 缺省为false
            tmp_chunks = []
            for chunk in chunks:
                try:
                    tmp_chunks.append(unicode_2_utf8_keep_native(chunk))
                except Exception as err:
                    logging.debug(traceback.format_exc())
            return "".join(tmp_chunks)

        # 保留老的逻辑, /usr/lib/python2.7/package/json/__init__.py dumps接口
        return "".join(chunks)

    def default(self, o):
        if isinstance(o, bytes):
            return str(o, encoding="utf-8")
        if not hasattr(o, "thrift_spec"):
            return super(ThriftJSONEncoder, self).default(o)

        spec = getattr(o, "thrift_spec")
        ret = {}
        for tag, field in spec.items():
            if field is None:
                continue
            # (tag, field_ttype, field_name, field_ttype_info, default) = field
            field_name = field[1]
            default = field[-1]
            field_type = field[0]
            field_ttype_info = field[2]
            # if field_type in [TType.STRING, TType.BINARY]: # 说明是string(明文string或者binary)
            # if field_type in [TType.STRING, TType.BYTE]: # 说明是string(明文string或者binary)
            if field_name in o.__dict__:
                val = o.__dict__[field_name]
                if field_type in [TType.LIST, TType.SET]:  # 数组类型
                    if val:  # val为非空数组/Set
                        val = list(val)  # 统一转成数组(list/set)
                        is_need_binary_bs64 = False
                        if type(field_ttype_info) != tuple:  # 基础类型
                            if (
                                field_ttype_info in [TType.BYTE]
                                and type(val[0]) in [str]
                                and not istext(val[0])
                            ):
                                is_need_binary_bs64 = True
                        if is_need_binary_bs64:
                            for index, item in enumerate(val):
                                if item and type(item) in [str] and not istext(item):
                                    val[index] = base64.b64encode(
                                        item
                                    )  # 判断为二进制字符串, 需要进行base64编码
                if field_type in [TType.BYTE] and type(val) in [
                    str
                ]:  # 说明是string(明文string或者binary)
                    # 需要对二进制字节字符串字段进行base64编码, 将二进制字节串字段->ascii字符编码的base64编码明文串
                    if val and not istext(val):  # 说明是该字段非空且为binary string
                        print("4" * 100, val)
                        val = base64.b64encode(val.encode("utf-8"))
                        # val = base64.b64encode(val)  # 进行base64编码处理, 不然该字段序列化为json时会报错
                # if val != default:
                ret[field_name] = val
        if "request_id" in o.__dict__:
            ret["request_id"] = o.__dict__["request_id"]
        if "rpc_latency" in o.__dict__:
            ret["rpc_latency"] = o.__dict__["rpc_latency"]
        return ret


def thrift2json(obj, skip_nonutf8_value=False):
    return json.dumps(
        obj,
        cls=ThriftJSONEncoder,
        ensure_ascii=False,
        skip_nonutf8_value=skip_nonutf8_value,
    )


def thrift2dict(obj):
    str = thrift2json(obj)
    return json.loads(str)


dict2thrift = json2thrift

if __name__ == "__main__":
    print(istext("Всего за {$price$}, а доставка - бесплатно!"))
    print(istext(b"\xe4\xb8\xad\xe6\x96\x87"))
    print(
        istext(
            '{"web_uri":"ad-site-i18n-sg/202103185d0d723d88b7f642452dac73","height":336,"width":336,"file_name":""}'
        )
    )
