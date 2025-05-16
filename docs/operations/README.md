# Operations Documentation

This directory contains documentation related to the operational aspects of the NRDOT Process-Metrics Optimization pipeline.

## Contents

- [Observability Stack Setup](observability_stack_setup.md) - Setting up monitoring infrastructure
- [Dashboard Metrics Audit](dashboard_metrics_audit.md) - Auditing metrics in dashboards
- [Completing Phase 5](completing_phase_5.md) - Final integration steps

## Operational Best Practices

### Deployment Recommendations

- Run the collector close to the metrics source
- Size the collector resources based on process count and metric volume
- Consider using separate collectors for metrics and traces

### Testing Changes

Always test configuration changes:
- Use the test script: `test/test_opt_plus_pipeline.sh`
- Verify cardinality reduction meets goals
- Check critical process preservation
- Monitor performance metrics

### Scaling Considerations

The pipeline scaling is primarily influenced by:
- Total number of processes being monitored
- Number of metrics per process
- Collection frequency

For high-scale environments, consider:
- Multiple collectors with load balancing
- Reducing collection frequency
- Pre-filtering metric sets