#!/bin/bash
set -e

# Define color codes for better readability
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m' # No Color

echo -e "${YELLOW}Starting New Relic OTEL Collector with Trace-Aware attributes...${NC}"

# Load environment variables from .env file if it exists
if [ -f "../.env" ]; then
  echo -e "${GREEN}Loading environment variables from ../.env file${NC}"
  export $(grep -v '^#' "../.env" | xargs)
elif [ -f ".env" ]; then
  echo -e "${GREEN}Loading environment variables from .env file${NC}"
  export $(grep -v '^#' ".env" | xargs)
fi

# Check if NEW_RELIC_LICENSE_KEY is set
if [ -z "$NEW_RELIC_LICENSE_KEY" ]; then
  echo -e "${RED}Error: NEW_RELIC_LICENSE_KEY environment variable is not set.${NC}"
  echo -e "Please set it using: ${YELLOW}export NEW_RELIC_LICENSE_KEY=your_license_key${NC}"
  echo -e "Or create a .env file with NEW_RELIC_LICENSE_KEY=your_license_key"
  exit 1
fi

# Run the Docker container with the official New Relic OTel collector
echo -e "${GREEN}Starting the New Relic OTel collector with trace-aware configuration...${NC}"
docker-compose up -d

# Show logs
echo -e "${GREEN}Collector started! Showing logs:${NC}"
docker-compose logs -f