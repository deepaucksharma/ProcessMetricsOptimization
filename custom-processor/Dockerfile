FROM golang:1.22-alpine AS builder

# Install necessary build tools
RUN apk add --no-cache git make

# Install the OpenTelemetry Collector Builder
RUN go install go.opentelemetry.io/collector/cmd/builder@v0.80.0

# Set up working directory
WORKDIR /build

# Copy our custom processor code and builder configuration
COPY processor/ ./processor/
COPY builder/ ./builder/

# Build the custom collector with our processor
WORKDIR /build/builder
RUN builder --config builder-config.yaml

# Create a minimal production image
FROM alpine:latest

# Add dependencies needed by the collector
RUN apk add --no-cache ca-certificates

# Create a non-root user to run the collector
RUN addgroup -S otel && adduser -S otel -G otel

# Setup work directory and permissions
WORKDIR /etc/otel
RUN chown -R otel:otel /etc/otel

# Copy the collector binary from the builder stage
COPY --from=builder /build/builder/bin/traceaware-otelcol /usr/local/bin/

# Use non-root user for security
USER otel

# Define entrypoint
ENTRYPOINT ["/usr/local/bin/traceaware-otelcol"]
CMD ["--config", "/etc/otel/config.yaml"]