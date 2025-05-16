# Dashboard Documentation

This directory contains documentation related to the Grafana dashboards for monitoring the NRDOT Process-Metrics Optimization pipeline.

## Contents

- [Grafana Dashboard Design](grafana_dashboard_design.md) - Dashboard design principles and best practices
- [Dashboard Overview](dashboard_overview.md) - Overview of available dashboards

## Available Dashboards

The project includes several Grafana dashboards for comprehensive monitoring:

1. **Full Pipeline Overview Dashboard**
   - Provides a holistic view of the entire optimization pipeline
   - Shows collector health, processor performance, and optimization impact
   - Displays cardinality reduction metrics

2. **Processor-Specific Dashboards**
   - PriorityTagger Dashboard
   - AdaptiveTopK Dashboard
   - OthersRollup Dashboard
   - ReservoirSampler Dashboard

3. **Algorithm Insight Dashboards**
   - Visualize the internal workings of each processor algorithm
   - Show decision-making processes and effectiveness

## Dashboard Usage

The dashboards are available in the `/dashboards` directory as JSON files that can be imported into Grafana. See the [Dashboard Overview](dashboard_overview.md) for detailed usage instructions.