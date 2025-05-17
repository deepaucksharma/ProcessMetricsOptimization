FROM golang:1.22-alpine AS builder

# Set working directory
WORKDIR /app

# Copy go.mod and go.sum
COPY go.mod ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the collector binary
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/bin/otelcol ./cmd/collector

# Verify dependencies
RUN go mod tidy && go mod verify

# Use distroless for the final image
FROM gcr.io/distroless/static-debian12 AS final

# Copy the binary from the builder stage
COPY --from=builder /app/bin/otelcol /usr/local/bin/otelcol

# Use non-root user
USER 10001:10001

EXPOSE 4317 4318 8888 8889 55679

# Run the collector
ENTRYPOINT ["otelcol"]
CMD ["--config=/etc/otel/config/base.yaml"]
