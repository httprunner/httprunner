
def setup_hook_add_kwargs(method, url, kwargs):
    kwargs["key"] = "value"

def setup_hook_remove_kwargs(method, url, kwargs):
    kwargs.pop("key")
