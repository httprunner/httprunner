# HttpRunner V2.x 中文使用文档

## 在线阅读

本文档托管在`GitHub Pages`上，访问地址：

https://docs.httprunner.org

## 本地预览

### 安装依赖

本项目文档采用[`mkdocs`][mkdocs]生成，如需在本地预览查看，则需安装该工具。

```bash
$ pip install mkdocs
```

`mkdocs`支持主题配置，本项目选择了第三方的[`mkdocs-material`][mkdocs-material]。

```bash
$ pip install mkdocs-material
```

### 启动本地server

在项目根目录中运行如下命令：

```bash
$ mkdocs serve
INFO    -  Building documentation...
INFO    -  Cleaning site directory
[I 180211 22:48:35 server:283] Serving on http://127.0.0.1:8000
[I 180211 22:48:35 handlers:60] Start watching changes
[I 180211 22:48:35 handlers:62] Start detecting changes
```

然后在浏览器中访问`http://127.0.0.1:8000`即可。

[mkdocs]: http://www.mkdocs.org/
[mkdocs-material]: https://squidfunk.github.io/mkdocs-material/