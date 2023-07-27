package device

import (
	"context"
	"fmt"
	"time"

	"golang.org/x/oauth2/google"

	"cloud.google.com/go/compute/metadata"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	awsiot "github.com/aws/aws-sdk-go/service/iot"
	fedident "github.com/mtraver/environmental-sensor/federatedidentity"
	cloudiot "google.golang.org/api/cloudiot/v1"
)

const gcpRegion = "us-central1"

func GetDevicesAWS(ctx context.Context, roleARN, region string) (*awsiot.ListThingsOutput, error) {
	config := aws.NewConfig().WithRegion(region)

	// If we're on GCE, assume AWS role and fetch credentials.
	if metadata.OnGCE() {
		creds, err := fedident.GetCredentialsForRole(roleARN, region)
		if err != nil {
			return nil, err
		}
		config = config.WithCredentials(creds)
	}

	session, err := session.NewSession()
	if err != nil {
		return nil, err
	}
	client := awsiot.New(session, config)
	return client.ListThings(nil)
}

func getRegistryPath(projectID, registryID string) string {
	return fmt.Sprintf("projects/%s/locations/%s/registries/%s",
		projectID, gcpRegion, registryID)
}

// GetDevices returns a list of the devices in the given registry.
func GetDevices(ctx context.Context, projectID, registryID string) ([]*cloudiot.Device, error) {
	client, err := google.DefaultClient(ctx, cloudiot.CloudiotScope)
	if err != nil {
		return []*cloudiot.Device{}, err
	}
	client.Timeout = time.Second * 10

	cloudiotService, err := cloudiot.New(client)
	if err != nil {
		return []*cloudiot.Device{}, err
	}

	response, err := cloudiotService.Projects.Locations.Registries.Devices.List(
		getRegistryPath(projectID, registryID)).Do()
	if err != nil {
		return []*cloudiot.Device{}, err
	}

	return response.Devices, nil
}

// GetDeviceIDs returns a list of the IDs (as strings) of the devices in the given registry.
func GetDeviceIDs(ctx context.Context, projectID, registryID string) ([]string, error) {
	devices, err := GetDevices(ctx, projectID, registryID)
	if err != nil {
		return []string{}, err
	}

	ids := make([]string, len(devices))
	for i := range devices {
		ids[i] = devices[i].Id
	}

	return ids, nil
}
