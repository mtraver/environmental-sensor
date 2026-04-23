module github.com/mtraver/environmental-sensor

go 1.25.5

require (
	cloud.google.com/go/compute/metadata v0.9.0
	cloud.google.com/go/datastore v1.22.0
	cloud.google.com/go/pubsub v1.50.2
	github.com/99designs/gqlgen v0.17.89
	github.com/aws/aws-lambda-go v1.54.0
	github.com/aws/aws-sdk-go-v2 v1.41.6
	github.com/aws/aws-sdk-go-v2/config v1.32.16
	github.com/aws/aws-sdk-go-v2/credentials v1.19.15
	github.com/aws/aws-sdk-go-v2/service/iot v1.72.7
	github.com/aws/aws-sdk-go-v2/service/secretsmanager v1.41.6
	github.com/aws/aws-sdk-go-v2/service/sts v1.42.0
	github.com/eclipse/paho.mqtt.golang v1.5.1
	github.com/google/go-cmp v0.7.0
	github.com/influxdata/influxdb-client-go/v2 v2.14.0
	github.com/mitchellh/go-homedir v1.1.0
	github.com/mtraver/awsiotcore v0.0.0-20260421190623-3b8ec2ce6466
	github.com/mtraver/envtools v0.0.0-20250815190129-f1d680d4b06a
	github.com/mtraver/gaelog v1.1.6
	github.com/mtraver/iotcore v0.0.0-20260421192442-a154403040db
	github.com/mtraver/sds011 v0.0.0-20221026204622-d61fb9543898
	github.com/robfig/cron/v3 v3.0.1
	github.com/vektah/gqlparser/v2 v2.5.32
	golang.org/x/oauth2 v0.36.0
	google.golang.org/api v0.276.0
	google.golang.org/appengine v1.6.8
	google.golang.org/grpc v1.80.0
	google.golang.org/protobuf v1.36.11
	periph.io/x/conn/v3 v3.7.3
	periph.io/x/devices/v3 v3.7.4
	periph.io/x/host/v3 v3.8.5
)

require (
	cloud.google.com/go v0.123.0 // indirect
	cloud.google.com/go/auth v0.20.0 // indirect
	cloud.google.com/go/auth/oauth2adapt v0.2.8 // indirect
	cloud.google.com/go/iam v1.9.0 // indirect
	cloud.google.com/go/logging v1.16.0 // indirect
	cloud.google.com/go/longrunning v0.11.0 // indirect
	cloud.google.com/go/pubsub/v2 v2.6.0 // indirect
	github.com/agnivade/levenshtein v1.2.1 // indirect
	github.com/albenik/go-serial/v2 v2.6.1 // indirect
	github.com/apapsch/go-jsonmerge/v2 v2.0.0 // indirect
	github.com/aws/aws-sdk-go-v2/feature/ec2/imds v1.18.22 // indirect
	github.com/aws/aws-sdk-go-v2/internal/configsources v1.4.22 // indirect
	github.com/aws/aws-sdk-go-v2/internal/endpoints/v2 v2.7.22 // indirect
	github.com/aws/aws-sdk-go-v2/internal/v4a v1.4.23 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/accept-encoding v1.13.8 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/presigned-url v1.13.22 // indirect
	github.com/aws/aws-sdk-go-v2/service/signin v1.0.10 // indirect
	github.com/aws/aws-sdk-go-v2/service/sso v1.30.16 // indirect
	github.com/aws/aws-sdk-go-v2/service/ssooidc v1.35.20 // indirect
	github.com/aws/smithy-go v1.25.0 // indirect
	github.com/cespare/xxhash/v2 v2.3.0 // indirect
	github.com/creack/goselect v0.1.3 // indirect
	github.com/felixge/httpsnoop v1.0.4 // indirect
	github.com/go-logr/logr v1.4.3 // indirect
	github.com/go-logr/stdr v1.2.2 // indirect
	github.com/go-viper/mapstructure/v2 v2.5.0 // indirect
	github.com/goccy/go-yaml v1.19.2 // indirect
	github.com/golang-jwt/jwt/v4 v4.5.2 // indirect
	github.com/golang/protobuf v1.5.4 // indirect
	github.com/google/s2a-go v0.1.9 // indirect
	github.com/google/uuid v1.6.0 // indirect
	github.com/googleapis/enterprise-certificate-proxy v0.3.15 // indirect
	github.com/googleapis/gax-go/v2 v2.22.0 // indirect
	github.com/gorilla/websocket v1.5.3 // indirect
	github.com/hashicorp/golang-lru/v2 v2.0.7 // indirect
	github.com/influxdata/line-protocol v0.0.0-20210922203350-b1ad95c89adf // indirect
	github.com/oapi-codegen/runtime v1.4.0 // indirect
	github.com/sosodev/duration v1.4.0 // indirect
	github.com/urfave/cli/v3 v3.7.0 // indirect
	go.opencensus.io v0.24.0 // indirect
	go.opentelemetry.io/auto/sdk v1.2.1 // indirect
	go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc v0.68.0 // indirect
	go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp v0.68.0 // indirect
	go.opentelemetry.io/otel v1.43.0 // indirect
	go.opentelemetry.io/otel/metric v1.43.0 // indirect
	go.opentelemetry.io/otel/trace v1.43.0 // indirect
	go.uber.org/multierr v1.11.0 // indirect
	golang.org/x/crypto v0.50.0 // indirect
	golang.org/x/mod v0.34.0 // indirect
	golang.org/x/net v0.53.0 // indirect
	golang.org/x/sync v0.20.0 // indirect
	golang.org/x/sys v0.43.0 // indirect
	golang.org/x/text v0.36.0 // indirect
	golang.org/x/time v0.15.0 // indirect
	golang.org/x/tools v0.43.0 // indirect
	google.golang.org/genproto v0.0.0-20260420184626-e10c466a9529 // indirect
	google.golang.org/genproto/googleapis/api v0.0.0-20260420184626-e10c466a9529 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20260420184626-e10c466a9529 // indirect
)
