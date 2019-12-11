## 案例介绍

- 被测案例：[klook](https://www.klook.com/)
- 案例作者：[readyou](https://github.com/readyou)

我们团队选择了 HttpRunner 作为接口测试框架，并整理了一份案例，供大家参考。

## 注意事项

1. 本例子中有些地方用到了`localhost:8085`作为base_url，这些接口是不能访问的，仅作为示例学习怎样组织测试用例。
2. `https://maps.googleapis.com`是可以用的，自己申请一个key，替换掉文件中的`your_google_map_key`即可。

## 相关文件说明

模块 | 文件 | 用途 | 备注
---|----|------|------
google map 接口测试 | api/find_place_api.yml | google map根据名称搜索地址的api | 比较全面地使用了api可以使用的关键字：name, base_url, request, variables, validate, extract
google map 接口测试 | testcases/find_place_testcase.yml | google map根据名称搜索地址的testcase | 使用了testcase标准的写法：testcase由teststep组成，teststep中引用api(just_request_testcase.yml中演示了直接使用request而不是引用api的方式)。teststep中还使用了variables。
google map 接口测试 | testcases/place_detail_testcase.yml | google map获取地址详情的testcase | config中使用variables
google map 接口测试 | testsuites/place_detail_testsuite.yml | google map接口测试的testsuite包含上面两个testcase | 使用了多种方式来做数据驱动测试
 | | 
klook地理位置搜索接口测试 | api/search_area_by_name_api.yml | 根据名字查询区域（支持多语言）——api | 
klook地理位置搜索接口测试 | api/search_area_by_name_testcase.yml | 根据名字查询区域（支持多语言）——testcase |
klook地理位置搜索接口测试 | api/get_area_groups_api.yml | 查询地理位置下面的组——api |
klook地理位置搜索接口测试 | api/get_area_groups_testcase.yml | 查询地理位置下面的组——testcase |
klook地理位置搜索接口测试 | api/area_manage_testsuite.yml | 区域管理——testsuite |
 | | 
baidu首页demo | testcases/just_request_testcase.yml | 提取百度首页title的demo | 演示了直接使用request而不是引用api的方式，使用了teardown_hooks的使用

完整的案例访问[地址](https://github.com/httprunner/httprunner/tree/master/docs/examples/demo-klook)。
