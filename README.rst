HttpRunner
==========

.. image:: https://img.shields.io/github/license/HttpRunner/HttpRunner.svg
    :target: https://github.com/HttpRunner/HttpRunner/blob/master/LICENSE

.. image:: https://travis-ci.org/debugtalk/HttpRunner.svg?branch=master
    :target: https://travis-ci.org/HttpRunner/HttpRunner

.. image:: https://coveralls.io/repos/github/debugtalk/HttpRunner/badge.svg?branch=master
    :target: https://coveralls.io/github/debugtalk/HttpRunner?branch=master

.. image:: https://img.shields.io/pypi/v/HttpRunner.svg
    :target: https://pypi.python.org/pypi/HttpRunner

.. image:: https://img.shields.io/pypi/pyversions/HttpRunner.svg
    :target: https://pypi.python.org/pypi/HttpRunner


New name for ``ApiTestEngine``.

Design Philosophy
-----------------

Take full reuse of Python's existing powerful libraries: `Requests`_, `unittest`_ and `Locust`_. And achieve the goal of API automation test, production environment monitoring, and API performance test, with a concise and elegant manner.

Key Features
------------

- Inherit all powerful features of `Requests`_, just have fun to handle HTTP in human way.
- Define testcases in YAML or JSON format in concise and elegant manner.
- Supports ``function``/``variable``/``extract``/``validate`` mechanisms to create full test scenarios.
- With ``debugtalk.py`` plugin, module functions can be auto-discovered in recursive upward directories.
- Testcases can be run in diverse ways, with single testset, multiple testsets, or entire project folder.
- Test report is concise and clear, with detailed log records. See `PyUnitReport`_.
- With reuse of `Locust`_, you can run performance test without extra work.
- CLI command supported, perfect combination with `Jenkins`_.

Documentation
-------------

HttpRunner is rich documented.

- `User documentation`_ helps you to make the most use of HttpRunner
- `Development process blogs`_ will make you fully understand HttpRunner


.. _Requests: http://docs.python-requests.org/en/master/
.. _unittest: https://docs.python.org/3/library/unittest.html
.. _Locust: http://locust.io/
.. _PyUnitReport: https://github.com/HttpRunner/PyUnitReport
.. _Jenkins: https://jenkins.io/index.html
.. _User documentation: http://httprunner.readthedocs.io/
.. _Development process blogs: http://debugtalk.com/tags/ApiTestEngine/
