# 常见问题

## HTTPS SSLError

请求 HTTPS 接口时，若本地开启了代理软件（Charles/Fiddler），由于 HTTPS 证书的原因，会导致 SSLError 的报错。

解决的方式是，在 config 中增加 `verify: False`，原理见 requests 的 [`SSL Cert Verification`](https://requests.kennethreitz.org/en/master/user/advanced/#ssl-cert-verification) 部分。

```yaml
config:
    name: XXX
    base_url: XXX
    verify: False
```
