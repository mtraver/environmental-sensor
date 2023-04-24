# Go CLI Programs

[iotcorelogger](iotcorelogger) and [readtemp](readtemp) contain code that runs
on the Raspberry Pi. `iotcorelogger` reads the temperature and logs it via
[Google Cloud IoT Core](https://cloud.google.com/iot-core/), storing the data
in Google Cloud Datastore (see the Google App Engine app in the [web](../web)
directory).

From the root of this repository,

    make

    # Log temp to Google Cloud IoT Core

    # Example config.pb.json
    # {
    #   "supported_sensors": ["mcp9808"],
    #
    #   "jobs": [
    #     {
    #       "cronspec": "0 */2 * * * *",
    #       "operation": "SENSE",
    #       "sensors": ["mcp9808"]
    #     }
    #   ]
    # }
    #
    # Example device.json:
    # {
    #   "project_id": "my-gcp-project",
    #   "registry_id": "my-iot-core-registry",
    #   "device_id": "my-device",
    #   "ca_certs_path": "roots.pem",
    #   "priv_key_path": "my-device.pem",
    #   "region": "us-central1"
    # }
    ./out/iotcorelogger -config config.pb.json -gcp-device device.json

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

On the Raspberry Pi:

    wget https://pki.google.com/roots.pem

This is a set of trustworthy root certificates. See [here](http://pki.google.com/faq.html)
for details. The path to this file is the value of `"ca_certs_path"` in the device file.

## Building

From the root of this repository,

    make

Simple as that. This will build the `iotcorelogger` program for the host
architecture as well as ARMv6 (e.g. Raspberry Pi Zero W) and ARMv7 (e.g.
Raspberry Pi 3 B<sup>1</sup>).

## Full usage

    Usage of iotcorelogger:
      -config string
          path to a file containing a JSON-encoded config proto
      -dryrun
          set to true to print rather than publish measurements
      -gcp-device string
          path to a JSON file describing a GCP IoT Core device. See github.com/mtraver/iotcore.
      -port int
          port on which the device's web server should listen (default 8080)

## Footnotes
<sup>1</sup> "How can this be!? The Raspberry Pi 3 B uses the BCM2837, a 64-bit
ARMv8 SoC!" you exclaim. "That is correct," I reply, "but Raspbian is 32-bit
only so the chip runs in 32-bit mode. It therefore cannot execute ARMv8 binaries."
