## 测试用例（testcase）

从 2.0 版本开始，HttpRunner 开始对测试用例的定义进行进一步的明确，参考 [wiki][wiki_testcase] 上的描述。

> A test case is a specification of the inputs, execution conditions, testing procedure, and expected results that define a single test to be executed to achieve a particular software testing objective, such as to exercise a particular program path or to verify compliance with a specific requirement.

概括下来，一条测试用例（testcase）应该是为了测试某个特定的功能逻辑而精心设计的，并且至少包含如下几点：

- 明确的测试目的（achieve a particular software testing objective）
- 明确的输入（inputs）
- 明确的运行环境（execution conditions）
- 明确的测试步骤描述（testing procedure）
- 明确的预期结果（expected results）

对应地，HttpRunner 的测试用例描述方式进行如下设计：

- 测试用例应该是完整且独立的，每条测试用例应该是都可以独立运行的；在 HttpRunner 中，每个 `YAML/JSON` 文件对应一条测试用例。
- 测试用例包含 `测试脚本` 和 `测试数据` 两部分：
    - `测试用例 = 测试脚本 + 测试数据`
    - `测试脚本` 重点是描述测试的 `业务功能逻辑`，包括预置条件、测试步骤、预期结果等，并且可以结合辅助函数（debugtalk.py）实现复杂的运算逻辑；可以将 `测试脚本` 理解为编程语言中的 `类（class）`；
    - `测试数据` 重点是对应测试的 `业务数据逻辑`，可以理解为类的实例化数据；
    - `测试数据` 和 `测试脚本` 分离后，就可以比较方便地实现数据驱动测试，通过对测试脚本传入一组数据，实现同一业务功能在不同数据逻辑下的测试验证。


## 测试步骤（teststep）

测试用例是测试步骤的 `有序` 集合，而对于接口测试来说，每一个测试步骤应该就对应一个 API 的请求描述。

## 测试用例集（testsuite）

`测试用例集` 是 `测试用例` 的 `无序` 集合，集合中的测试用例应该都是相互独立，不存在先后依赖关系的。

如果确实存在先后依赖关系怎么办，例如登录功能和下单功能。正确的做法应该是，在下单测试用例的前置步骤中执行登录操作。

```yaml
- config:
    name: order product

- test:
    name: login
    testcase: testcases/login.yml

- test:
    name: add to cart
    api: api/add_cart.yml

- test:
    name: make order
    api: api/make_order.yml
```

## 测试场景

`测试场景` 和 `测试用例集` 是同一概念，都是 `测试用例` 的 `无序` 集合。


- 接口
- 测试用例集
- 参数
- 变量
- 测试脚本（YAML/JSON）
- debugtalk.py
- 环境变量

## 项目根目录

[wiki_testcase]: https://en.wikipedia.org/wiki/Test_case