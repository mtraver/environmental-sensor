# This Dockerfile produces an image that serves the SPA app with nginx.
# IMPORTANT: This file must be kept up-to-date with ../backend.Dockerfile and ../combined.Dockerfile.

FROM node:24-trixie AS build

COPY . /app
WORKDIR /app

# Clean before building.
RUN rm -rf .react-router build node_modules

RUN npm ci
RUN npm run lint
RUN npm run build

FROM nginx:trixie

RUN set -x && apt-get update && DEBIAN_FRONTEND=noninteractive apt-get install -y \
    ca-certificates && \
    rm -rf /var/lib/apt/lists/*

COPY --from=build /app/build/client /app
COPY nginx.conf.template /etc/nginx/templates/default.conf.template

EXPOSE 8080
CMD ["nginx", "-g", "daemon off;"]
