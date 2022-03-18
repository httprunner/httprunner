import logging
from typing import List

import funppy


def sum(*args):
    result = 0
    for arg in args:
        result += arg
    return result

def sum_ints(*args: List[int]) -> int:
    result = 0
    for arg in args:
        result += arg
    return result

def sum_two_int(a: int, b: int) -> int:
    return a + b

def sum_two_string(a: str, b: str) -> str:
    return a + b

def sum_strings(*args: List[str]) -> str:
    result = ""
    for arg in args:
        result += arg
    return result

def concatenate(*args: List[str]) -> str:
    result = ""
    for arg in args:
        result += str(arg)
    return result

def setup_hook_example(name):
    logging.warning("setup_hook_example")
    return f"setup_hook_example: {name}"

def teardown_hook_example(name):
    logging.warning("teardown_hook_example")
    return f"teardown_hook_example: {name}"


if __name__ == '__main__':
    funppy.register("sum", sum)
    funppy.register("sum_ints", sum_ints)
    funppy.register("concatenate", concatenate)
    funppy.register("sum_two_int", sum_two_int)
    funppy.register("sum_two_string", sum_two_string)
    funppy.register("sum_strings", sum_strings)
    funppy.register("setup_hook_example", setup_hook_example)
    funppy.register("teardown_hook_example", teardown_hook_example)
    funppy.serve()
