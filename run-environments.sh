#!/bin/bash

# ANSI color codes
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[0;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Function to check if a command exists
command_exists() {
    command -v "$1" >/dev/null 2>&1
}

# Check for required tools
if ! command_exists docker; then
    echo -e "${RED}Docker is not installed. Please install Docker to continue.${NC}"
    exit 1
fi

# Change to the directory containing this script
cd "$(dirname "$0")"

# Set the New Relic license key - Replace with your actual key
export NEW_RELIC_LICENSE_KEY=2a326cab47d49f5aab8db7ee009e4a57FFFFNRAL

# Prompt user for proper license key if it contains placeholder
if [[ "$NEW_RELIC_LICENSE_KEY" == *"FFFFNRAL"* ]]; then
    echo -e "${YELLOW}WARNING: You're using a placeholder New Relic license key.${NC}"
    read -p "Enter your actual New Relic license key (or press Enter to continue with placeholder): " USER_KEY
    if [ ! -z "$USER_KEY" ]; then
        export NEW_RELIC_LICENSE_KEY="$USER_KEY"
        echo "Using provided license key."
    else
        echo -e "${YELLOW}Continuing with placeholder key. Data may not appear in New Relic.${NC}"
    fi
fi

# Function to run Docker environment
run_docker() {
    echo -e "${BLUE}=== Starting Docker Environment ===${NC}"
    
    # Stop any existing containers
    docker-compose -f docker-compose-entity-fixed.yml down 2>/dev/null
    
    # Start the container
    docker-compose -f docker-compose-entity-fixed.yml up -d
    
    # Check if container started successfully
    if [ $? -eq 0 ]; then
        echo -e "${GREEN}Docker container started successfully.${NC}"
        echo "To view logs: docker logs entity-fixed-collector -f"
        echo "To stop: docker-compose -f docker-compose-entity-fixed.yml down"
    else
        echo -e "${RED}Failed to start Docker container. Check docker-compose-entity-fixed.yml.${NC}"
        exit 1
    fi
}

# Function to run Mac environment
run_mac() {
    echo -e "${BLUE}=== Starting Mac Environment ===${NC}"
    
    # Kill any existing otelcol processes
    pkill -f otelcol > /dev/null 2>&1
    
    # Check if we have the otelcol binary
    if [ ! -f "./otelcol" ]; then
        echo -e "${YELLOW}OpenTelemetry Collector binary not found. Downloading...${NC}"
        
        # Check OS architecture
        ARCH=$(uname -m)
        if [ "$ARCH" == "arm64" ]; then
            OTEL_BINARY_URL="https://github.com/open-telemetry/opentelemetry-collector-releases/releases/download/v0.125.0/otelcol_0.125.0_darwin_arm64.tar.gz"
        else
            OTEL_BINARY_URL="https://github.com/open-telemetry/opentelemetry-collector-releases/releases/download/v0.125.0/otelcol_0.125.0_darwin_amd64.tar.gz"
        fi
        
        # Download and extract the binary
        curl -sL $OTEL_BINARY_URL -o otelcol.tar.gz
        tar -xzf otelcol.tar.gz otelcol
        chmod +x otelcol
        rm otelcol.tar.gz
        
        if [ ! -f "./otelcol" ]; then
            echo -e "${RED}Failed to download OpenTelemetry Collector binary.${NC}"
            exit 1
        fi
    fi
    
    # Run the entity fixed script
    ./run-entity-fixed.sh
}

# Main menu
echo -e "${BLUE}====== OpenTelemetry Collector Runner ======${NC}"
echo "This script runs the entity-fixed OpenTelemetry Collector in different environments."
echo
echo "Choose an environment to run:"
echo "1) Docker"
echo "2) Mac (local)"
echo "3) Both"
echo "4) Exit"
echo

read -p "Enter your choice (1-4): " CHOICE

case $CHOICE in
    1)
        run_docker
        ;;
    2)
        run_mac
        ;;
    3)
        run_docker
        echo
        run_mac
        ;;
    4)
        echo "Exiting."
        exit 0
        ;;
    *)
        echo -e "${RED}Invalid choice. Exiting.${NC}"
        exit 1
        ;;
esac

echo
echo -e "${GREEN}====== Setup Complete ======${NC}"
echo -e "${YELLOW}Important NRQL Queries for Testing:${NC}"
echo
echo "Check for entity synthesis:"
echo "FROM Metric SELECT count(*) WHERE instrumentation.provider = 'opentelemetry' FACET host.id, host.name, entity.type SINCE 30 minutes ago"
echo
echo "Check for trace-aware attributes:"
echo "FROM Metric SELECT count(*) WHERE trace.aware.collector = 'true' FACET collector.name SINCE 30 minutes ago"
echo
echo "Check process metrics:"
echo "FROM Metric SELECT latest(process.cpu.pct) FACET host.name, process.executable.name LIMIT 100 SINCE 30 minutes ago"
echo
echo "Remember to check the New Relic UI Entities explorer to see if your host appears after a few minutes."
echo "For troubleshooting, refer to ENTITY_SYNTHESIS.md"