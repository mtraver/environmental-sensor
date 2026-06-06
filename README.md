# MCP9808 Temperature Logger

[![GoDoc](https://godoc.org/github.com/mtraver/environmental-sensor?status.svg)](https://godoc.org/github.com/mtraver/environmental-sensor)
[![Go Report Card](https://goreportcard.com/badge/github.com/mtraver/environmental-sensor)](https://goreportcard.com/report/github.com/mtraver/environmental-sensor)

Log temperature from an [MCP9808 sensor](https://www.adafruit.com/product/1782)
connected to a Raspberry Pi.

Send temperature to [AWS IoT Core](https://aws.amazon.com/iot-core/) (RIP Google
Cloud IoT Core, which I used before it was decommissioned 😭), which can then be
saved and plotted using the web app in the [web](web) directory. Running `make`
will build all binaries locally. Running `make web-image` will build a Docker
image that can be deployed to Cloud Run to serve the web app.

Run the `iotcorelogger` binary on the Raspberry Pi:

```sh
make

# Example device.json:
# {
#   "endpoint": "endpoint-name",
#   "device_id": "my-device",
#   "cert_path": "my-device.x509",
#   "priv_key_path": "my-device.pem"
# }
./out/iotcorelogger -aws-device device.json
```

The device file specifies how to connect to AWS IoT Core's MQTT broker.

## Prerequisites

  - **Wire up the hardware.** Adafruit have a nice tutorial:
https://learn.adafruit.com/mcp9808-temperature-sensor-python-library/overview
  - **Enable I<sup>2</sup>C on your board.** For Raspberry Pi,
this can be done with `raspi-config`. You'll find the "I2C" option under
either "Advanced Options" or "Interfacing Options".

## `iotcorelogger` sensor and job configuration

The `iotcorelogger` program is told which sensors to use and the frequency at
which to take measurements via a JSON job spec. A job has:

- A cronspec
- An operation, which must be one of `"SETUP"`, `"SENSE"`, or `"SHUTDOWN"`
- A list of sensors

Example of a simple config that gets a measurement from an MCP9808 temperature
sensor every 2 minutes:
```json
{
  "jobs": [
    {
      "cronspec": "0 */2 * * * *",
      "operation": "SENSE",
      "sensors": ["mcp9808"]
    }
  ]
}
```

Example of a more complex config that gets particulate matter measurements from an
SDS011 sensor every 2 minutes, but that runs setup and shutdown jobs before taking
measurements.
```json
{
  "jobs": [
    {
      "cronspec": "35 1-59/2 * * * *",
      "operation": "SETUP",
      "sensors": [
        "sds011"
      ]
    },
    {
      "cronspec": "0 0-59/2 * * * *",
      "operation": "SENSE",
      "sensors": [
        "sds011"
      ]
    },
    {
      "cronspec": "8 0-59/2 * * * *",
      "operation": "SHUTDOWN",
      "sensors": [
        "sds011"
      ]
    }
  ]
}
```

The device receives this config from an AWS IoT Core Device Shadow. See Device Shadow service
documentation [here](https://docs.aws.amazon.com/iot/latest/developerguide/iot-device-shadows.html).

When a device connects to the MQTT broker it will either create a shadow if one doesn't exist,
or fetch the current desired config from the shadow. Set the `desired` config in the device's
shadow configuration to push it to the device; the device will receive and apply the new config
any time it is changed.

## Setting up Google Cloud IoT Core logging

TODO(mtraver) Re-write for AWS IoT Core

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
  push to the `/push-handlers/telemetry` endpoint of the web app.
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

## Running and deploying the web app

### Environment variables

The web app expects the following environment variables to be set. Define them
in a file called `.env` and the `make run-web` command below will pick them up.

In production, you'll need to make sure that they are available to the container
via env var config, secrets config, or whatever other method you like.

TODO(mtraver) add descriptions of the env vars

- `AWS_REGION`
- `AWS_ROLE_ARN`
- `IGNORED_DEVICES`
- `INFLUXDB_BUCKET`
- `INFLUXDB_ORG`
- `INFLUXDB_SERVER`
- `INFLUXDB_TOKEN`
- `PUBSUB_AUDIENCE`
- `PUBSUB_VERIFICATION_TOKEN`

For local development you'll need to set `GOOGLE_CLOUD_PROJECT` to your GCP
project ID. In production on Cloud Run it's fetched automatically.

For local development you'll also want to put a key for a service account that
allows reading from Google Cloud Datastore and Google Cloud IoT Core in a dir
called `keys` and then set the `GOOGLE_APPLICATION_CREDENTIALS` env var, e.g.:

```sh
GOOGLE_APPLICATION_CREDENTIALS=/keys/my-key.json
```

In production `GOOGLE_APPLICATION_CREDENTIALS` isn't necessary because the service
will have the proper permissions granted to it.

### Build and run locally

Did you make your `.env` file and put your service account key in `keys`?
Do that first (see above).

```sh
PROJECT=my-gcp-project-id \
REPO=my-artifact-repository-repo-name \
SERVICE=my-cloud-run-service-name \
make web-image

PROJECT=my-gcp-project-id \
REPO=my-artifact-repository-repo-name \
SERVICE=my-cloud-run-service-name \
make run-web
```

### Build on Google Cloud Build

This will build the image remotely and push it to Google Artifact Repository.

```sh
PROJECT=my-gcp-project-id \
REPO=my-artifact-repository-repo-name \
SERVICE=my-cloud-run-service-name \
make web-image-remote
```

### Deploying to Cloud Run

Deploy the image built with `make web-image-remote` to Cloud Run and make sure
that the env vars (aside from `GOOGLE_CLOUD_PROJECT` and `GOOGLE_APPLICATION_CREDENTIALS`)
are made available to it.

Subsequent deploys can be done using this make command:

```sh
PROJECT=my-gcp-project-id \
REPO=my-artifact-repository-repo-name \
SERVICE=my-cloud-run-service-name \
make deploy-web
```

`make deploy-web` doesn't set env vars so it can't be used for the first deploy.
