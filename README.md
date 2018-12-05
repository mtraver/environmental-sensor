# MCP9808 Temperature Logger

[![Go Report Card](https://goreportcard.com/badge/github.com/mtraver/environmental-sensor)](https://goreportcard.com/report/github.com/mtraver/environmental-sensor)

Log temperature from an [MCP9808 sensor](https://www.adafruit.com/product/1782)
connected to a Raspberry Pi.

Send temperature to [Google Cloud IoT Core](https://cloud.google.com/iot-core/),
which can then be saved and plotted using the Google App Engine app in the
[web](web) directory:

    ./iotcorelogger -project my-gcp-project -registry my-iot-core-registry \
      -key device_key.pem -cacerts roots.pem

Set up a cron job, use it in a daemon, the world's your oyster...as long as the
world is temperature values read from the MCP9808.

## Choose a Client

This project includes clients — code that runs on the Raspberry Pi to read the
temperature and log it — written in Go and Python.

Use the [Go client](cmd) if you're sending temperature data to Google
Cloud IoT Core (it only supports Cloud IoT Core at the moment). It's easier to
work with than the Python client because you get a statically-linked binary that
just works on the Raspberry Pi. You don't have to clone this repository on the
Raspberry Pi or install dependencies or set up a virtualenv. Just `make` and run.

Use the [Python client](client_python) if you want to log directly to Google
Cloud Pub/Sub, or to Google Sheets. The Python client supports Cloud IoT Core
as well. Note: This project only supports Python 3. The future is now. The
future was in [2008](https://www.python.org/download/releases/3.0/). Come with
us into the future.

## Prerequisites

Each client's README has information about its prerequisites.

Regardless of your choice of client, you'll need to:

  - **Wire up the hardware.** Adafruit have a nice tutorial:
https://learn.adafruit.com/mcp9808-temperature-sensor-python-library/overview
  - **Enable I<sup>2</sup>C on your board.** For Raspberry Pi,
this can be done with ``raspi-config``. You'll find the "I2C" option under
either "Advanced Options" or "Interfacing Options".

## Setting up Google Cloud IoT Core logging

The scripts at https://github.com/mtraver/provisioning are useful for creating
the CA key and cert and device-specific keys and certs described below.

- Create an IoT Core registry.
  The [IoT Core quickstart](https://cloud.google.com/iot/docs/quickstart)
  provides more info. The registry includes:
  - A Pub/Sub topic for telemetry (you'll need to create the topic if it
    doesn't already exist)
  - A Pub/Sub topic for state (you'll need to create the topic if it
    doesn't already exist)
  - A CA cert for verifying device certs. This can be self-signed.
- Add devices to the registry. This requires a device-specific cert that chains
  to the CA cert. The key and cert can be made with the scripts in the repo
  linked above. Heed the information there about key handling and about the
  device ID (the device ID you use when making the cert must be the same as the
  one you set when adding the device to the registry).
- Create a subscription to the registry's telemetry topic. Configure it to
  push to the ``/_ah/push-handlers/telemetry`` endpoint of the web app.
  This is how IoT Core is tied to the web app.

The end-to-end flow is like this:
1. A device sends a payload (in this case a protobuf; see
   [measurement.proto](measurement.proto)) to IoT Core.
2. IoT Core publishes the payload as a Pub/Sub message to the registry's
   telemetry Pub/Sub topic.
3. Pub/Sub pushes the message to the web app's endpoint, as configured in
   the subscription to the topic.
4. The web app receives the request, decodes the payload, and writes
   it to the database.

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
