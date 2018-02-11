# MCP9808 Temperature Logger

Log temperature from an MCP9808 sensor connected to a Raspberry Pi or
BeagleBone Black.

Print timestamp and temperature to stdout:

    python log_temp.py stdout

Log timestamp and three temperature values, taken two seconds apart,
to a CSV file:

    python log_temp.py -n 3 csv temp_log.csv

Log via [Google Cloud IoT Core](https://cloud.google.com/iot-core/), storing
the data in Google Cloud Datastore or Google Cloud Bigtable (see the App Engine
app in the [receiver](receiver) directory):

    python log_temp.py iotcore -p my-gcp-project -r my-iot-core-registry -k device_key.pem --device_id my-device

Log via [Google Cloud Pub/Sub](https://cloud.google.com/pubsub/), which like
IoT Core can store the data in Google Cloud Datastore or Google Cloud Bigtable
with the [receiver](receiver) App Engine app:

    python log_temp.py pubsub -p my-gcp-project -t my-pubsub-topic --device_id my-device

Log timestamp and three temperature values, taken five seconds apart, to a
Google Sheets spreadsheet:

    python log_temp.py -n 3 -i 5 sheets -k keyfile.json -s 1BxiMVs0XRA5nFMdKvBdBZjgmUUqptlbs74OgvE2upms

Set up a cron job, use it in a daemon, the world's your oyster...as long as the
world is temperature values read from the MCP9808.

## Prerequisites

On the Raspberry Pi / BeagleBone, install packages from
``requirements_logging.txt``:

    # This is required for the cryptography package
    sudo apt-get install build-essential libssl-dev libffi-dev python-dev

    pip install --user -r requirements_logging.txt

On your development machine, install packages from ``requirements_dev.txt``.
This will install [pandas](http://pandas.pydata.org), as it's needed for
plotting. Pandas is not required for logging and can take a long time to build
and install on a system like a Raspberry Pi so it's only included in
``requirements_dev.txt``.

    pip install --user -r requirements_dev.txt

Adafruit have a tutorial with information on wiring up the hardware:
https://learn.adafruit.com/mcp9808-temperature-sensor-python-library/overview

__NOTE:__ You may need to enable I<sup>2</sup>C on your board. For Raspberry Pi,
this can be done with ``raspi-config``. You'll find the "I2C" option under
either "Advanced Options" or "Interfacing Options".

## Setting up Google Cloud IoT Core logging

TODO

## Setting up Google Cloud Pub/Sub logging

TODO

## Setting up Google Sheets logging

Google provide a guide to using the Sheets API in Python:
https://developers.google.com/sheets/api/quickstart/python.

1. Follow "Step 1: Turn on the Google Sheets API" and use the wizard linked
   there to make a project and set up access credentials. What you want is a
   "service account", which is what the wizard will recommend if you say that
   you want access from a headless device/crontab/etc. You'll get an email
   address that looks something like this:

   ``<something>@<something>.iam.gserviceaccount.com``

   You'll also get a JSON file containing the key for that service account.
   Put it in a safe place.

2. Now make a Google Sheets spreadsheet and share it with the service account
   email address, giving edit permissions.

   Note the spreadsheet ID. For the example URL below, the ID is
   ``1BxiMVs0XRA5nFMdKvBdBZjgmUUqptlbs74OgvE2upms``:

   ``https://docs.google.com/spreadsheets/d/1BxiMVs0XRA5nFMdKvBdBZjgmUUqptlbs74OgvE2upms/edit``

The JSON key file and spreadsheet ID are the two things you'll need to log to
the sheet.

## Full usage

    usage: log_temp.py [-h] [-n NUM_SAMPLES] [-i SAMPLE_INTERVAL]
                       {iotcore,pubsub,sheets,csv,stdout} ...

    Log temperature in degrees Celsius from MCP9808 sensor.

    Temperature can be logged via Google Cloud IoT Core, Google Cloud Pub/Sub,
    to a Google Sheets spreadsheet, a CSV file, or stdout.

    positional arguments:
      {iotcore,pubsub,sheets,csv,stdout}
                            Run one of these commands with the -h/--help flag to
                            see its usage.
        iotcore             Log via Google Cloud IoT Core
        pubsub              Publish to Google Cloud Pub/Sub
        sheets              Log to a Google Sheet
        csv                 Log to a CSV file
        stdout              Log to standard out

    optional arguments:
      -h, --help            show this help message and exit

    Data sampling:
      -n NUM_SAMPLES, --num_samples NUM_SAMPLES
                            Number of samples to take. Defaults to 1.
      -i SAMPLE_INTERVAL, --sample_interval SAMPLE_INTERVAL
                            Number of seconds to wait between samples. Defaults to
                            2.
