# Go Client

This directory contains code that runs on the Raspberry Pi. It reads the
temperature and logs it via [Google Cloud IoT Core](https://cloud.google.com/iot-core/),
storing the data in Google Cloud Datastore or Google Cloud Bigtable (see the
Google App Engine app in the [web](web) directory).

    make

    # Log temp to Google Cloud IoT Core
    ./out/iotcorelogger -project my-gcp-project -registry my-iot-core-registry \
      -key device_key.pem -cacerts roots.pem

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
for details. The `cacerts` flag takes the path to this file.

## Building

    make

Simple as that. This will build the `iotcorelogger` program for the host
architecture as well as ARMv6 (e.g. Raspberry Pi Zero W) and ARMv7 (e.g.
Raspberry Pi 3 B<sup>1</sup>).

## Full usage

    Usage of ./iotcorelogger:
      -cacerts string
          Path to a set of trustworthy CA certs.
          Download Google's from https://pki.google.com/roots.pem.
      -interval int
          number of seconds to wait between samples (default 1)
      -key string
          path to device's private key
      -mqtthost string
          MQTT host (default "mqtt.googleapis.com")
      -mqttport int
          MQTT port (default 8883)
      -numsamples int
          number of samples to take (default 3)
      -project string
          Google Cloud Platform project ID
      -region string
          Google Cloud Platform region (default "us-central1")
      -registry string
          Google Cloud IoT Core registry ID

## Footnotes
<sup>1</sup> "How can this be!? The Raspberry Pi 3 B uses the BCM2837, a 64-bit
ARMv8 SoC!" you exclaim. "That is correct," I reply, "but Raspbian is 32-bit
only so the chip runs in 32-bit mode. It therefore cannot execute ARMv8 binaries."
