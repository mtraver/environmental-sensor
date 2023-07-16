module github.com/mtraver/environmental-sensor

go 1.20

require (
	cloud.google.com/go/compute/metadata v0.2.3
	cloud.google.com/go/datastore v1.12.1
	cloud.google.com/go/pubsub v1.32.0
	github.com/aws/aws-lambda-go v1.41.0
	github.com/aws/aws-sdk-go-v2 v1.19.0
	github.com/aws/aws-sdk-go-v2/config v1.18.28
	github.com/aws/aws-sdk-go-v2/service/secretsmanager v1.19.11
	github.com/eclipse/paho.mqtt.golang v1.4.3
	github.com/google/go-cmp v0.5.9
	github.com/influxdata/influxdb-client-go/v2 v2.12.3
	github.com/mitchellh/go-homedir v1.1.0
	github.com/mtraver/awsiotcore v0.0.0-20230516040800-8cdc5f352d44
	github.com/mtraver/envtools v0.0.0-20221024021805-ac686dd068a6
	github.com/mtraver/gaelog v1.1.3
	github.com/mtraver/iotcore v0.0.0-20230423225757-7fc79eb2c3c4
	github.com/mtraver/sds011 v0.0.0-20221026204622-d61fb9543898
	github.com/robfig/cron/v3 v3.0.1
	golang.org/x/oauth2 v0.10.0
	google.golang.org/api v0.131.0
	google.golang.org/appengine v1.6.7
	google.golang.org/grpc v1.56.2
	google.golang.org/protobuf v1.31.0
	periph.io/x/conn/v3 v3.7.0
	periph.io/x/devices/v3 v3.7.1
	periph.io/x/host/v3 v3.8.2
)

require (
	cloud.google.com/go v0.110.6 // indirect
	cloud.google.com/go/compute v1.21.0 // indirect
	cloud.google.com/go/iam v1.1.1 // indirect
	cloud.google.com/go/logging v1.7.0 // indirect
	cloud.google.com/go/longrunning v0.5.1 // indirect
	github.com/albenik/go-serial/v2 v2.6.0 // indirect
	github.com/apapsch/go-jsonmerge/v2 v2.0.0 // indirect
	github.com/aws/aws-sdk-go-v2/credentials v1.13.27 // indirect
	github.com/aws/aws-sdk-go-v2/feature/ec2/imds v1.13.5 // indirect
	github.com/aws/aws-sdk-go-v2/internal/configsources v1.1.35 // indirect
	github.com/aws/aws-sdk-go-v2/internal/endpoints/v2 v2.4.29 // indirect
	github.com/aws/aws-sdk-go-v2/internal/ini v1.3.36 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/presigned-url v1.9.29 // indirect
	github.com/aws/aws-sdk-go-v2/service/sso v1.12.13 // indirect
	github.com/aws/aws-sdk-go-v2/service/ssooidc v1.14.13 // indirect
	github.com/aws/aws-sdk-go-v2/service/sts v1.19.3 // indirect
	github.com/aws/smithy-go v1.13.5 // indirect
	github.com/bytedance/sonic v1.9.2 // indirect
	github.com/chenzhuoyu/base64x v0.0.0-20221115062448-fe3a3abad311 // indirect
	github.com/creack/goselect v0.1.2 // indirect
	github.com/deepmap/oapi-codegen v1.13.0 // indirect
	github.com/gabriel-vasile/mimetype v1.4.2 // indirect
	github.com/gin-contrib/sse v0.1.0 // indirect
	github.com/gin-gonic/gin v1.9.1 // indirect
	github.com/go-playground/locales v0.14.1 // indirect
	github.com/go-playground/universal-translator v0.18.1 // indirect
	github.com/go-playground/validator/v10 v10.14.1 // indirect
	github.com/goccy/go-json v0.10.2 // indirect
	github.com/golang-jwt/jwt/v4 v4.5.0 // indirect
	github.com/golang/groupcache v0.0.0-20210331224755-41bb18bfe9da // indirect
	github.com/golang/protobuf v1.5.3 // indirect
	github.com/google/s2a-go v0.1.4 // indirect
	github.com/google/uuid v1.3.0 // indirect
	github.com/googleapis/enterprise-certificate-proxy v0.2.5 // indirect
	github.com/googleapis/gax-go/v2 v2.12.0 // indirect
	github.com/gorilla/websocket v1.5.0 // indirect
	github.com/influxdata/line-protocol v0.0.0-20210922203350-b1ad95c89adf // indirect
	github.com/json-iterator/go v1.1.12 // indirect
	github.com/klauspost/cpuid/v2 v2.2.5 // indirect
	github.com/labstack/echo/v4 v4.11.0 // indirect
	github.com/labstack/gommon v0.4.0 // indirect
	github.com/leodido/go-urn v1.2.4 // indirect
	github.com/mattn/go-colorable v0.1.13 // indirect
	github.com/mattn/go-isatty v0.0.19 // indirect
	github.com/modern-go/concurrent v0.0.0-20180306012644-bacd9c7ef1dd // indirect
	github.com/modern-go/reflect2 v1.0.2 // indirect
	github.com/pelletier/go-toml/v2 v2.0.9 // indirect
	github.com/pkg/errors v0.9.1 // indirect
	github.com/twitchyliquid64/golang-asm v0.15.1 // indirect
	github.com/ugorji/go/codec v1.2.11 // indirect
	github.com/valyala/bytebufferpool v1.0.0 // indirect
	github.com/valyala/fasttemplate v1.2.2 // indirect
	go.opencensus.io v0.24.0 // indirect
	go.uber.org/multierr v1.11.0 // indirect
	golang.org/x/arch v0.4.0 // indirect
	golang.org/x/crypto v0.11.0 // indirect
	golang.org/x/net v0.12.0 // indirect
	golang.org/x/sync v0.3.0 // indirect
	golang.org/x/sys v0.10.0 // indirect
	golang.org/x/text v0.11.0 // indirect
	golang.org/x/xerrors v0.0.0-20220907171357-04be3eba64a2 // indirect
	google.golang.org/genproto v0.0.0-20230711160842-782d3b101e98 // indirect
	google.golang.org/genproto/googleapis/api v0.0.0-20230711160842-782d3b101e98 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20230711160842-782d3b101e98 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)
