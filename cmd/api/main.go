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
	"github.com/mtraver/envtools"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/emptypb"
)

const (
	datastoreKind = "measurement"

	// awsRoleARNEnvVar is the name of the env var that should contain the ARN of the
	// AWS role that we'll assume and use to authenticate with AWS IoT to fetch the
	// list of devices.
	awsRoleARNEnvVar = "AWS_ROLE_ARN"

	awsRegionEnvVar = "AWS_REGION"
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
	awsRoleARN string
	awsRegion  string
	database   Database
}

func (s *apiServer) GetDevices(ctx context.Context, in *emptypb.Empty) (*mpb.GetDevicesResponse, error) {
	deviceIDs, err := device.GetDeviceIDsAWS(ctx, s.awsRoleARN, s.awsRegion)
	if err != nil {
		return nil, err
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

		fmt.Fprint(flag.CommandLine.Output(), message)
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

	onGCE := metadata.OnGCE()

	projectID := os.Getenv("PROJECT_ID")
	if projectID == "" && onGCE {
		var err error
		projectID, err = metadata.ProjectID()
		if err != nil {
			log.Fatalf("Failed to get project ID: %v", err)
		}
	}

	roleARN := os.Getenv(awsRoleARNEnvVar)
	if roleARN == "" && onGCE {
		log.Printf("On GCE and $%s is not set. Fetching devices will probably fail.", awsRoleARNEnvVar)
	}

	database, err := db.NewDatastoreDB(projectID, datastoreKind, nil)
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
		awsRoleARN: roleARN,
		awsRegion:  envtools.MustGetenv(awsRegionEnvVar),
		database:   database,
	})

	log.Printf("gRPC server listening on port %d", port)
	log.Fatal(grpcServer.Serve(lis))
}
