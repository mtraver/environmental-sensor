import csv
import os

from . import base_logger


class CsvLogger(base_logger.Logger):

  DATE_COL_HEADER = 'Date'

  def __init__(self, filename):
    self._filename = filename

  def log(self, timestamp, values):
    # Write headers if the file doesn't exist or if it's empty
    write_header = (not os.path.isfile(self._filename)
                    or os.stat(self._filename).st_size == 0)

    with open(self._filename, 'a') as f:
      csv_writer = csv.writer(f)

      if write_header:
        headers = [self.DATE_COL_HEADER] + ['Temp%d' % (i + 1)
                                            for i in range(len(values))]
        csv_writer.writerow(headers)

      csv_writer.writerow([timestamp.isoformat()] + values)
