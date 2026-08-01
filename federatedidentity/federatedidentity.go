package federatedidentity

import (
	"context"
	"fmt"
	"net/url"
	"os"
	"time"

	"cloud.google.com/go/compute/metadata"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/sts"
	"github.com/maypok86/otter/v2"
)

const (
	cacheTTL = time.Second * 840
)

var (
	awsCredentialDurationSeconds int32 = 900

	cache = otter.Must(&otter.Options[string, aws.CredentialsProvider]{
		MaximumSize:      100,
		ExpiryCalculator: otter.ExpiryWriting[string, aws.CredentialsProvider](cacheTTL),
	})
)

func getGCEToken() (string, error) {
	if !metadata.OnGCE() {
		return "", fmt.Errorf("not on GCE or similar environment so cannot get token from metadata service")
	}

	v := url.Values{}
	v.Set("audience", "gcp")
	suffix := fmt.Sprintf("instance/service-accounts/default/identity?%v", v.Encode())

	return metadata.Get(suffix)
}

func GetCredentialsForRole(ctx context.Context, roleARN, region string) (aws.CredentialsProvider, error) {
	entry, ok := cache.GetEntry(roleARN)
	if ok && entry.Value != nil {
		return entry.Value, nil
	}

	gceToken, err := getGCEToken()
	if err != nil {
		return nil, err
	}

	// Use an identifier of the environment we're running in as the session name when assuming
	// the AWS role. First try the Cloud Run revision. If that's not present, then try to get
	// the GCE instance's hostname.
	sessionName := os.Getenv("K_REVISION")
	if sessionName == "" {
		// The hostname is of the form "<instanceID>.c.<projID>.internal".
		hostname, err := metadata.Hostname()
		if err != nil {
			return nil, err
		}
		sessionName = hostname
	}

	// Load minimal config to create an STS client.
	cfg, err := config.LoadDefaultConfig(
		ctx, config.WithRegion(region), config.WithCredentialsProvider(aws.AnonymousCredentials{}))
	if err != nil {
		return nil, err
	}
	client := sts.NewFromConfig(cfg)

	out, err := client.AssumeRoleWithWebIdentity(ctx, &sts.AssumeRoleWithWebIdentityInput{
		DurationSeconds:  &awsCredentialDurationSeconds,
		RoleArn:          &roleARN,
		RoleSessionName:  &sessionName,
		WebIdentityToken: &gceToken,
	})
	if err != nil {
		return nil, err
	}

	cred := credentials.NewStaticCredentialsProvider(
		*out.Credentials.AccessKeyId,
		*out.Credentials.SecretAccessKey,
		*out.Credentials.SessionToken,
	)

	cache.Set(roleARN, cred)

	return cred, nil
}
