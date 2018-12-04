# Python Client

This directory contains code that runs on the Raspberry Pi. It reads the
temperature and logs it to your platform of choice.

Note: This project only supports Python 3. The future is now. The future was in
[2008](https://www.python.org/download/releases/3.0/). Come with us into the
future.

Print timestamp and temperature to stdout:

    python3 log_temp.py stdout

Log timestamp and three temperature values, taken two seconds apart,
to a CSV file:

    python3 log_temp.py -n 3 csv temp_log.csv

Log via [Google Cloud IoT Core](https://cloud.google.com/iot-core/), storing
the data in Google Cloud Datastore or Google Cloud Bigtable (see the Google App
Engine app in the [web](web) directory):

    python3 log_temp.py iotcore -p my-gcp-project -r my-iot-core-registry \
      -k device_key.pem --device_id my-device

Log via [Google Cloud Pub/Sub](https://cloud.google.com/pubsub/), which like
IoT Core can store the data in Google Cloud Datastore or Google Cloud Bigtable
with the App Engine app in [web](web):

    python3 log_temp.py pubsub -p my-gcp-project -t my-pubsub-topic --device_id my-device

Log timestamp and three temperature values, taken five seconds apart, to a
Google Sheets spreadsheet:

    python3 log_temp.py -n 3 -i 5 sheets -k keyfile.json -s 1BxiMVs0XRA5nFMdKvBdBZjgmUUqptlbs74OgvE2upms

## Prerequisites

On the Raspberry Pi, install packages from
``requirements_logging.txt``:

    # This is required for the cryptography package
    sudo apt-get install build-essential libssl-dev libffi-dev python-dev python3-dev

    # I highly recommend using a virtualenv. This makes an isolated Python environment
    # so that reasoning about your dependencies is easier. One of the commands below
    # should work, depending on how virtualenv is installed.
    virtualenv -p python3 env3
    # OR
    python3 -m virtualenv env3

    # Enter the virtualenv (your prompt will change to signal you're inside)
    . env3/bin/activate

    # Upgrade the virtualenv's version of pip and setuptools
    pip install --upgrade pip setuptools

    # Install dependencies in the virtualenv
    pip install -r requirements_logging.txt

    # If you want to leave the virtualenv execute this
    deactivate

On your development machine, install packages from ``requirements_dev.txt``.
This will install [pandas](http://pandas.pydata.org), as it's needed for
plotting. Pandas is not required for logging and can take a long time to build
and install on a system like a Raspberry Pi so it's only included in
``requirements_dev.txt``.

    # virtualenv is useful on your dev machine too! See above if you want to use it.

    # Install dev dependencies
    pip install --user -r requirements_dev.txt

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
