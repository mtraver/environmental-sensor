# This Dockerfile produces an image that runs a gRPC server implementing MeasurementService.
FROM golang:1.12 as builder

RUN mkdir /build
WORKDIR /build
COPY cmd/api api/
COPY measurement measurement/
COPY measurementpb measurementpb/
COPY web web/
COPY go.mod .
COPY go.sum .

RUN CGO_ENABLED=0 GOOS=linux go build -v -o serve api/main.go

FROM alpine
RUN apk add --no-cache ca-certificates

COPY --from=builder /build/serve /serve

ENTRYPOINT ["/serve"]
