"""Utility functions related to argument parsing."""
import argparse
import re


def non_empty_string(s):
  if s is not None and not s:
    raise argparse.ArgumentTypeError('argument may not be empty')

  return s


def date_string(s):
  if not re.match(r'^[0-9]{4}-(0[1-9]|1[0-2])-(0[1-9]|[12][0-9]|3[01])$', s):
    raise argparse.ArgumentTypeError('Date must be of format YYYY-MM-DD')

  return s
