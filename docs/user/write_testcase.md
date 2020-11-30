# Write Testcase

HttpRunner v3.x supports three testcase formats, `pytest`, `YAML` and `JSON`. It is extremely recommended to write and maintain testcases in `pytest` format instead of former `YAML/JSON` format.

The format relations are illustrated as below:

![](/images/httprunner-formats.png)

## record & generate testcase

If the SUT (system under test) is ready, the most efficient way is to capture HTTP traffic first and then generate testcases with HAR file. Refer to [`Record & Generate testcase`](/user/gen_tests/) for more details.

Based on the generated pytest testcase, you can then do some adjustment as needed, thus you need to know the details of testcase format.

## testcase structure

Each testcase is a subclass of `HttpRunner`, and must have two class attributes: `config` and `teststeps`.

- config: configure testcase level settings, including `base_url`, `verify`, `variables`, `export`.
- teststeps: list of teststep (`List[Step]`), each step is corresponding to a API request or another testcase reference call. Besides, `variables`/`extract`/`validate`/`hooks` mechanisms are supported to create extremely complex test scenarios.

```python
from httprunner import HttpRunner, Config, Step, RunRequest, RunTestCase


class TestCaseRequestWithFunctions(HttpRunner):
    config = (
        Config("request methods testcase with functions")
        .variables(
            **{
                "foo1": "config_bar1",
                "foo2": "config_bar2",
                "expect_foo1": "config_bar1",
                "expect_foo2": "config_bar2",
            }
        )
        .base_url("https://postman-echo.com")
        .verify(False)
        .export(*["foo3"])
    )

    teststeps = [
        Step(
            RunRequest("get with params")
            .with_variables(
                **{"foo1": "bar11", "foo2": "bar21", "sum_v": "${sum_two(1, 2)}"}
            )
            .get("/get")
            .with_params(**{"foo1": "$foo1", "foo2": "$foo2", "sum_v": "$sum_v"})
            .with_headers(**{"User-Agent": "HttpRunner/${get_httprunner_version()}"})
            .extract()
            .with_jmespath("body.args.foo2", "foo3")
            .validate()
            .assert_equal("status_code", 200)
            .assert_equal("body.args.foo1", "bar11")
            .assert_equal("body.args.sum_v", "3")
            .assert_equal("body.args.foo2", "bar21")
        ),
        Step(
            RunRequest("post form data")
            .with_variables(**{"foo2": "bar23"})
            .post("/post")
            .with_headers(
                **{
                    "User-Agent": "HttpRunner/${get_httprunner_version()}",
                    "Content-Type": "application/x-www-form-urlencoded",
                }
            )
            .with_data("foo1=$foo1&foo2=$foo2&foo3=$foo3")
            .validate()
            .assert_equal("status_code", 200)
            .assert_equal("body.form.foo1", "$expect_foo1")
            .assert_equal("body.form.foo2", "bar23")
            .assert_equal("body.form.foo3", "bar21")
        ),
    ]


if __name__ == "__main__":
    TestCaseRequestWithFunctions().test_start()
```

## chain call

One of the most awesome features of HttpRunner v3.x is `chain call`, with which you do not need to remember any testcase format details and you can get intelligent completion when you write testcases in IDE.

![](/images/httprunner-chain-call-config.png)

![](/images/httprunner-chain-call-step-validate.png)

## config

Each testcase should have one `config` part, in which you can configure testcase level settings.

### name (required)

Specify testcase name. This will be displayed in execution log and test report.

### base_url (optional)

Specify common schema and host part of the SUT, e.g. `https://postman-echo.com`. If `base_url` is specified, url in teststep can only set relative path part. This is especially useful if you want to switch between different SUT environments.

### variables (optional)

Specify common variables of testcase. Each teststep can reference config variable which is not set in step variables. In other words, step variables have higher priority than config variables.

### verify (optional)

Specify whether to verify the server’s TLS certificate. This is especially useful if we want to record HTTP traffic of testcase execution, because SSLError will be occurred if verify is not set or been set to True.

> SSLError(SSLCertVerificationError(1, '[SSL: CERTIFICATE_VERIFY_FAILED] certificate verify failed: self signed certificate in certificate chain (_ssl.c:1076)'))

### export (optional)

Specify the exported session variables of testcase. Consider each testcase as a black box, config `variables` is the input part, and config `export` is the output part. In particular, when a testcase is referenced in another testcase's step, and will be extracted some session variables to be used in subsequent teststeps, then the extracted session variables should be configured in config `export` part.

