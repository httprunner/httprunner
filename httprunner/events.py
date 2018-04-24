# encoding: utf-8
from httprunner.exception import MyBaseError


class EventHook(object):
    """
    Simple event class used to provide hooks for different types of events in HttpRunner.

    Here's how to use the EventHook class::

        my_event = EventHook()
        def on_my_event(a, b, **kw):
            print "Event was fired with arguments: %s, %s" % (a, b)
        my_event += on_my_event
        my_event.fire(a="foo", b="bar")
    """

    def __init__(self):
        self._handlers = []

    def __iadd__(self, handler):
        self._handlers.append(handler)
        return self

    def __isub__(self, handler):
        if handler not in self._handlers:
            raise MyBaseError("handler not found: {}".format(handler))

        index = self._handlers.index(handler)
        self._handlers.pop(index)
        return self

    def fire(self, **kwargs):
        for handler in self._handlers:
            handler(**kwargs)
