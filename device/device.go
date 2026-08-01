package device

import (
	"context"

	"cloud.google.com/go/compute/metadata"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	awsiot "github.com/aws/aws-sdk-go-v2/service/iot"
	fedident "github.com/mtraver/environmental-sensor/federatedidentity"
)

// GetDevicesAWS gets all devices (called "things" in AWS IoT Core).
func GetDevicesAWS(ctx context.Context, roleARN, region string) (*awsiot.ListThingsOutput, error) {
	var cfg aws.Config

	// If we're on GCE, assume AWS role and fetch credentials.
	if metadata.OnGCE() {
		creds, err := fedident.GetCredentialsForRole(ctx, roleARN, region)
		if err != nil {
			return nil, err
		}

		cfg, err = config.LoadDefaultConfig(
			ctx, config.WithRegion(region), config.WithCredentialsProvider(creds))
		if err != nil {
			return nil, err
		}
	} else {
		var err error
		cfg, err = config.LoadDefaultConfig(ctx, config.WithRegion(region))
		if err != nil {
			return nil, err
		}
	}

	client := awsiot.NewFromConfig(cfg)
	return client.ListThings(ctx, &awsiot.ListThingsInput{})
}

// GetDeviceIDsAWS gets the IDs of all devices (called "things" in AWS IoT Core).
func GetDeviceIDsAWS(ctx context.Context, roleARN, region string) ([]string, error) {
	things, err := GetDevicesAWS(ctx, roleARN, region)
	if err != nil {
		return []string{}, err
	}

	ids := make([]string, len(things.Things))
	for i, t := range things.Things {
		ids[i] = *t.ThingName
	}

	return ids, nil
}