## teststeps

Each testcase should have one or multiple ordered test steps (`List[Step]`), each step is corresponding to a API request or another testcase reference call.

![](/images/httprunner-testcase.png)

> Notice: The concept of API in HttpRunner v2.x has been deprecated for simplification. You can consider API as a testcase that has only one request step.

### RunRequest(name)

`RunRequest` is used in a step to make request to API and do some extraction or validations for response.

The argument `name` of RunRequest is used to specify teststep name, which will be displayed in execution log and test report.

#### .with_variables

Specify teststep variables. The variables of each step are independent, thus if you want to share variables in multiple steps, you should define variables in config variables. Besides, the step variables will override the ones that have the same name in config variables.

#### .method(url)

Specify HTTP method and the url of SUT. These are corresponding to `method` and `url` arguments of [`requests.request`][requests.request].

If `base_url` is set in config, url can only set relative path part.

#### .with_params

Specify query string for the request url. This is corresponding to the `params` argument of [`requests.request`][requests.request].

#### .with_headers

Specify HTTP headers for the request. This is corresponding to the `headers` argument of [`requests.request`][requests.request].

#### .with_cookies

Specify HTTP request cookies. This is corresponding to the `cookies` argument of [`requests.request`][requests.request].

#### .with_data

Specify HTTP request body. This is corresponding to the `data` argument of [`requests.request`][requests.request].

#### .with_json

Specify HTTP request body in json. This is corresponding to the `json` argument of [`requests.request`][requests.request].

#### extract

##### .with_jmespath

Extract JSON response body with [jmespath][jmespath].

> with_jmespath(jmes_path: Text, var_name: Text)

- jmes_path: jmespath expression, refer to [JMESPath Tutorial][jmespath_tutorial] for more details
- var_name: the variable name that stores extracted value, it can be referenced by subsequent test steps

#### validate

##### .assert_XXX

Extract JSON response body with [jmespath][jmespath] and validate with expected value.

> assert_XXX(jmes_path: Text, expected_value: Any, message: Text = "")

- jmes_path: jmespath expression, refer to [JMESPath Tutorial][jmespath_tutorial] for more details
- expected_value: the specified expected value, variable or function reference can also be used here
- message (optional): used to indicate assertion error reason

The image below shows HttpRunner builtin validators.

![](/images/httprunner-step-request-validate.png)

### RunTestCase(name)

`RunTestCase` is used in a step to reference another testcase call.

The argument `name` of RunTestCase is used to specify teststep name, which will be displayed in execution log and test report.

#### .with_variables

Same with RunRequest's `.with_variables`.

#### .call

Specify referenced testcase class.

#### .export

Specify session variable names to export from referenced testcase. The exported variables can be referenced by subsequent test steps.

```python
import os
import sys

sys.path.insert(0, os.getcwd())

from httprunner import HttpRunner, Config, Step, RunRequest, RunTestCase

from examples.postman_echo.request_methods.request_with_functions_test import (
    TestCaseRequestWithFunctions as RequestWithFunctions,
)


class TestCaseRequestWithTestcaseReference(HttpRunner):
    config = (
        Config("request methods testcase: reference testcase")
        .variables(
            **{
                "foo1": "testsuite_config_bar1",
                "expect_foo1": "testsuite_config_bar1",
                "expect_foo2": "config_bar2",
            }
        )
        .base_url("https://postman-echo.com")
        .verify(False)
    )

    teststeps = [
        Step(
            RunTestCase("request with functions")
            .with_variables(
                **{"foo1": "testcase_ref_bar1", "expect_foo1": "testcase_ref_bar1"}
            )
            .call(RequestWithFunctions)
            .export(*["foo3"])
        ),
        Step(
            RunRequest("post form data")
            .with_variables(**{"foo1": "bar1"})
            .post("/post")
            .with_headers(
                **{
                    "User-Agent": "HttpRunner/${get_httprunner_version()}",
                    "Content-Type": "application/x-www-form-urlencoded",
                }
            )
            .with_data("foo1=$foo1&foo2=$foo3")
            .validate()
            .assert_equal("status_code", 200)
            .assert_equal("body.form.foo1", "bar1")
            .assert_equal("body.form.foo2", "bar21")
        ),
    ]


if __name__ == "__main__":
    TestCaseRequestWithTestcaseReference().test_start()
```


[requests.request]: https://requests.readthedocs.io/en/master/api/#requests.request
[jmespath]: https://jmespath.org/
[jmespath_tutorial]: https://jmespath.org/tutorial.html
