package federatedidentity

import (
	"fmt"
	"net/url"

	"cloud.google.com/go/compute/metadata"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sts"
)

var (
	awsCredentialDurationSeconds int64 = 900
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

func GetCredentialsForRole(roleARN, region string) (*credentials.Credentials, error) {
	gceToken, err := getGCEToken()
	if err != nil {
		return nil, err
	}

	// Get the GCE instance's hostname, which we'll use as the session name when assuming
	// the AWS role. The hostname is of the form "<instanceID>.c.<projID>.internal".
	hostname, err := metadata.Hostname()
	if err != nil {
		return nil, err
	}

	session, err := session.NewSession()
	if err != nil {
		return nil, err
	}
	client := sts.New(session, aws.NewConfig().WithRegion(region))

	out, err := client.AssumeRoleWithWebIdentity(&sts.AssumeRoleWithWebIdentityInput{
		DurationSeconds:  &awsCredentialDurationSeconds,
		RoleArn:          &roleARN,
		RoleSessionName:  &hostname,
		WebIdentityToken: &gceToken,
	})
	if err != nil {
		return nil, err
	}

	return credentials.NewStaticCredentials(
		*out.Credentials.AccessKeyId, *out.Credentials.SecretAccessKey, *out.Credentials.SessionToken), nil
}
