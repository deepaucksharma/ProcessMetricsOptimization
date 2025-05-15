FROM otel/opentelemetry-collector:0.125.0

# Use non-root user
USER nobody

# Copy configuration files
COPY config/entity-fixed-config.yaml /etc/otel/config.yaml

# Expose ports
EXPOSE 13133 55678 55679

# Set environment variables
ENV OTEL_EXPORTER_OTLP_METRICS_TEMPORALITY_PREFERENCE=delta
ENV OTEL_DEPLOYMENT_ENVIRONMENT=production
ENV OTEL_SERVICE_NAME=otel-collector-host

# Use a health check
HEALTHCHECK --interval=30s --timeout=5s --start-period=5s --retries=3 \
  CMD wget --no-verbose --tries=1 --spider http://localhost:13133/ || exit 1

# Start the collector
CMD ["--config=/etc/otel/config.yaml"]