#!/bin/bash
set -e

# Define color codes for better readability
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m' # No Color

echo -e "${YELLOW}Starting Hello World OpenTelemetry Processor...${NC}"

# Run the Docker container
echo -e "${GREEN}Starting the OpenTelemetry Collector with Hello World processor...${NC}"
docker-compose up -d

# Show logs
echo -e "${GREEN}Collector started! Showing logs:${NC}"
docker-compose logs -f