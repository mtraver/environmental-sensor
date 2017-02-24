"""Utility functions releated to argument parsing."""
import argparse


def non_empty_string(s):
  if s is not None and not s:
    raise argparse.ArgumentTypeError('argument may not be empty')

  return s
