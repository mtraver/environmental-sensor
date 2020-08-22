// Binary api implements the gRPC service MeasurementService.
package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net"
	"os"

	"cloud.google.com/go/compute/metadata"
	"github.com/mtraver/environmental-sensor/measurement"
	mpb "github.com/mtraver/environmental-sensor/measurementpb"
	"github.com/mtraver/environmental-sensor/web/db"
	"github.com/mtraver/environmental-sensor/web/device"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/emptypb"
)

const (
	datastoreKind = "measurement"
)

var (
	port int
)

type Database interface {
	Latest(ctx context.Context, deviceIDs []string) (map[string]measurement.StorableMeasurement, error)
}

type apiServer struct {
	mpb.UnimplementedMeasurementServiceServer
	projectID  string
	registryID string
	database   Database
}

func (s *apiServer) GetDevices(ctx context.Context, in *emptypb.Empty) (*mpb.GetDevicesResponse, error) {
	devices, err := device.GetDevices(ctx, s.projectID, s.registryID)
	if err != nil {
		return nil, err
	}

	deviceIDs := make([]string, 0, len(devices))
	for _, d := range devices {
		deviceIDs = append(deviceIDs, d.Id)
	}

	return &mpb.GetDevicesResponse{
		DeviceId: deviceIDs,
	}, nil
}

func (s *apiServer) GetLatest(ctx context.Context, r *mpb.GetLatestRequest) (*mpb.Measurement, error) {
	latest, err := s.database.Latest(ctx, []string{r.GetDeviceId()})
	if err != nil {
		return nil, err
	}

	sm, ok := latest[r.GetDeviceId()]
	if !ok {
		return nil, fmt.Errorf("api: device ID %q not found", r.GetDeviceId())
	}

	m, err := measurement.NewMeasurement(&sm)
	if err != nil {
		return nil, err
	}

	return &m, nil
}

func init() {
	flag.IntVar(&port, "p", 9090, "port on which the gRPC server will listen")

	flag.Usage = func() {
		message := `usage: api registry

Positional Arguments (required):
  registry
	the ID of the Google Cloud IoT Core registry to query for device information

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
	registryID := flag.Args()[0]

	projectID := os.Getenv("PROJECT_ID")
	if projectID == "" && metadata.OnGCE() {
		var err error
		projectID, err = metadata.ProjectID()
		if err != nil {
			log.Fatalf("Failed to get project ID: %v", err)
		}
	}

	database, err := db.NewDatastoreDB(projectID, datastoreKind, false)
	if err != nil {
		log.Fatalf("Failed to make datastore DB: %v", err)
	}

	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	grpcServer := grpc.NewServer()
	mpb.RegisterMeasurementServiceServer(grpcServer, &apiServer{
		projectID:  projectID,
		registryID: registryID,
		database:   database,
	})

	log.Printf("gRPC server listening on port %d", port)
	log.Fatal(grpcServer.Serve(lis))
}
