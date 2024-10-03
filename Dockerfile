# Dockerfile using multistage build to create a small image.
# Build with golang:1.23-alpine as it contains the golang compiler and its dependencies.
# These take up disk space, and we don't need them in the final image.
FROM golang:1.23-alpine AS build
WORKDIR /go/src/wal
COPY . .
RUN CGO_ENABLED=0 go build -o /go/bin/wal ./cmd/wal

# Kubernetes has a built-in health check endpoint for gRPC services nowdays,
# therefore we don't need to use this tool and can use the native Kubernetes health check.
# ORIGINAL RUN GRPC_HEALTH_PROBE_VERSION=v0.3.2 && \
# LATEST   RUN GRPC_HEALTH_PROBE_VERSION=v0.4.34 && \
#     wget -qO/go/bin/grpc_health_probe \
#     https:#github.com/grpc-ecosystem/grpc-health-probe/releases/download/\
#     ${GRPC_HEALTH_PROBE_VERSION}/grpc_health_probe-linux-amd64 && \
#     chmod +x /go/bin/grpc_health_probe
# END_HIGHLIGHT
# START: beginning

# Use scratch as the base image.
# This is a minimal image that contains only the runtime environment.
FROM scratch
COPY --from=build /go/bin/wal /bin/wal
# COPY --from=build /go/bin/grpc_health_probe /bin/grpc_health_probe
ENTRYPOINT ["/bin/wal"]