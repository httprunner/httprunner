# Introduction

## Design Philosophy

Take full reuse of Python's existing powerful libraries: [`Requests`][requests], [`unittest`][unittest] and [`Locust`][Locust]. And achieve the goal of API automation test, production environment monitoring, and API performance test, with a concise and elegant manner.

## Key Features

- Inherit all powerful features of [`Requests`][requests], just have fun to handle HTTP in human way.
- Define testcases in YAML or JSON format in concise and elegant manner.
- Supports `function`/`variable`/`extract`/`validate` mechanisms to create full test scenarios.
- With `debugtalk.py` plugin, module functions can be auto-discovered in recursive upward directories.
- Testcases can be run in diverse ways, with single testset, multiple testsets, or entire project folder.
- Test report is concise and clear, with detailed log records. See [`PyUnitReport`][PyUnitReport].
- With reuse of [`Locust`][Locust], you can run performance test without extra work.
- CLI command supported, perfect combination with [Jenkins][Jenkins].

## Learn more

You can read this [blog][HttpRunner-blog] to learn more about the background and initial thoughts of `HttpRunner`.


[requests]: http://docs.python-requests.org/en/master/
[unittest]: https://docs.python.org/3/library/unittest.html
[Locust]: http://locust.io/
[PyUnitReport]: https://github.com/HttpRunner/PyUnitReport
[Jenkins]: https://jenkins.io/index.html
[HttpRunner-blog]: http://debugtalk.com/post/ApiTestEngine-api-test-best-practice/
