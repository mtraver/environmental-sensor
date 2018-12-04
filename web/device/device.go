package device

import (
	"context"
	"fmt"
	"time"

	"golang.org/x/oauth2/google"

	cloudiot "google.golang.org/api/cloudiot/v1"
)

const region = "us-central1"

func getRegistryPath(projectID, registryID string) string {
	return fmt.Sprintf("projects/%s/locations/%s/registries/%s",
		projectID, region, registryID)
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
