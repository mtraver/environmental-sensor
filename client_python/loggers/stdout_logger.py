from . import base_logger


class StdoutLogger(base_logger.GCPLogger):

  def __init__(self, asciiproto, device_id):
    super(StdoutLogger, self).__init__('', device_id)
    self._asciiproto = asciiproto

  def log(self, timestamp, values):
    if self._asciiproto:
      print(self._get_proto(timestamp, values))
    else:
      print(','.join([timestamp.isoformat()] + [str(x) for x in values]))
