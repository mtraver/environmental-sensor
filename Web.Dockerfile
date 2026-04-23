# This Dockerfile builds an image that serves the web backend and frontend together.
# The frontend is served by nginx on port 8080 and the backend binary is simply run.

####################
# Build Go backend #
####################
FROM golang:1.25-trixie AS backend-builder

WORKDIR /build

# Retrieve application dependencies.
# This allows the container build to reuse cached dependencies.
COPY go.* ./
RUN go mod download

# Copy in code.
COPY aqi aqi/
COPY cache cache/
COPY database database/
COPY federatedidentity federatedidentity/
COPY graph/ graph
COPY measurement measurement/
COPY measurementpb measurementpb/
COPY measurementpbutil measurementpbutil/
COPY web web/

RUN mkdir out
RUN CGO_ENABLED=0 GOOS=linux go build -v -o out ./web/...

######################
# Build React client #
######################
FROM node:24-trixie AS client-builder

COPY client /app
WORKDIR /app

# Clean before building.
RUN rm -rf .react-router build node_modules

RUN npm ci
RUN npm run lint
RUN npm run build

#######
# Run #
#######
FROM nginx:trixie

RUN set -x && apt-get update && DEBIAN_FRONTEND=noninteractive apt-get install -y \
    ca-certificates tini && \
    rm -rf /var/lib/apt/lists/*

COPY --from=client-builder /app/build/client /app
COPY --from=backend-builder /build/out/web /serve

# Copy in resources required at runtime.
COPY --from=backend-builder /build/web/templates /web/templates

COPY nginx.conf.template /etc/nginx/templates/default.conf.template

COPY start.sh /start.sh
RUN chmod +x /start.sh

# Use tini for its signal forwarding and zombie reaping functionality.
ENTRYPOINT ["/usr/bin/tini", "--"]

EXPOSE 8080
ENV BACKEND_HOST http://127.0.0.1
ENV BACKEND_PORT 3001
CMD ["/start.sh"]
