
对于上传文件类型的测试场景，HttpRunner 集成 [requests_toolbelt][1] 实现了上传功能。

在使用之前，确保已安装如下依赖库：

- [requests_toolbelt](https://github.com/requests/toolbelt)
- [filetype](https://github.com/h2non/filetype.py)

使用内置 `upload` 关键字，可轻松实现上传功能（适用版本：2.4.1+）。

```yaml
- test:
    name: upload file
    request:
        url: http://httpbin.org/upload
        method: POST
        headers:
            Cookie: session=AAA-BBB-CCC
        upload:
            file: "data/file_to_upload"
            field1: "value1"
            field2: "value2"
    validate:
        - eq: ["status_code", 200]
```

同时，你也可以继续使用之前描述形式（适用版本：2.0+）。

```yaml
- test:
    name: upload file
    variables:
        file: "data/file_to_upload"
        field1: "value1"
        field2: "value2"
        m_encoder: ${multipart_encoder(file=$file, field1=$field1, field2=$field2)}
    request:
        url: http://httpbin.org/upload
        method: POST
        headers:
            Content-Type: ${multipart_content_type($m_encoder)}
            Cookie: session=AAA-BBB-CCC
        data: $m_encoder
    validate:
        - eq: ["status_code", 200]
```

参考案例：[httprunner/tests/httpbin/upload.v2.yml][2]

[1]: https://toolbelt.readthedocs.io/en/latest/uploading-data.html
[2]: https://github.com/httprunner/httprunner/blob/master/tests/httpbin/upload.v2.yml