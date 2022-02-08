import unittest

import main


class TestMain(unittest.TestCase):
    def test_print_hi(self):
        self.assertEqual(main.print_hi("Tux"), "Hi, Tux")
