#!/bin/bash
set -e

# Define color codes for better readability
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m' # No Color

echo -e "${YELLOW}Starting Hello World OpenTelemetry Processor...${NC}"

# Load environment variables from .env file
if [ -f ".env" ]; then
  echo -e "${GREEN}Loading environment variables from .env file${NC}"
  export $(grep -v '^#' ".env" | xargs)
fi

# Check if NEW_RELIC_LICENSE_KEY is set
if [ -z "$NEW_RELIC_LICENSE_KEY" ]; then
  echo -e "${RED}Error: NEW_RELIC_LICENSE_KEY environment variable is not set.${NC}"
  echo -e "Please check your .env file."
  exit 1
fi

echo -e "${GREEN}Success! Environment configured correctly.${NC}"
echo -e "${GREEN}Ready to run with New Relic License Key: ${NC}${NEW_RELIC_LICENSE_KEY}"
echo -e "${GREEN}Environment: ${NC}${OTEL_DEPLOYMENT_ENVIRONMENT}"
echo -e "${GREEN}Log Level: ${NC}${OTEL_LOG_LEVEL}"