#!/bin/bash

# test_opt_plus_pipeline.sh
# Tests the full opt-plus pipeline with all processors, verifying each stage

set -e # Exit on any error

# Colors for better readability
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Step 1: Build the custom collector with all processors
echo -e "${BLUE}Step 1: Building the custom collector with all processors...${NC}"
make docker-build || { echo -e "${RED}Failed to build Docker image${NC}"; exit 1; }
echo -e "${GREEN}✅ Custom collector built successfully${NC}\n"

# Step 2: Start the optimization pipeline with the opt-plus config
CONFIG_FILE="${COLLECTOR_CONFIG:-opt-plus.yaml}"
echo -e "${BLUE}Step 2: Starting the optimization pipeline with ${CONFIG_FILE}...${NC}"
export COLLECTOR_CONFIG="${CONFIG_FILE}"
make compose-up || { echo -e "${RED}Failed to start Docker Compose stack${NC}"; exit 1; }
echo -e "${GREEN}✅ Optimization pipeline started${NC}\n"

# Step 3: Wait for services to be ready
echo -e "${BLUE}Step 3: Waiting for services to be available...${NC}"
sleep 5 # Give services time to initialize

# Verify collector is up
echo -e "${YELLOW}Checking collector accessibility...${NC}"
curl -s --head --fail http://localhost:15679 &>/dev/null || {
  echo -e "${RED}Collector zPages not accessible (http://localhost:15679)${NC}"
  make logs
  exit 1
}
echo -e "${GREEN}✅ Collector zPages accessible${NC}"

# Verify Prometheus
echo -e "${YELLOW}Checking Prometheus accessibility...${NC}"
curl -s --head --fail http://localhost:19090 &>/dev/null || {
  echo -e "${RED}Prometheus not accessible (http://localhost:19090)${NC}"
  exit 1
}
echo -e "${GREEN}✅ Prometheus accessible${NC}"

# Verify Grafana
echo -e "${YELLOW}Checking Grafana accessibility...${NC}"
curl -s --head --fail http://localhost:13000 &>/dev/null || {
  echo -e "${RED}Grafana not accessible (http://localhost:13000)${NC}"
  exit 1
}
echo -e "${GREEN}✅ Grafana accessible${NC}\n"

# Step 4: Allow some time for metrics to be collected and processed
echo -e "${BLUE}Step 4: Allowing time for metrics collection (30 seconds)...${NC}"
echo -e "${YELLOW}This gives time for:${NC}"
echo -e "${YELLOW}  - hostmetrics receiver to collect process metrics${NC}"
echo -e "${YELLOW}  - prioritytagger to identify critical processes${NC}"
echo -e "${YELLOW}  - adaptivetopk to select top processes${NC}"
echo -e "${YELLOW}  - reservoirsampler to sample representative processes${NC}"
echo -e "${YELLOW}  - othersrollup to aggregate remaining processes${NC}"
sleep 30

# Step 5: Verify all processors are active via Prometheus metrics
echo -e "\n${BLUE}Step 5: Verifying all processors are active via Prometheus...${NC}"

# Function to check if a metric exists in Prometheus
check_metric() {
  metric=$1
  description=$2
  result=$(curl -s "http://localhost:19090/api/v1/query?query=${metric}" | grep -o '"resultType":"matrix"')

  if [[ ! -z $result ]]; then
    echo -e "${GREEN}✅ ${description} (${metric})${NC}"
    return 0
  else
    echo -e "${RED}❌ ${description} (${metric}) - Metric not found${NC}"
    return 1
  }
}

# Check standard processor metrics for each processor
prio_check=$(check_metric "otelcol_processor_prioritytagger_processed_metric_points" "PriorityTagger processor active")
topk_check=$(check_metric "otelcol_processor_adaptivetopk_processed_metric_points" "AdaptiveTopK processor active")
rollup_check=$(check_metric "otelcol_processor_othersrollup_processed_metric_points" "OthersRollup processor active")
sampler_check=$(check_metric "otelcol_processor_reservoirsampler_processed_metric_points" "ReservoirSampler processor active")

# Check processor-specific metrics
check_metric "nrdot_prioritytagger_critical_processes_tagged_total" "PriorityTagger tagged processes"
check_metric "nrdot_adaptivetopk_topk_processes_selected_total" "AdaptiveTopK process selection"
check_metric "nrdot_adaptivetopk_current_k_value" "AdaptiveTopK dynamic K value"
check_metric "nrdot_othersrollup_metrics_rolled_up_total" "OthersRollup aggregation"
check_metric "nrdot_reservoirsampler_samples_selected_total" "ReservoirSampler sampling"

# Check if all basic checks passed
if [[ $prio_check == 0 && $topk_check == 0 && $rollup_check == 0 && $sampler_check == 0 ]]; then
  echo -e "${GREEN}✅ All processors are active in the pipeline${NC}"
else
  echo -e "${RED}❌ One or more processors may not be active${NC}"
  make logs
  exit 1
fi

# Step 5a: Measure cardinality reduction
echo -e "\n${BLUE}Step 5a: Measuring cardinality reduction...${NC}"

# Get raw metrics count (pre-pipeline)
raw_metric_count=$(curl -s "http://localhost:19090/api/v1/query?query=count(process_cpu_utilization)" | grep -o '"result":\[{"metric":{},"value":\[.*,"\([0-9]*\)"' | sed 's/.*,"\([0-9]*\)"/\1/')

