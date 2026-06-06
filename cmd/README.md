# Go CLI Programs

[iotcorelogger](iotcorelogger) and [readtemp](readtemp) contain code that runs
on the Raspberry Pi. `iotcorelogger` reads the temperature and logs it via
[AWS IoT Core](https://aws.amazon.com/iot-core/), storing the data in Google
Cloud Datastore (see the Docker image in the [web](../web) directory).

From the root of this repository,

    make

    # Example device.json:
    # {
    #   "endpoint": "endpoint-name",
    #   "device_id": "my-device",
    #   "cert_path": "my-device.x509",
    #   "priv_key_path": "my-device.pem"
    # }
    ./out/iotcorelogger -aws-device device.json

    # Print temp to stdout
    ./out/readtemp

## Prerequisites

On your development machine / where you'll build (No, you do not need to build
on the Raspberry Pi! In fact it is slow and painful to do so.):

  1. Don't have Go installed? It's [super easy](https://golang.org/doc/install).
  2. You'll need the protocol buffer compiler, version 3.0.0 or higher. Follow
  the instructions [here](https://github.com/google/protobuf) — all you have to
  do is download a pre-built release for your platform and make sure the compiler,
  `protoc`, is on your `PATH`.
  3. You'll also need the protobuf compiler plugin that generates Go code. Follow
  the instructions [here](https://github.com/golang/protobuf), or TL;DR:
  `go get -u github.com/golang/protobuf/protoc-gen-go`

## Building

From the root of this repository,

    make

Simple as that. This will build the `iotcorelogger` program for the host
architecture as well as ARMv6 (e.g. Raspberry Pi Zero W) and ARMv7 (e.g.
Raspberry Pi 3 B<sup>1</sup>).

## Full usage

    usage: iotcorelogger [options]

    Options:
      -aws-device string
          path to a device config file describing an AWS IoT Core device
      -dryrun
          set to true to print rather than publish measurements
      -port int
          port on which the device's web server should listen (default 8080)

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

## Footnotes
<sup>1</sup> "How can this be!? The Raspberry Pi 3 B uses the BCM2837, a 64-bit
ARMv8 SoC!" you exclaim. "That is correct," I reply, "but Raspbian is 32-bit
only so the chip runs in 32-bit mode. It therefore cannot execute ARMv8 binaries."
