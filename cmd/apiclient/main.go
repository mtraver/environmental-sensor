// Binary apiclient is a command line tool for calling the gRPC service MeasurementService.
package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/golang/protobuf/ptypes/empty"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"

	mpb "github.com/mtraver/environmental-sensor/measurementpb"
)

var (
	key   string
	token string
)

func devices(ctx context.Context, client mpb.MeasurementServiceClient) (*mpb.GetDevicesResponse, error) {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	return client.GetDevices(ctx, &empty.Empty{})
}

func latest(ctx context.Context, client mpb.MeasurementServiceClient, deviceID string) (*mpb.Measurement, error) {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	return client.GetLatest(ctx, &mpb.GetLatestRequest{DeviceId: deviceID})
}

func init() {
	flag.StringVar(&key, "k", "", "API key")
	flag.StringVar(&token, "t", "", "JWT")

	flag.Usage = func() {
		message := `usage: apiclient [options] ip

Positional Arguments (required):
  ip
	the IP address of the server, including the port

Options:
`

		fmt.Fprintf(flag.CommandLine.Output(), message)
		flag.PrintDefaults()
	}
}

func main() {
	flag.Parse()

	if len(flag.Args()) != 1 {
		flag.Usage()
		os.Exit(1)
	}
	serverAddr := flag.Args()[0]

	// Set up authentication metadata.
	ctx := context.Background()
	if key != "" {
		ctx = metadata.NewOutgoingContext(ctx, metadata.Pairs("x-api-key", key))
	}
	if token != "" {
		ctx = metadata.NewOutgoingContext(ctx, metadata.Pairs("Authorization", fmt.Sprintf("Bearer %s", token)))
	}

	conn, err := grpc.Dial(serverAddr, grpc.WithInsecure())
	if err != nil {
		log.Fatalf("Failed to dial: %v", err)
	}

	client := mpb.NewMeasurementServiceClient(conn)

	devices, err := devices(ctx, client)
	if err != nil {
		log.Fatalf("Failed to GetDevices: %v", err)
	}
	fmt.Printf("%v\n", devices)

	for _, d := range devices.DeviceId {
		m, err := latest(ctx, client, d)
		if err != nil {
			log.Fatalf("Failed to GetLatest for %q: %v", d, err)
		}
		fmt.Printf("%v\n", m)
	}
}
