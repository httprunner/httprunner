## Unable to install PyUnitReport dependency library automatically

If there is something goes wrong in installation like below.

```text
Downloading/unpacking PyUnitReport (from HttpRunner)
  Could not find any downloads that satisfy the requirement PyUnitReport (from HttpRunner)
```

You could install `PyUnitReport` manully at first.

```bash
$ pip install git+https://github.com/debugtalk/PyUnitReport.git#egg=PyUnitReport
```

And then everything will be OK when you reinstall `HttpRunner`.

```bash
$ pip install git+https://github.com/debugtalk/HttpRunner.git#egg=HttpRunner
```
