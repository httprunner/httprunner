
## 版本号（Version）

从 2.0 版本开始，HttpRunner 开始使用 [`Semantic Versioning`][SemVer] 版本号机制。该机制由 GitHub 联合创始人 Tom Preston-Werner 编写，当前被广泛采用，遵循该机制也可以更好地与开源生态统一，避免出现 “dependency hell” 的情况。

具体地，HttpRunner 将采用 `MAJOR.MINOR.PATCH` 的版本号机制。

- MAJOR: 重大版本升级并出现前后版本不兼容时加 1
- MINOR: 大版本内新增功能并且保持版本内兼容性时加 1
- PATCH: 功能迭代过程中进行问题修复（bugfix）时加 1

当然，在实际迭代开发过程中，肯定也不会每次提交（commit）都对 PATCH 加 1；在遵循如上主体原则的前提下，也会根据需要，在版本号后面添加先行版本号（-alpha/beta/rc）或版本编译元数据（+20190101）作为延伸。

## 分支策略

HttpRunner 的开发分支策略采用 GitHub Flow。

![](images/github-flow.png)

## 提交信息（Commit Message）

代码提交的注释信息遵循如下格式规范：

```xml
<type>(<scope>): <subject>
<BLANK LINE>
<body>
<BLANK LINE>
<footer>
```

- **type**【必填】，大致分类如下：
    - feat：新功能（feature）
    - fix：修补 bug
    - docs：文档（documentation）
    - style： 格式（不影响代码运行的变动）
    - perf：性能提升
    - refactor：重构（即不是新增功能，也不是修改 bug 的代码变动）
    - test：增加测试
    - build：构建过程或辅助工具的变动
- **subject**【必填】，对提交内容的简要概述
- scope【可选项】，用于说明 commit 影响的范围，视项目而定，一般建议写对应具体模块
- body【可选项】，对提交内容的详细描述
- footer【可选项】，一般为BREAKING CHANGE或关联的issue等内容 详见参考文档


[SemVer]: https://semver.org/
