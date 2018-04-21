package device

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"

	cloudiot "google.golang.org/api/cloudiot/v1"
	"google.golang.org/appengine/urlfetch"
)

const region = "us-central1"

func getRegistryPath(projectID, registryID string) string {
	return fmt.Sprintf("projects/%s/locations/%s/registries/%s",
		projectID, region, registryID)
}

// GetDevices returns a list of the devices in the given registry.
func GetDevices(ctx context.Context, projectID, registryID string) ([]*cloudiot.Device, error) {
	// We need a client that supports OAuth2 *and* that uses the urlfetch
	// transport, so we have to create it manually instead of using either
	// package's helper functions.
	client := &http.Client{
		Transport: &oauth2.Transport{
			Source: google.AppEngineTokenSource(ctx, cloudiot.CloudiotScope),
			Base:   &urlfetch.Transport{Context: ctx},
		},
		Timeout: time.Second * 10,
	}

	cloudiotService, err := cloudiot.New(client)

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
