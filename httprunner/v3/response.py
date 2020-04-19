import requests


class ResponseObject(object):

    def __init__(self, resp_obj: requests.Response):
        """ initialize with a requests.Response object

        Args:
            resp_obj (instance): requests.Response instance

        """
        self.obj = resp_obj
