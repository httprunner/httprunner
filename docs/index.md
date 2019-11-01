
HttpRunner 是一款面向 HTTP(S) 协议的通用测试框架，只需编写维护一份 `YAML/JSON` 脚本，即可实现自动化测试、性能测试、线上监控、持续集成等多种测试需求。

此文档适用于全新发布的 `HttpRunner 2.x` 版本，`1.x` 版本的使用文档请查看[历史链接][httprunner1]。

## 设计理念

- 充分复用优秀的开源项目，不追求重复造轮子，而是将强大的轮子组装成战车
- 遵循 `约定大于配置` 的准则，在框架功能中融入自动化测试最佳工程实践
- 追求投入产出比，一份投入即可实现多种测试需求

## 核心特性

- 继承 [Requests][Requests] 的全部特性，轻松实现 HTTP(S) 的各种测试需求
- 采用 `YAML/JSON` 的形式描述测试场景，保障测试用例描述的统一性和可维护性
- 借助辅助函数（debugtalk.py），在测试脚本中轻松实现复杂的动态计算逻辑
- 支持完善的测试用例分层机制，充分实现测试用例的复用
- 测试前后支持完善的 hook 机制
- 响应结果支持丰富的校验机制
- 基于 HAR 实现接口录制和用例生成功能（[har2case][har2case]）
- 结合 [Locust][Locust] 框架，无需额外的工作即可实现分布式性能测试
- 执行方式采用 CLI 调用，可与 Jenkins 等持续集成工具完美结合
- 测试结果统计报告简洁清晰，附带详尽统计信息和日志记录
- 极强的可扩展性，轻松实现二次开发和 Web 平台化

## 更多信息

关注 HttpRunner 的微信公众号，第一时间获得最新资讯。

![](./assets/qrcode.jpg)


[httprunner1]: https://v1.httprunner.org/
[Requests]: http://docs.python-requests.org/en/master/
[Locust]: http://locust.io/
[har2case]: https://github.com/HttpRunner/har2case