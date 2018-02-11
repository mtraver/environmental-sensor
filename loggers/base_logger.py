class Logger(object):
  """Base class for temperature measurement loggers."""

  def log(self, timestamp, values):
    """Logs the given timestamp and temperature values.

    Args:
      timestamp: A datetime.datetime.
      values: A list of temperature values to log. The subclass can choose to
              log all of them, or if the backend doesn't support that, to
              reduce the list to a single value by some strategy such as
              taking the mean.
    """
    raise NotImplementedError()
