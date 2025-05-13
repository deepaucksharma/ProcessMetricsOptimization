package reservoirsampler_with_badger

import (
	"context"

	"github.com/deepaucksharma/reservoir"
	procimpl "github.com/deepaucksharma/trace-aware-reservoir-otel/apps/collector/processor"
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/consumer"
	"go.opentelemetry.io/collector/processor"
	"go.uber.org/zap"
)

const (
	// The type of the processor
	typeStr = "reservoir_sampler"
)

// Wrap the reservoir.DefaultConfig to match the expected function signature
func createDefaultConfig() component.Config {
	return reservoir.DefaultConfig()
}

// NewFactoryWithBadger creates a new processor factory for the reservoir sampler
// with BadgerDB checkpoint manager
func NewFactoryWithBadger() processor.Factory {
	return processor.NewFactory(
		typeStr,
		createDefaultConfig,
		processor.WithTraces(createTracesProcessor, component.StabilityLevelDevelopment),
	)
}

// createTracesProcessor creates a trace processor based on this config
func createTracesProcessor(
	ctx context.Context,
	params processor.CreateSettings,
	cfg component.Config,
	nc consumer.Traces,
) (processor.Traces, error) {
	// Cast to the core config
	coreCfg := cfg.(*reservoir.Config)
	logger := params.Logger

	// Create metrics manager
	metricsManager := procimpl.NewMetricsManager(
		ctx,
	)

	// Register metrics
	if err := metricsManager.RegisterMetrics(); err != nil {
		logger.Error("Failed to register metrics", zap.Error(err))
	}

	// For this simplified version, just return a fake processor for compilation testing
	// In a real implementation, this would create a proper processor with checkpoint manager and adapter
	logger.Info("Creating reservoir sampler processor with reduced functionality for testing",
		zap.String("checkpoint_path", coreCfg.CheckpointPath),
		zap.Bool("trace_aware", coreCfg.TraceAware))
		
	return &procimpl.SimpleReservoirProcessor{
		NextConsumer: nc,
	}, nil
}