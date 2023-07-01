// Binary lambda is an AWS Lambda function that receives IoT telemetry messages and re-publishes them to Google Cloud Pub/Sub.
package main

import (
	"context"
	"encoding/base64"
	"log"

	"cloud.google.com/go/pubsub"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
	mpb "github.com/mtraver/environmental-sensor/measurementpb"
	"github.com/mtraver/envtools"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/option"
	"google.golang.org/protobuf/proto"
)

const (
	gcpCredentialsScope = "https://www.googleapis.com/auth/pubsub"
)

var (
	gcpProjectID    = envtools.MustGetenv("GCP_PROJECT_ID")
	pubSubTopicName = envtools.MustGetenv("GCP_PUBSUB_TOPIC")

	secretsManagerRegion = envtools.MustGetenv("AWS_SECRETS_MANAGER_REGION")
	secretName           = envtools.MustGetenv("GCP_CREDENTIALS_SECRET_NAME")
)

func getServiceAccountKey(ctx context.Context) (*google.Credentials, error) {
	config, err := config.LoadDefaultConfig(ctx, config.WithRegion(secretsManagerRegion))
	if err != nil {
		return nil, err
	}
	client := secretsmanager.NewFromConfig(config)

	input := &secretsmanager.GetSecretValueInput{
		SecretId:     aws.String(secretName),
		VersionStage: aws.String("AWSCURRENT"), // VersionStage defaults to AWSCURRENT if unspecified
	}

	result, err := client.GetSecretValue(ctx, input)
	if err != nil {
		return nil, err
	}

	// Decrypt the secret using the associated KMS key.
	key := *result.SecretString

	return google.CredentialsFromJSON(ctx, []byte(key), gcpCredentialsScope)
}

func handle(ctx context.Context, message string) (string, error) {
	// message is unmarshaled from JSON before being passed into this function, and
	// because on the other end the message was made by JSON-encoding a byte slice,
	// the unmarshaled JSON here is a base64-encoded string. Decode from base64 to get
	// the bytes of the marshaled protobuf.
	protoBytes, err := base64.StdEncoding.DecodeString(message)
	if err != nil {
		log.Printf("Failed to decode base64: %v", err)
		return "error", err
	}

	// Unmarshal the protobuf to make sure that it is indeed a valid Measurement that we've received.
	m := &mpb.Measurement{}
	if err := proto.Unmarshal(protoBytes, m); err != nil {
		log.Printf("Failed to unmarshal protobuf: %v", err)
		return "error", err
	}

	serviceAccountKey, err := getServiceAccountKey(ctx)
	if err != nil {
		log.Printf("Failed to get service account key from Secrets Manager: %v", err)
		return "error", err
	}

	client, err := pubsub.NewClient(ctx, gcpProjectID, option.WithCredentials(serviceAccountKey))
	if err != nil {
		log.Printf("Failed to make Pub/Sub client: %v", err)
		return "error", err
	}

	topic := client.Topic(pubSubTopicName)
	defer topic.Stop()

	r := topic.Publish(ctx, &pubsub.Message{
		Data: protoBytes,
		Attributes: map[string]string{
			"source": "AWS",
		},
	})

	if _, err := r.Get(ctx); err != nil {
		log.Printf("Failed to publish to Pub/Sub: %v", err)
		return "error", err
	}

	return "ok", nil
}

func main() {
	lambda.Start(handle)
}
