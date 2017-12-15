"""A dummy implementation of Adafruit's MCP9808 sensor class."""
import random


class MCP9808(object):
  """A dummy implementation of the class of the same name in
  https://github.com/adafruit/Adafruit_Python_MCP9808/blob/master/Adafruit_MCP9808/MCP9808.py
  """

  def __init__(self, address=None, i2c=None, **kwargs):
    # This keeps pylint from complaining about unused vars.
    del address, i2c, kwargs

    self._device = None

  def begin(self):
    pass

  # pylint: disable=no-self-use
  def readTempC(self):  # pylint: disable=invalid-name
    """Returns a random temperature in degrees Celsius."""
    # 15 to 19 Celsius is a fantastic temperature.
    return round(random.uniform(15, 19), 2)
  # pylint: enable=no-self-use
