package main

import (
	"bufio"
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"os"
	"path"
	"strconv"
	"time"

	"golang.org/x/net/context"

	"cloud.google.com/go/pubsub"
	"github.com/golang/protobuf/proto"
	"github.com/golang/protobuf/ptypes"

	"receiver/measurement"
)

const usageStr = `usage: %v csv_file project_id topic_id device_id

Reads lines from a CSV file, converts them into Measurement protobufs as defined
in measurement.proto, and publishes them to Google Cloud Pub/Sub. It's expected
that an instance of the receiver app is running on App Engine to process the
Pub/Sub messages.

This program was written to migrate data stored in a Google Sheet to the
storage back ends supported by the receiver app.

The first line of the CSV file is expected to be column headers, followed by
lines of this format:

  timestamp,temp1,temp2,...

The timestamp field must be formatted like this: %v

There must be at least one temperature value. If there is more than one
they are averaged before packing into the Measurement protobuf.

Arguments:
  csv_file: Path to the CSV file.
  project_id: Google Cloud Platform project ID.
  topic_id: Google Cloud Pub/Sub topic ID. This topic must be configured to
            push to the receiver app's Pub/Sub endpoint URL.
  device_id: The device ID to include in the Pub/Sub message. Data from multiple
             devices can be saved and displayed by the receiver app, so this
             must be a unique identifier for the device from which the data in
             the CSV file came.
`

const timeFormat = "2006-01-02T15:04:05.999999"

func strsToFloats(x []string) ([]float32, error) {
	var numbers []float32
	for _, v := range x {
		f, err := strconv.ParseFloat(v, 32)
		if err != nil {
			return numbers, err
		}
		numbers = append(numbers, float32(f))
	}
	return numbers, nil
}

func mean(x []float32) float32 {
	var total float32
	for _, v := range x {
		total += v
	}
	return total / float32(len(x))
}

func lineToProto(
	line []string, deviceID string) (*measurement.Measurement, error) {
	if len(line) < 2 {
		return nil, errors.New(
			"Line length must be at least 2 (timestamp and one measurement)")
	}

	// Convert the timestamp string to a timestamp for packing into a protobuf
	timestamp, err := time.Parse(timeFormat, line[0])
	if err != nil {
		return nil, err
	}
	timestampProto, err := ptypes.TimestampProto(timestamp)
	if err != nil {
		return nil, err
	}

	temps, err := strsToFloats(line[1:])
	if err != nil {
		return nil, err
	}

	m := &measurement.Measurement{
		DeviceId:  deviceID,
		Timestamp: timestampProto,
		Temp:      mean(temps),
	}

	return m, nil
}

func publishFromCSV(csvFile string, projectID string, pubsubTopic string,
	deviceID string) error {
	f, err := os.Open(csvFile)
	defer f.Close()
	if err != nil {
		return err
	}
	reader := csv.NewReader(bufio.NewReader(f))

	// Set up the Pub/Sub client
	ctx := context.Background()
	client, err := pubsub.NewClient(ctx, projectID)
	if err != nil {
		return err
	}

	if exists, err := client.Topic(pubsubTopic).Exists(ctx); err != nil {
		return err
	} else if !exists {
		return fmt.Errorf("Topic does not exist. Create it first: %v", pubsubTopic)
	}

	topic := client.Topic(pubsubTopic)
	defer topic.Stop()

	topic.PublishSettings = pubsub.PublishSettings{
		DelayThreshold: 10 * time.Millisecond,
		CountThreshold: 50,
	}

	// Assume that the CSV file has column headers, and strip them off first
	if _, err := reader.Read(); err == io.EOF {
		return err
	}

	for {
		line, err := reader.Read()
		if err == io.EOF {
			break
		} else if err != nil {
			return err
		}

		// Make and marshal the protobuf
		m, err := lineToProto(line, deviceID)
		if err != nil {
			return err
		}
		data, err := proto.Marshal(m)
		if err != nil {
			return err
		}

		topic.Publish(ctx, &pubsub.Message{Data: data})
	}

	return nil
}

func usage() {
	fmt.Printf(usageStr, path.Base(os.Args[0]), timeFormat)
}

func main() {
	if len(os.Args) != 5 {
		usage()
		os.Exit(2)
	}

	err := publishFromCSV(os.Args[1], os.Args[2], os.Args[3], os.Args[4])
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
