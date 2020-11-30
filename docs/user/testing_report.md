# Testing Report

Benefit from the integration of `pytest`, HttpRunner v3.x can make use of all the pytest plugins, including testing report plugins like `pytest-html` and `allure-pytest`.

## builtin html report

`pytest-html` plugin comes with HttpRunner installation. When you want to generate a html report for testcase execution, you can add a command argument `--html`.

```text
$ hrun /path/to/testcase --html=report.html
```

If you want to create a self-contained report, which is a single HTML file that can be more convenient when sharing results, you can add another command argument `--self-contained-html`.

```text
$ hrun /path/to/testcase --html=report.html --self-contained-html
```

You can refer to [`pytest-html`](https://pypi.org/project/pytest-html/) for more details.

## allure report

`allure-pytest` is an optional dependency for HttpRunner, thus if you want to generate allure report, you should install `allure-pytest` plugin separately.

```text
$ pip3 install "allure-pytest"
```

Or you can install HttpRunner with allure extra package.

```text
$ pip3 install "httprunner[allure]"
```

Once `allure-pytest` is ready, the following arguments can be used with `hrun/pytest` command.

- `--alluredir=DIR`: Generate Allure report in the specified directory (may not exist)
- `--clean-alluredir`: Clean alluredir folder if it exists
- `--allure-no-capture`: Do not attach pytest captured logging/stdout/stderr to report

To enable Allure listener to collect results during the test execution simply add `--alluredir` option and provide path to the folder where results should be stored. E.g.:

```text
$ hrun /path/to/testcase --alluredir=/tmp/my_allure_results
```

To see the actual report after your tests have finished, you need to use Allure commandline utility to generate report from the results.

```text
$ allure serve /tmp/my_allure_results
```

This command will show you generated report in your default browser.

You can refer to [`allure-pytest`](https://docs.qameta.io/allure/#_pytest) for more details.
