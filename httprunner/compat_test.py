import os
import unittest

from httprunner import compat


class TestCompat(unittest.TestCase):

    def test_convert_jmespath(self):

        self.assertEqual(
            compat.convert_jmespath("content.abc"),
            "body.abc"
        )
        self.assertEqual(
            compat.convert_jmespath("json.abc"),
            "body.abc"
        )
        self.assertEqual(
            compat.convert_jmespath("headers.Content-Type"),
            'headers."Content-Type"'
        )
