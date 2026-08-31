module github.com/mtraver/environmental-sensor

go 1.26.5

require (
	cloud.google.com/go/compute/metadata v0.9.0
	cloud.google.com/go/datastore v1.26.0
	cloud.google.com/go/pubsub/v2 v2.7.0
	github.com/99designs/gqlgen v0.17.94
	github.com/aws/aws-lambda-go v1.55.0
	github.com/aws/aws-sdk-go-v2 v1.45.1
	github.com/aws/aws-sdk-go-v2/config v1.33.1
	github.com/aws/aws-sdk-go-v2/credentials v1.20.1
	github.com/aws/aws-sdk-go-v2/service/iot v1.81.0
	github.com/aws/aws-sdk-go-v2/service/secretsmanager v1.46.1
	github.com/aws/aws-sdk-go-v2/service/sts v1.47.1
	github.com/eclipse/paho.golang v0.23.0
	github.com/google/go-cmp v0.7.0
	github.com/influxdata/influxdb-client-go/v2 v2.14.0
	github.com/maypok86/otter/v2 v2.3.0
	github.com/mtraver/awsiotcore v0.0.0-20260721194309-2e4f4bb6b112
	github.com/mtraver/envtools v0.0.0-20260504053214-7b571519c787
	github.com/mtraver/gaelog v1.1.6
	github.com/mtraver/sds011 v0.0.0-20221026204622-d61fb9543898
	github.com/netresearch/go-cron v0.16.0
	github.com/vektah/gqlparser/v2 v2.5.36
	golang.org/x/oauth2 v0.36.0
	google.golang.org/api v0.296.0
	google.golang.org/grpc v1.83.2
	google.golang.org/protobuf v1.36.12
	periph.io/x/conn/v3 v3.7.3
	periph.io/x/devices/v3 v3.7.5-0.20260721214149-7820343353b2
	periph.io/x/host/v3 v3.8.5
)

require (
	cloud.google.com/go v0.123.0 // indirect
	cloud.google.com/go/auth v0.23.2 // indirect
	cloud.google.com/go/auth/oauth2adapt v0.2.8 // indirect
	cloud.google.com/go/iam v1.13.0 // indirect
	cloud.google.com/go/logging v1.19.1 // indirect
	cloud.google.com/go/longrunning v1.2.0 // indirect
	github.com/agnivade/levenshtein v1.2.1 // indirect
	github.com/albenik/go-serial/v2 v2.6.1 // indirect
	github.com/apapsch/go-jsonmerge/v2 v2.0.0 // indirect
	github.com/aws/aws-sdk-go-v2/feature/ec2/imds v1.19.1 // indirect
	github.com/aws/aws-sdk-go-v2/internal/configsources v1.5.1 // indirect
	github.com/aws/aws-sdk-go-v2/internal/endpoints/v2 v2.8.1 // indirect
	github.com/aws/aws-sdk-go-v2/internal/v4a v1.5.1 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/accept-encoding v1.13.19 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/presigned-url v1.14.1 // indirect
	github.com/aws/aws-sdk-go-v2/service/signin v1.7.1 // indirect
	github.com/aws/aws-sdk-go-v2/service/sso v1.35.1 // indirect
	github.com/aws/aws-sdk-go-v2/service/ssooidc v1.40.1 // indirect
	github.com/aws/smithy-go v1.28.1 // indirect
	github.com/cespare/xxhash/v2 v2.3.0 // indirect
	github.com/coder/websocket v1.8.15 // indirect
	github.com/creack/goselect v0.1.3 // indirect
	github.com/felixge/httpsnoop v1.1.0 // indirect
	github.com/go-logr/logr v1.4.4 // indirect
	github.com/go-logr/stdr v1.2.2 // indirect
	github.com/go-viper/mapstructure/v2 v2.5.0 // indirect
	github.com/goccy/go-yaml v1.19.2 // indirect
	github.com/google/s2a-go v0.1.9 // indirect
	github.com/google/uuid v1.6.0 // indirect
	github.com/googleapis/enterprise-certificate-proxy v0.3.21 // indirect
	github.com/googleapis/gax-go/v2 v2.24.0 // indirect
	github.com/gorilla/websocket v1.5.3 // indirect
	github.com/hashicorp/golang-lru/v2 v2.0.7 // indirect
	github.com/influxdata/line-protocol v0.0.0-20210922203350-b1ad95c89adf // indirect
	github.com/oapi-codegen/runtime v1.7.0 // indirect
	github.com/sosodev/duration v1.4.0 // indirect
	github.com/stretchr/testify v1.12.1 // indirect
	github.com/urfave/cli/v3 v3.10.1 // indirect
	go.opencensus.io v0.24.0 // indirect
	go.opentelemetry.io/auto/sdk v1.2.1 // indirect
	go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc v0.71.0 // indirect
	go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp v0.71.0 // indirect
	go.opentelemetry.io/otel v1.46.0 // indirect
	go.opentelemetry.io/otel/metric v1.46.0 // indirect
	go.opentelemetry.io/otel/trace v1.46.0 // indirect
	go.uber.org/multierr v1.11.0 // indirect
	go.yaml.in/yaml/v3 v3.0.5 // indirect
	golang.org/x/crypto v0.55.0 // indirect
	golang.org/x/mod v0.38.0 // indirect
	golang.org/x/net v0.58.0 // indirect
	golang.org/x/sync v0.22.0 // indirect
	golang.org/x/sys v0.47.0 // indirect
	golang.org/x/text v0.41.0 // indirect
	golang.org/x/time v0.15.0 // indirect
	golang.org/x/tools v0.48.0 // indirect
	google.golang.org/genproto v0.0.0-20260831171406-18b4a7587f8a // indirect
	google.golang.org/genproto/googleapis/api v0.0.0-20260831171406-18b4a7587f8a // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20260831171406-18b4a7587f8a // indirect
)

tool golang.org/x/tools/cmd/stringer
