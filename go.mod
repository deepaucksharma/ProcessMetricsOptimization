module github.com/newrelic/nrdot-process-optimization

go 1.22

require (
	github.com/open-telemetry/opentelemetry-collector-contrib/processor/metricstransformprocessor v0.94.0
	go.opentelemetry.io/collector/component v0.94.0
	go.opentelemetry.io/collector/confmap v0.94.0
	go.opentelemetry.io/collector/consumer v0.94.0
	go.opentelemetry.io/collector/exporter v0.94.0
	go.opentelemetry.io/collector/extension v0.94.0
	go.opentelemetry.io/collector/pdata v1.2.0
	go.opentelemetry.io/collector/processor v0.94.0
	go.opentelemetry.io/collector/receiver v0.94.0
	go.opentelemetry.io/collector/service v0.94.0
	go.opentelemetry.io/otel v1.21.0
	go.opentelemetry.io/otel/metric v1.21.0
	go.opentelemetry.io/otel/trace v1.21.0
	go.uber.org/multierr v1.11.0
	go.uber.org/zap v1.26.0
)

require (
	github.com/cenkalti/backoff/v4 v4.2.1
	github.com/gogo/protobuf v1.3.2
	github.com/golang/protobuf v1.5.3
	github.com/knadh/koanf/v2 v2.0.1
	github.com/stretchr/testify v1.8.4
	google.golang.org/grpc v1.59.0
	google.golang.org/protobuf v1.31.0
)