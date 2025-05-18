package main

import (
	"log"

	"github.com/newrelic/nrdot-process-optimization/processors/helloworld"
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/exporter"
	"go.opentelemetry.io/collector/extension"
	"go.opentelemetry.io/collector/processor"
	"go.opentelemetry.io/collector/receiver"
	"go.opentelemetry.io/collector/service"
	
	"go.opentelemetry.io/collector/exporter/otlpexporter"
	"go.opentelemetry.io/collector/exporter/otlphttpexporter"
	"go.opentelemetry.io/collector/exporter/prometheusexporter"
	"go.opentelemetry.io/collector/extension/zpagesextension"
	"go.opentelemetry.io/collector/processor/batchprocessor"
	"go.opentelemetry.io/collector/processor/memorylimiterprocessor"
	"go.opentelemetry.io/collector/receiver/hostmetricsreceiver"
	"go.opentelemetry.io/collector/receiver/otlpreceiver"
)

func main() {
	factories, err := components()
	if err != nil {
		log.Fatalf("Failed to build components: %v", err)
	}

	info := component.BuildInfo{
		Command:     "nrdot-process-optimization",
		Description: "New Relic Distribution of OpenTelemetry Process Metrics Optimization",
		Version:     "0.1.0",
	}

	if err = service.Run(service.CollectorSettings{
		BuildInfo:     info,
		Factories:     factories,
		ConfigSources: service.ConfigSources{},
	}); err != nil {
		log.Fatal(err)
	}
}

func components() (component.Factories, error) {
	factories := component.Factories{
		Extensions: make(map[component.Type]extension.Factory),
		Receivers:  make(map[component.Type]receiver.Factory),
		Processors: make(map[component.Type]processor.Factory),
		Exporters:  make(map[component.Type]exporter.Factory),
	}

	// Register extensions
	extensions := []extension.Factory{
		zpagesextension.NewFactory(),
	}
	for _, ext := range extensions {
		factories.Extensions[ext.Type()] = ext
	}

	// Register receivers
	receivers := []receiver.Factory{
		hostmetricsreceiver.NewFactory(),
		otlpreceiver.NewFactory(),
	}
	for _, rcv := range receivers {
		factories.Receivers[rcv.Type()] = rcv
	}

	// Register processors
	processors := []processor.Factory{
		batchprocessor.NewFactory(),
		memorylimiterprocessor.NewFactory(),
		helloworld.NewFactory(),
	}
	for _, proc := range processors {
		factories.Processors[proc.Type()] = proc
	}

	// Register exporters
	exporters := []exporter.Factory{
		otlpexporter.NewFactory(),
		otlphttpexporter.NewFactory(),
		prometheusexporter.NewFactory(),
	}
	for _, exp := range exporters {
		factories.Exporters[exp.Type()] = exp
	}

	return factories, nil
}
