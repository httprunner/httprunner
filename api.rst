.. _api:

Developer Interface
===================

.. module:: httprunner

This part of the documentation covers all the interfaces of HttpRunner.


Main Interface
--------------

All of HttpRunner' functionality can be accessed by these 7 methods.
They all return an instance of the :class:`Response <Response>` object.

.. autofunction:: request

Exceptions
----------

.. autoexception:: httprunner.exceptions.ValidationFailure
.. autoexception:: requests.ConnectionError
.. autoexception:: requests.HTTPError
.. autoexception:: requests.URLRequired
.. autoexception:: requests.TooManyRedirects
.. autoexception:: requests.ConnectTimeout
.. autoexception:: requests.ReadTimeout
.. autoexception:: requests.Timeout

