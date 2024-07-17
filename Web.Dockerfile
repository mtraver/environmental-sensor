# This Dockerfile produces an image that runs the web server.
FROM golang:1.22-bookworm as builder

WORKDIR /build

# Retrieve application dependencies.
# This allows the container build to reuse cached dependencies.
COPY go.* ./
RUN go mod download

COPY aqi aqi/
COPY cache cache/
COPY federatedidentity federatedidentity/
COPY measurement measurement/
COPY measurementpb measurementpb/
COPY measurementpbutil measurementpbutil/
COPY web web/

RUN mkdir out
RUN CGO_ENABLED=0 GOOS=linux go build -v -o out ./web/...

FROM debian:bookworm-slim
RUN set -x && apt-get update && DEBIAN_FRONTEND=noninteractive apt-get install -y \
    ca-certificates && \
    rm -rf /var/lib/apt/lists/*

# Copy in binary from the builder stage.
COPY --from=builder /build/out/web /serve

# Copy in resources required at runtime.
COPY --from=builder /build/web/templates /web/templates
COPY --from=builder /build/web/static /web/static

ENV SERVE_STATIC=1

ENTRYPOINT ["/serve"]
