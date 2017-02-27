"""Plot temperature log from Google Sheet or CSV file.

In the case of multiple values per row the mean is plotted.
"""
import argparse
import dateutil

import matplotlib.dates as mdates
import matplotlib.pyplot as plt
import pandas

import google_sheets_logger
import util

DATE_COL_HEADER = 'Date'

# Data is logged in UTC and will be converted to local timezone for display
TZ_UTC = dateutil.tz.tzutc()
TZ_LOCAL = dateutil.tz.tzlocal()


class TemperatureUnit(object):
  CELSIUS = 'C'
  FAHRENHEIT = 'F'


CELSIUS_CHOICES = ['c', 'celsius']
FAHRENHEIT_CHOICES = ['f', 'fahrenheit']

MEAN_COL_HEADER_FORMAT = 'Mean %s'
C_MEAN_COL_HEADER = MEAN_COL_HEADER_FORMAT % TemperatureUnit.CELSIUS
F_MEAN_COL_HEADER = MEAN_COL_HEADER_FORMAT % TemperatureUnit.FAHRENHEIT


def c_to_f(c):
  return c * 9.0 / 5.0 + 32.0


if __name__ == '__main__':
  parser = argparse.ArgumentParser(description=__doc__)

  data_group = parser.add_argument_group('Data source')
  data_group.add_argument('-s', '--sheet_id',
                          type=util.argparse_utils.non_empty_string,
                          help='Google Sheets spreadsheet ID. If given, '
                               '-k/--keyfile is required. The sheet must '
                               'be shared with the service account email '
                               'address associated with the key.')
  data_group.add_argument('-k', '--keyfile',
                          type=util.argparse_utils.non_empty_string,
                          help='Path to Google API service account JSON key '
                               'file. If given, -s/--sheet_id is required.')
  data_group.add_argument('-f', '--log_file',
                          type=util.argparse_utils.non_empty_string,
                          help='CSV file to read')

  date_group = parser.add_argument_group('Date filtering')
  date_group.add_argument('--start', type=util.argparse_utils.date_string,
                          metavar='YYYY-MM-DD',
                          help='First day to plot. Should be of '
                               'format YYYY-MM-DD.')
  date_group.add_argument('--end', type=util.argparse_utils.date_string,
                          metavar='YYYY-MM-DD',
                          help='Last day to plot. Should be of '
                               'format YYYY-MM-DD.')

  parser.add_argument('-t', '--temp_unit', type=str.lower,
                      choices=CELSIUS_CHOICES + FAHRENHEIT_CHOICES,
                      default=CELSIUS_CHOICES[0],
                      help='Display temperature as this unit.')

  args = parser.parse_args()

  # -f/--log_file and the pair of (-k/--keyfile, -s/--sheet_id)
  # are mutually exclusive
  if args.log_file is not None and (args.keyfile or args.sheet_id):
    parser.error('You may specify a log file OR a Google Sheet, not both')

  # -k/--keyfile and -s/--sheet_id are mutually inclusive
  if args.keyfile is not None and args.sheet_id is None:
    parser.error('-k/--keyfile requires -s/--sheet_id')
  if args.keyfile is None and args.sheet_id is not None:
    parser.error('-s/--sheet_id requires -k/--keyfile')

  if args.keyfile is None and args.log_file is None:
    parser.error('A data source must be specified')

  if args.temp_unit in CELSIUS_CHOICES:
    args.temp_unit = TemperatureUnit.CELSIUS
  else:
    args.temp_unit = TemperatureUnit.FAHRENHEIT

  # Read data from file or Google Sheet
  data = None
  if args.keyfile is not None:
    data = google_sheets_logger.read_sheet(args.keyfile, args.sheet_id,
                                           parse_dates=[DATE_COL_HEADER],
                                           index_col=DATE_COL_HEADER)

    # Temp measurements may be strings in the spreadsheet,
    # so convert just to be sure they're recorded as floats
    data = data.astype(float)
  else:
    data = pandas.read_csv(args.log_file, index_col=0, parse_dates=True,
                           infer_datetime_format=True)

  data.index = data.index.tz_localize(TZ_UTC)
  data.index = data.index.tz_convert(TZ_LOCAL)

  if args.start is not None and args.end is None:
    data = data.ix[args.start:]
  elif args.start is None and args.end is not None:
    data = data.ix[:args.end]
  elif args.start is not None and args.end is not None:
    data = data.ix[args.start:args.end]

  data[C_MEAN_COL_HEADER] = data.mean(axis=1, numeric_only=True)
  if args.temp_unit == TemperatureUnit.FAHRENHEIT:
    data[F_MEAN_COL_HEADER] = data[C_MEAN_COL_HEADER].apply(c_to_f)

  fig, ax = plt.subplots(figsize=(20, 10))
  ax.plot(data.index, data[MEAN_COL_HEADER_FORMAT % args.temp_unit])

  # Major ticks are dates, minor ticks are every three hours (skipping 00:00
  # because that's the major tick time and the date is displayed there instead
  # of the hour). Matplotlib claims to be timezone-aware but without specifying
  # the timezone for all locators and formatters it's actually not.
  ax.xaxis.set_major_locator(mdates.DayLocator(tz=TZ_LOCAL))
  ax.xaxis.set_major_formatter(mdates.DateFormatter('%Y-%m-%d', tz=TZ_LOCAL))
  ax.xaxis.set_minor_locator(mdates.HourLocator(byhour=range(3, 24, 3),
                                                tz=TZ_LOCAL))
  ax.xaxis.set_minor_formatter(mdates.DateFormatter('%H:%M', tz=TZ_LOCAL))

  for label in ax.xaxis.get_majorticklabels() + ax.xaxis.get_minorticklabels():
    label.set_rotation(45)
    label.set_horizontalalignment('right')

  # Vertical grid lines for each day
  ax.xaxis.grid(True, which='major')

  ax.set_xlabel('Time')
  ax.set_ylabel('Degrees %s' % args.temp_unit)

  # Make space below the x-axis label
  fig.subplots_adjust(bottom=0.15)

  plt.show()
