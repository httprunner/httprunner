def teardown_hook_set_encoding(response, encoding):
    """
    Set encoding of response.
    """
    response.resp_obj.encoding = encoding
    return response
