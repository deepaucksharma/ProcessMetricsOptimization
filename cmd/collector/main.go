package main

import (
	"log"

	// Our custom processors
	"github.com/newrelic/nrdot-process-optimization/processors/adaptivetopk"
	"github.com/newrelic/nrdot-process-optimization/processors/helloworld"
	"github.com/newrelic/nrdot-process-optimization/processors/othersrollup"
	"github.com/newrelic/nrdot-process-optimization/processors/prioritytagger"
	"github.com/newrelic/nrdot-process-optimization/processors/reservoirsampler"

	// OTel core
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/exporter"
	"go.opentelemetry.io/collector/extension"
	"go.opentelemetry.io/collector/otelcol"
	"go.opentelemetry.io/collector/processor"
	"go.opentelemetry.io/collector/receiver"

	// Standard processors
	"go.opentelemetry.io/collector/processor/batchprocessor"
	"go.opentelemetry.io/collector/processor/memorylimiterprocessor"

	// Standard exporters
	"go.opentelemetry.io/collector/exporter/otlphttpexporter"

	// Standard extensions
	"go.opentelemetry.io/collector/extension/zpagesextension"

	// Contrib packages for processors, exporters and receivers
	"github.com/open-telemetry/opentelemetry-collector-contrib/exporter/prometheusexporter"
	"github.com/open-telemetry/opentelemetry-collector-contrib/processor/attributesprocessor"
	"github.com/open-telemetry/opentelemetry-collector-contrib/receiver/hostmetricsreceiver"
)

func main() {
	info := component.BuildInfo{
		Command:     "nrdot-process-optimization",
		Description: "New Relic Distribution of OpenTelemetry Process Metrics Optimization",
		Version:     "0.1.0",
	}

	// Create collector settings
	settings := otelcol.CollectorSettings{
		BuildInfo: info,
		Factories: components,
	}

	// Create the collector command
	cmd := otelcol.NewCommand(settings)

	// Execute the command
	if err := cmd.Execute(); err != nil {
		log.Fatal(err)
	}
}

// components returns the set of components used by the collector.
func components() (otelcol.Factories, error) {
	factories := otelcol.Factories{
		Extensions: make(map[component.Type]extension.Factory),
		Receivers:  make(map[component.Type]receiver.Factory),
		Processors: make(map[component.Type]processor.Factory),
		Exporters:  make(map[component.Type]exporter.Factory),
	}

	// Add receivers
	factories.Receivers[hostmetricsreceiver.NewFactory().Type()] = hostmetricsreceiver.NewFactory()

	// Add processors
	factories.Processors[helloworld.NewFactory().Type()] = helloworld.NewFactory()
	factories.Processors[prioritytagger.NewFactory().Type()] = prioritytagger.NewFactory()
	factories.Processors[adaptivetopk.NewFactory().Type()] = adaptivetopk.NewFactory()
	factories.Processors[othersrollup.NewFactory().Type()] = othersrollup.NewFactory()
	factories.Processors[reservoirsampler.NewFactory().Type()] = reservoirsampler.NewFactory()
	factories.Processors[attributesprocessor.NewFactory().Type()] = attributesprocessor.NewFactory()
	factories.Processors[batchprocessor.NewFactory().Type()] = batchprocessor.NewFactory()
	factories.Processors[memorylimiterprocessor.NewFactory().Type()] = memorylimiterprocessor.NewFactory()

	// Add exporters
	factories.Exporters[prometheusexporter.NewFactory().Type()] = prometheusexporter.NewFactory()
	factories.Exporters[otlphttpexporter.NewFactory().Type()] = otlphttpexporter.NewFactory()

	// Add extensions
	factories.Extensions[zpagesextension.NewFactory().Type()] = zpagesextension.NewFactory()

	return factories, nil
}
