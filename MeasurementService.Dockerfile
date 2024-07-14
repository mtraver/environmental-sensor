# This Dockerfile produces an image that runs a gRPC server implementing MeasurementService.
FROM golang:1.22-bullseye as builder

WORKDIR /build

# Retrieve application dependencies.
# This allows the container build to reuse cached dependencies.
COPY go.* ./
RUN go mod download

COPY aqi aqi/
COPY cache cache/
COPY cmd/api api/
COPY federatedidentity federatedidentity/
COPY measurement measurement/
COPY measurementpb measurementpb/
COPY web web/

RUN CGO_ENABLED=0 GOOS=linux go build -v -o serve api/main.go

FROM debian:bullseye-slim
RUN set -x && apt-get update && DEBIAN_FRONTEND=noninteractive apt-get install -y \
    ca-certificates && \
    rm -rf /var/lib/apt/lists/*

COPY --from=builder /build/serve /serve

ENTRYPOINT ["/serve"]