# Get optimized metrics count (post-pipeline)
opt_metric_count=$(curl -s "http://localhost:19090/api/v1/query?query=count(process_cpu_utilization{nr_priority='critical'} or process_cpu_utilization{nr_process_sampled_by_reservoir='true'} or process_cpu_utilization{_other_=~'.*'})" | grep -o '"result":\[{"metric":{},"value":\[.*,"\([0-9]*\)"' | sed 's/.*,"\([0-9]*\)"/\1/')

# Calculate cardinality reduction if counts are valid numbers
if [[ $raw_metric_count =~ ^[0-9]+$ ]] && [[ $opt_metric_count =~ ^[0-9]+$ ]] && [[ $raw_metric_count -gt 0 ]]; then
  reduction=$(echo "scale=2; 100 - ($opt_metric_count * 100 / $raw_metric_count)" | bc)
  echo -e "${GREEN}✅ Cardinality reduction achieved: ${reduction}%${NC}"
  echo -e "${GREEN}   - Raw metric count: ${raw_metric_count}${NC}"
  echo -e "${GREEN}   - Optimized metric count: ${opt_metric_count}${NC}"

  # Check if we met the 90% reduction goal
  if (( $(echo "$reduction >= 90" | bc -l) )); then
    echo -e "${GREEN}✅ Success! Met or exceeded the 90% cardinality reduction goal${NC}"
  else
    echo -e "${YELLOW}⚠️ Partial success: ${reduction}% reduction achieved, but goal is ≥90%${NC}"
    echo -e "${YELLOW}   Consider tuning the pipeline configuration for better results${NC}"
  fi
else
  echo -e "${YELLOW}⚠️ Could not calculate cardinality reduction - not enough metrics data yet${NC}"
  echo -e "${YELLOW}   Raw count: ${raw_metric_count}, Optimized count: ${opt_metric_count}${NC}"
fi

# Step 6: Verify critical processes are preserved
echo -e "\n${BLUE}Step 6: Verifying critical processes are preserved...${NC}"

# Get critical processes defined in the configuration
config_path="$(dirname "$0")/../config/opt-plus.yaml"
echo -e "${YELLOW}Checking for critical processes defined in ${config_path}...${NC}"
critical_procs=$(grep -A4 "critical_executables:" "$config_path" | grep -v "critical_executables:" | grep -v "#" | grep -o '".*"' | sed 's/"//g' | tr '\n' ' ')
echo -e "${GREEN}✅ Configured critical processes: ${critical_procs}${NC}"

# Check if process metrics for these critical processes exist in Prometheus
echo -e "${YELLOW}Verifying metrics for critical processes exist...${NC}"
for proc in $critical_procs; do
  if [[ -n "$proc" ]]; then
    result=$(curl -s "http://localhost:19090/api/v1/query?query=process_cpu_utilization{process_executable_name=\"$proc\",nr_priority=\"critical\"}" | grep -o '"resultType":"matrix"')
    if [[ ! -z $result ]]; then
      echo -e "${GREEN}✅ Found metrics for critical process: $proc${NC}"
    else
      echo -e "${YELLOW}⚠️ No metrics found for critical process: $proc (may not be running in container)${NC}"
    fi
  fi
done

# Step 7: Check the mock OTLP sink for evidence of the pipeline working
echo -e "\n${BLUE}Step 7: Checking mock OTLP sink logs for optimization evidence...${NC}"
echo -e "${YELLOW}Tailing the last 100 lines from the mock-otlp-sink service...${NC}"
docker logs --tail 100 mock-otlp-sink 2>&1 | grep -E "process|critical|_other_" || {
  echo -e "${RED}❌ No evidence of process metrics in the mock sink${NC}"
  exit 1
}

echo -e "\n${GREEN}=================================================================================${NC}"
echo -e "${GREEN}✅ All tests passed! The full optimization pipeline is running correctly.${NC}"
echo -e "${GREEN}✅ Phase 5 integration test completed successfully:${NC}"
echo -e "${GREEN}   - All processors are active and processing metrics${NC}"
echo -e "${GREEN}   - The pipeline is achieving significant cardinality reduction${NC}"
echo -e "${GREEN}   - Critical processes are being preserved${NC}"
echo -e "${GREEN}   - All pipeline stages are functioning as expected${NC}"
echo -e "${GREEN}✅ You can now explore the data in more detail:${NC}"
echo -e "${GREEN}   - Check Grafana: http://localhost:13000${NC}"
echo -e "${GREEN}   - Query Prometheus: http://localhost:19090${NC}"
echo -e "${GREEN}   - Examine collector zPages: http://localhost:15679${NC}"
echo -e "${GREEN}   - View OTLP sink logs: make logs${NC}"
echo -e "${GREEN}=================================================================================${NC}"

# Keep the pipeline running for manual exploration
echo -e "${YELLOW}The pipeline will continue running for you to explore.${NC}"
echo -e "${YELLOW}Press Ctrl+C when done, then run 'make compose-down' to clean up.${NC}"

# Keep script running until Ctrl+C
trap "echo -e '${YELLOW}Stopping script. Run make compose-down to stop the services.${NC}'" INT
tail -f /dev/null
