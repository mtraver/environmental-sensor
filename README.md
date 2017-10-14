# MCP9808 Temperature Logger

Log temperature from an MCP9808 sensor connected to a Raspberry Pi or BeagleBone Black.

Print timestamp and temperature to stdout:

    python log_temp.py

Log timestamp and three temperature values, taken two seconds apart, to a CSV file:

    python log_temp.py -n 3 -f temp_log.csv

Log timestamp and three temperature values, taken five seconds apart, to a Google Sheets spreadsheet:

    python log_temp.py -n 3 -d 5 -k keyfile.json -s 1BxiMVs0XRA5nFMdKvBdBZjgmUUqptlbs74OgvE2upms

Set up a cron job, use it in a daemon, the world's your oyster...as long as the world is temperature values read from the MCP9808.

## Prerequisites

This will install [pandas](http://pandas.pydata.org), as it's needed for plotting. It may take some time for pandas to build and install, so if you'd rather not install it, comment it out in ``requirements.txt``.

    pip install --user -r requirements.txt

Adafruit have a tutorial with information on wiring up the hardware: https://learn.adafruit.com/mcp9808-temperature-sensor-python-library/overview

__NOTE:__ You may need to enable I<sup>2</sup>C on your board. For Raspberry Pi, this can be done with ``raspi-config``. Go to "Advanced Options" and find the "I2C" option.

## Setting up Google Sheets logging

Google provide a guide to using the Sheets API in Python: https://developers.google.com/sheets/api/quickstart/python.

1. Follow "Step 1: Turn on the Google Sheets API" and use the wizard linked there to make a project and set up access credentials. What you want is a "service account", which is what the wizard will recommend if you say that you want access from a headless device/crontab/etc. You'll get an email address that looks something like this:

   ``<something>@<something>.iam.gserviceaccount.com``

   You'll also get a JSON file containing the key for that service account. Put it in a safe place.

2. Now make a Google Sheets spreadsheet and share it with the service account email address, giving edit permissions.

   Note the spreadsheet ID. For the example URL below, the ID is ``1BxiMVs0XRA5nFMdKvBdBZjgmUUqptlbs74OgvE2upms``:

   ``https://docs.google.com/spreadsheets/d/1BxiMVs0XRA5nFMdKvBdBZjgmUUqptlbs74OgvE2upms/edit``

The JSON key file and spreadsheet ID are the two things you'll need to log to the sheet.

## Full usage

    usage: log_temp.py [-h] [-s SHEET_ID] [-k KEYFILE] [-f LOG_FILE]
                       [-n NUM_SAMPLES] [-d SAMPLE_DELAY]

    Log temperature from MCP9808 sensor.

    Temperature can be logged to a file and/or a Google Sheets spreadsheet. If
    neither a file nor a spreadsheet is specified, data is logged to stdout.

    Temperature is recorded in degrees Celsius.

    optional arguments:
      -h, --help            show this help message and exit

    Logging:
      -s SHEET_ID, --sheet_id SHEET_ID
                            Google Sheets spreadsheet ID. If given, -k/--keyfile
                            is required. The sheet must be shared with the service
                            account email address associated with the key.
      -k KEYFILE, --keyfile KEYFILE
                            Path to Google API service account JSON key file. If
                            given, -s/--sheet_id is required.
      -f LOG_FILE, --log_file LOG_FILE
                            CSV file to which to log data

    Data sampling:
      -n NUM_SAMPLES, --num_samples NUM_SAMPLES
                            Number of samples to take. Defaults to 1.
      -d SAMPLE_DELAY, --sample_delay SAMPLE_DELAY
                            Number of seconds to sleep between samples. Defaults
                            to 2.
