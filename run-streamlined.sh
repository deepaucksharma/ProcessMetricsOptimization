#!/bin/bash

# Color codes
GREEN='\033[0;32m'
YELLOW='\033[0;33m'
BLUE='\033[0;34m'
RED='\033[0;31m'
NC='\033[0m' # No Color

# Initialize host ID
if [ -f /etc/machine-id ]; then
  export HOST_ID=$(cat /etc/machine-id)
elif [ -f /var/lib/dbus/machine-id ]; then
  export HOST_ID=$(cat /var/lib/dbus/machine-id)
else
  export HOST_ID=$(hostname | md5 | cut -d' ' -f1)
fi

# Set up environment
export NEW_RELIC_LICENSE_KEY=${NEW_RELIC_LICENSE_KEY:-2a326cab47d49f5aab8db7ee009e4a57FFFFNRAL}
export OTEL_DEPLOYMENT_ENVIRONMENT=${OTEL_DEPLOYMENT_ENVIRONMENT:-production}
export OTEL_SERVICE_NAME=${OTEL_SERVICE_NAME:-otel-collector-host}
export OTEL_RESOURCE_ATTRIBUTES="host.name=$(hostname),entity.type=HOST"
export OTEL_EXPORTER_OTLP_METRICS_TEMPORALITY_PREFERENCE=delta

# Base directory
cd "$(dirname "$0")"
CONFIG_FILE="./config/streamlined-config.yaml"

# Validate license key
if [[ "$NEW_RELIC_LICENSE_KEY" == *"FFFFNRAL"* ]]; then
  echo -e "${YELLOW}WARNING: Using placeholder license key. Data won't appear in New Relic.${NC}"
  read -p "Enter your actual New Relic license key (or press Enter to continue): " USER_KEY
  if [[ -n "$USER_KEY" ]]; then
    export NEW_RELIC_LICENSE_KEY="$USER_KEY"
  fi
fi

# Function to run in Docker
run_docker() {
  echo -e "${BLUE}Starting collector in Docker...${NC}"
  docker-compose -f docker-compose-streamlined.yml down 2>/dev/null
  docker-compose -f docker-compose-streamlined.yml up -d
  if [ $? -eq 0 ]; then
    echo -e "${GREEN}Docker container started successfully${NC}"
    echo "View logs: docker logs otel-collector"
  else
    echo -e "${RED}Failed to start Docker container${NC}"
    return 1
  fi
}

# Function to run locally
run_local() {
  echo -e "${BLUE}Starting collector locally...${NC}"
  pkill -f otelcol > /dev/null 2>&1
  
  # Check for binary
  if [ ! -f "./otelcol" ]; then
    echo -e "${YELLOW}Downloading OpenTelemetry Collector...${NC}"
    ARCH=$(uname -m)
    OS=$(uname -s | tr '[:upper:]' '[:lower:]')
    if [ "$ARCH" == "x86_64" ]; then ARCH="amd64"; fi
    if [ "$ARCH" == "aarch64" ]; then ARCH="arm64"; fi
    
    curl -sSL "https://github.com/open-telemetry/opentelemetry-collector-releases/releases/download/v0.125.0/otelcol_0.125.0_${OS}_${ARCH}.tar.gz" | tar xz otelcol
    chmod +x otelcol
  fi
  
  # Launch collector
  ./otelcol --config="$CONFIG_FILE" > otel-collector.log 2>&1 &
  PID=$!
  
  # Verify it's running
  sleep 2
  if kill -0 $PID 2>/dev/null; then
    echo -e "${GREEN}Collector started successfully with PID $PID${NC}"
    echo "View logs: tail -f otel-collector.log"
    # Auto-shutdown after 2 hours
    (
      sleep 7200
      if kill -0 $PID 2>/dev/null; then
        echo "Auto-stopping collector after 2 hours" >> otel-collector.log
        kill $PID
      fi
    ) &
  else
    echo -e "${RED}Collector failed to start - check otel-collector.log${NC}"
    return 1
  fi
}

# Menu function
show_menu() {
  echo -e "${BLUE}========== OpenTelemetry Collector ===========${NC}"
  echo "Choose run environment:"
  echo "1) Docker container"
  echo "2) Local system"
  echo "3) Both Docker and local"
  echo "4) Quit"
  
  # Auto-select option if environment variable is set
  if [ -n "$AUTO_OPTION" ]; then
    CHOICE=$AUTO_OPTION
    echo "Auto-selected option: $CHOICE"
  else
    read -p "Select option [1-4]: " CHOICE
  fi
  
  case $CHOICE in
    1) run_docker ;;
    2) run_local ;;
    3) run_docker && run_local ;;
    4) echo "Exiting"; exit 0 ;;
    *) echo -e "${RED}Invalid option${NC}"; exit 1 ;;
  esac
  
  if [ $? -eq 0 ]; then
    echo -e "\n${GREEN}Collector is now running${NC}"
    echo -e "${YELLOW}NRQL example: FROM Metric SELECT count(*) WHERE collector.name='trace-aware-collector'${NC}"
  fi
}

# Main logic
show_menu