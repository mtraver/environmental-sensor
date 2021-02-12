# MCP9808 Temperature Logger

[![GoDoc](https://godoc.org/github.com/mtraver/environmental-sensor?status.svg)](https://godoc.org/github.com/mtraver/environmental-sensor)
[![Go Report Card](https://goreportcard.com/badge/github.com/mtraver/environmental-sensor)](https://goreportcard.com/report/github.com/mtraver/environmental-sensor)

Log temperature from an [MCP9808 sensor](https://www.adafruit.com/product/1782)
connected to a Raspberry Pi.

Send temperature to [Google Cloud IoT Core](https://cloud.google.com/iot-core/),
which can then be saved and plotted using the Google App Engine app in the
[web](web) directory:

    make

    # Example device.json:
    # {
    #   "project_id": "my-gcp-project",
    #   "registry_id": "my-iot-core-registry",
    #   "device_id": "my-device",
    #   "priv_key_path": "my-device.pem",
    #   "region": "us-central1"
    # }
    ./out/iotcorelogger -device device.json -cacerts roots.pem

Set up a cron job, use it in a daemon, the world's your oyster...as long as the
world is temperature values read from the MCP9808.

## Prerequisites

  - **Wire up the hardware.** Adafruit have a nice tutorial:
https://learn.adafruit.com/mcp9808-temperature-sensor-python-library/overview
  - **Enable I<sup>2</sup>C on your board.** For Raspberry Pi,
this can be done with ``raspi-config``. You'll find the "I2C" option under
either "Advanced Options" or "Interfacing Options".

## Setting up Google Cloud IoT Core logging

### Google Cloud setup

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

### Client program

The program in [cmd/iotcorelogger](cmd/iotcorelogger) runs on the Raspberry Pi
and sends data to Google Cloud IoT Core. [cmd/README](cmd/README)
has information on building and configuring `iotcorelogger`.

TODO(mtraver) add info on systemd config
