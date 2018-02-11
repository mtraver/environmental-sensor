"""Log temperature in degrees Celsius from MCP9808 sensor.

Temperature can be logged via Google Cloud IoT Core, Google Cloud Pub/Sub,
to a Google Sheets spreadsheet, a CSV file, or stdout.
"""
import argparse
import csv
from datetime import datetime
import os
import re
import time

import cryptography.x509

import loggers
import loggers.sheets
import util

# pylint: disable=wrong-import-position
DEBUG = False
if DEBUG:
  import util.dummy_mcp9808 as MCP9808
else:
  import Adafruit_MCP9808.MCP9808 as MCP9808  # pylint: disable=import-error
# pylint: enable=wrong-import-position

DEVICE_ID_REGEX = re.compile(r'^[a-zA-Z0-9_-]+$')

DEFAULT_NUM_SAMPLES = 1
DEFAULT_SAMPLE_INTERVAL_SECS = 2

DATE_COL_HEADER = 'Date'

IOTCORE_COMMAND = 'iotcore'
PUBSUB_COMMAND = 'pubsub'
SHEETS_COMMAND = 'sheets'
CSV_COMMAND = 'csv'
STDOUT_COMMAND = 'stdout'

PROJECT_ID_HELP = 'Google Cloud Platform project name'


def log_to_csv(filename, timestamp, data):
  # Write headers if the file doesn't exist or if it's empty
  write_header = not os.path.isfile(filename) or os.stat(filename).st_size == 0

  with open(filename, 'a') as f:
    csv_writer = csv.writer(f)

    if write_header:
      headers = [DATE_COL_HEADER] + ['Temp%d' % (i + 1)
                                     for i in xrange(len(data))]
      csv_writer.writerow(headers)

    csv_writer.writerow([timestamp.isoformat()] + data)


def valid_device_id(device_id):
  """Validates a device ID, raising an argparse.ArgumentTypeError if invalid."""
  match = DEVICE_ID_REGEX.match(device_id)
  if not match:
    raise argparse.ArgumentTypeError('Invalid device ID. Must contain only '
                                     'letters, numbers, dash, and underscore.')

  return device_id


def parse_args():
  parser = argparse.ArgumentParser(
      description=__doc__, formatter_class=argparse.RawDescriptionHelpFormatter)

  # These options are common to all logging modes
  sampling_group = parser.add_argument_group('Data sampling')
  sampling_group.add_argument(
      '-n', '--num_samples', type=int,
      default=DEFAULT_NUM_SAMPLES,
      help='Number of samples to take. Defaults to %d.' % DEFAULT_NUM_SAMPLES)
  sampling_group.add_argument(
      '-i', '--sample_interval', type=int,
      default=DEFAULT_SAMPLE_INTERVAL_SECS,
      help='Number of seconds to wait between samples. Defaults '
           'to {:d}.'.format(DEFAULT_SAMPLE_INTERVAL_SECS))

  # Specifying the dest kwarg puts the name of the subparser that was invoked
  # into the argparse namespace.
  subparsers = parser.add_subparsers(
      help=('Run one of these commands with the -h/--help flag '
            'to see its usage.'), dest='command')

  # Google Cloud IoT Core
  iotcore_parser = subparsers.add_parser(
      IOTCORE_COMMAND, help='Log via Google Cloud IoT Core')

  auth_group = iotcore_parser.add_argument_group('Authentication')
  auth_group.add_argument(
      '-p', '--project_id', required=True,
      default=os.environ.get('GOOGLE_CLOUD_PROJECT'),
      help=PROJECT_ID_HELP)
  auth_group.add_argument(
      '-r', '--registry_id', required=True,
      help='Google Cloud IoT Core registry ID')
  auth_group.add_argument(
      '-k', '--private_key_file', required=True,
      help=("Path to the device's private key file. It must have been added "
            "to the IoT Core registry."))

  device_id_group = iotcore_parser.add_argument_group(
      'Device ID',
      description=('A device ID is stored with each record saved in Google '
                   'Cloud Bigtable or Google Cloud Datastore. Set it using '
                   'one of these options.'))
  device_id_mutex_group = device_id_group.add_mutually_exclusive_group()
  device_id_mutex_group.add_argument('--device_id', type=valid_device_id)
  device_id_mutex_group.add_argument(
      '--device_id_from_cert', action='store_true', default=False,
      help='Get device ID from CN (common name) field of device cert')

  gcp_group = iotcore_parser.add_argument_group('Google Cloud Platform')
  gcp_group.add_argument(
      '--cloud_region', default=loggers.CloudIotMqttLogger.DEFAULT_CLOUD_REGION,
      help=('Google Cloud Platform region. Defaults to '
            '{}.').format(loggers.CloudIotMqttLogger.DEFAULT_CLOUD_REGION))

  # Paho always reports '1: Out of memory.' in the disconnect callback,
  # so disable MQTT for now and just use HTTP.
  # See https://github.com/GoogleCloudPlatform/python-docs-samples/issues/1357
  # mqtt_group = iotcore_parser.add_argument_group('MQTT Options')
  # mqtt_group.add_argument(
  #     '--mqtt_bridge_hostname',
  #     default=loggers.CloudIotMqttLogger.DEFAULT_MQTT_BRIDGE,
  #     help=('MQTT bridge hostname. Defaults to '
  #           '{}.').format(loggers.CloudIotMqttLogger.DEFAULT_MQTT_BRIDGE))
  # mqtt_group.add_argument(
  #     '--mqtt_bridge_port', type=int,
  #     choices=loggers.CloudIotMqttLogger.MQTT_PORTS,
  #     default=loggers.CloudIotMqttLogger.DEFAULT_MQTT_PORT,
  #     help=('MQTT bridge port. Defaults to '
  #           '{:d}.').format(loggers.CloudIotMqttLogger.DEFAULT_MQTT_PORT))

  # Google Cloud Pub/Sub
  pubsub_parser = subparsers.add_parser(
      PUBSUB_COMMAND, help='Publish to Google Cloud Pub/Sub')
  pubsub_parser.add_argument(
      '-p', '--project_id', required=True,
      default=os.environ.get('GOOGLE_CLOUD_PROJECT'),
      help=PROJECT_ID_HELP)
  pubsub_parser.add_argument('-t', '--topic', required=True,
                             help='Cloud Pub/Sub topic')
  pubsub_parser.add_argument(
      '--device_id',
      type=util.argparse_utils.non_empty_string, required=True)

  # Google Sheets
  sheets_parser = subparsers.add_parser(
      SHEETS_COMMAND, help='Log to a Google Sheet')
  sheets_parser.add_argument(
      '-k', '--keyfile', required=True,
      type=util.argparse_utils.non_empty_string,
      help='Path to Google API service account JSON key file')
  sheets_parser.add_argument(
      '-s', '--sheet_id', required=True,
      type=util.argparse_utils.non_empty_string,
      help=('Google Sheets spreadsheet ID. The sheet must be shared with the '
            'service account email address associated with the key.'))

  # CSV file
  csv_parser = subparsers.add_parser(CSV_COMMAND, help='Log to a CSV file')
  csv_parser.add_argument('log_file',
                          type=util.argparse_utils.non_empty_string,
                          help='CSV file to which to log data.')

  # Standard out
  csv_parser = subparsers.add_parser(STDOUT_COMMAND, help='Log to standard out')

  return parser.parse_args()


