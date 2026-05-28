# Go CLI Programs

[iotcorelogger](iotcorelogger) and [readtemp](readtemp) contain code that runs
on the Raspberry Pi. `iotcorelogger` reads the temperature and logs it via
[AWS IoT Core](https://aws.amazon.com/iot-core/), storing the data in Google
Cloud Datastore (see the Docker image in the [web](../web) directory).

From the root of this repository,

    make

    # Log temp to AWS IoT Core

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
    #   "endpoint": "endpoint-name",
    #   "device_id": "my-device",
    #   "ca_certs_path": "amazon_root_cas.pem",
    #   "cert_path": "my-device.x509",
    #   "priv_key_path": "my-device.pem"
    # }
    ./out/iotcorelogger -config config.pb.json -aws-device device.json

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

Finally, you'll need Amazon's root CA certificates. Information on AWS IoT Core's
usage of the root CA certs and links to download them can be found here:
https://docs.aws.amazon.com/iot/latest/developerguide/server-authentication.html#server-authentication-certs.

Download the relevant CA certs and save them on your Raspberry Pi. Set `"ca_certs_path"`
in your device file to the path of the file containing the cert(s).

## Building

From the root of this repository,

    make

Simple as that. This will build the `iotcorelogger` program for the host
architecture as well as ARMv6 (e.g. Raspberry Pi Zero W) and ARMv7 (e.g.
Raspberry Pi 3 B<sup>1</sup>).

## Full usage

    Usage of iotcorelogger:
      -aws-device string
          path to a device config file describing an AWS IoT Core device
      -config string
          path to a file containing a JSON-encoded config proto
      -dryrun
          set to true to print rather than publish measurements
      -port int
          port on which the device's web server should listen (default 8080)

## Footnotes
<sup>1</sup> "How can this be!? The Raspberry Pi 3 B uses the BCM2837, a 64-bit
ARMv8 SoC!" you exclaim. "That is correct," I reply, "but Raspbian is 32-bit
only so the chip runs in 32-bit mode. It therefore cannot execute ARMv8 binaries."
