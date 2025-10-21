# Multi-stage build for obs-mcp server
FROM registry.access.redhat.com/ubi9/go-toolset:latest AS build

# Set working directory
WORKDIR /usr/src/app

# Copy go mod files first for better layer caching
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o obs-mcp ./cmd/obs-mcp

# Production stage
FROM registry.access.redhat.com/ubi9/ubi-minimal:latest

# Install ca-certificates for HTTPS requests
RUN microdnf update -y && \
    microdnf install -y ca-certificates && \
    microdnf clean all

# Create non-root user
RUN useradd -r -u 1001 -g root -s /bin/false obs-mcp

# Set working directory
WORKDIR /app

# Copy binary from build stage
COPY --from=build /usr/src/app/obs-mcp /app/obs-mcp

USER 1001

# Expose default port (can be overridden via LISTEN_ADDRESS env var)
EXPOSE 9100

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:${LISTEN_ADDRESS:-:9100}/health || exit 1

# Default environment variables
ENV LISTEN_ADDRESS="0.0.0.0:9100"
ENV PROMETHEUS_URL="http://thanos-querier.openshift-monitoring.svc.cluster.local:9091"

# Start the server
CMD ["/app/obs-mcp", "--listen", "0.0.0.0:9100"]