def get_device_id(args):
  """Returns the device ID based on the command line args."""
  if args.device_id is not None:
    return args.device_id
  elif args.device_id_from_cert:
    # Get the cert filename from the key filename
    base, _ = os.path.splitext(args.private_key_file)
    cert_path = base + '.x509'

    if not os.path.isfile(cert_path):
      raise Exception('Certificate file does not exist: {}'.format(cert_path))

    with open(cert_path, 'r') as f:
      cert = cryptography.x509.load_pem_x509_certificate(
          f.read(), cryptography.hazmat.backends.default_backend())

    # Get the common name from the cert
    common_name = cert.subject.get_attributes_for_oid(
        cryptography.x509.oid.NameOID.COMMON_NAME)[0].value

    match = DEVICE_ID_REGEX.match(common_name)
    if not match:
      raise Exception('Invalid device ID from cert common name: "{}". Must '
                      'contain only letters, numbers, dash, and '
                      'underscore.'.format(common_name))

    return common_name
  else:
    # This will only happen in case of programming error
    raise Exception('Unknown device ID choice!')


def main():
  args = parse_args()

  sensor = MCP9808.MCP9808()
  sensor.begin()

  timestamp = datetime.utcnow()

  # Construct list of temperature measurements
  data = []
  for i in xrange(args.num_samples):
    data.append(sensor.readTempC())

    # No need to sleep after last measurement is recorded
    if not DEBUG and i < args.num_samples - 1:
      time.sleep(args.sample_interval)

  if args.command == IOTCORE_COMMAND:
    # Paho always reports '1: Out of memory.' in the disconnect callback,
    # so disable MQTT for now and just use HTTP.
    # See https://github.com/GoogleCloudPlatform/python-docs-samples/issues/1357
    # iotcore_logger = loggers.CloudIotMqttLogger(
    #     args.project_id, args.registry_id, args.private_key_file,
    #     get_device_id(args), cloud_region=args.cloud_region,
    #     bridge_hostname=args.mqtt_bridge_hostname,
    #     bridge_port=args.mqtt_bridge_port)

    iotcore_logger = loggers.CloudIotHttpLogger(
        args.project_id, args.registry_id, args.private_key_file,
        get_device_id(args), cloud_region=args.cloud_region)

    iotcore_logger.log(timestamp, data)
  elif args.command == PUBSUB_COMMAND:
    pubsub_logger = loggers.PubSubLogger(
        args.project_id, args.topic, args.device_id)
    pubsub_logger.log(timestamp, data)
  elif args.command == SHEETS_COMMAND:
    sheets_writer = loggers.sheets.Writer(args.keyfile, args.sheet_id)
    sheets_writer.append(timestamp, data)
  elif args.command == CSV_COMMAND:
    log_to_csv(args.log_file, timestamp, data)
  elif args.command == STDOUT_COMMAND:
    print ','.join([timestamp.isoformat()] + [str(x) for x in data])
  else:
    raise Exception('Unknown command: "{}"'.format(args.command))


if __name__ == '__main__':
  main()
