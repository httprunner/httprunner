.. default-role:: code

Installation
============

``HttpRunner`` is available on `PyPI`_ and can be installed through pip or easy_install. ::

    $ pip install HttpRunner

or ::

    $ easy_install HttpRunner


If you want to keep up with the latest version, you can install with github repository url. ::

    $ pip install git+https://github.com/HttpRunner/HttpRunner.git#egg=HttpRunner


Upgrade
-------

If you have installed ``HttpRunner`` before and want to upgrade to the latest version, you can use the ``-U`` option.

This option works on each installation method described above. ::

    $ pip install -U HttpRunner
    $ easy_install -U HttpRunner
    $ pip install -U git+https://github.com/HttpRunner/HttpRunner.git#egg=HttpRunner


Check Installation
------------------

When HttpRunner is installed, a **httprunner** (**hrun** for short) command should be available in your shell (if you're not using
virtualenv—which you should—make sure your python script directory is on your path).

To see ``HttpRunner`` version: ::

    $ httprunner -V     # same as: hrun -V
    HttpRunner version: 0.8.1b
    PyUnitReport version: 0.1.3b

To see available options, run::

    $ httprunner -h     # same as: hrun -h
    usage: httprunner [-h] [-V] [--log-level LOG_LEVEL] [--report-name REPORT_NAME]
            [--failfast] [--startproject STARTPROJECT]
            [testset_paths [testset_paths ...]]

    HttpRunner.

    positional arguments:
    testset_paths         testset file path

    optional arguments:
    -h, --help            show this help message and exit
    -V, --version         show version
    --log-level LOG_LEVEL
                            Specify logging level, default is INFO.
    --report-name REPORT_NAME
                            Specify report name, default is generated time.
    --failfast            Stop the test run on the first error or failure.
    --startproject STARTPROJECT
                            Specify new project name.


Supported Python Versions
-------------------------

HttpRunner supports Python 2.7, 3.4, 3.5, and 3.6. And we strongly recommend you to use ``Python 3.6``.

``HttpRunner`` has been tested on ``macOS``, ``Linux`` and ``Windows`` platforms.


.. _PyPI: https://pypi.python.org/pypi
