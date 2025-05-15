#!/bin/bash
set -e

# Define color codes for better readability
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo -e "${YELLOW}Starting NRDOT Hello World Processor Demo...${NC}"

# Get the machine's IP address (works on Linux, macOS)
if [[ "$OSTYPE" == "darwin"* ]]; then
  # macOS
  IP=$(ifconfig | grep "inet " | grep -v 127.0.0.1 | awk '{print $2}' | head -n 1)
elif [[ "$OSTYPE" == "linux-gnu"* ]]; then
  # Linux
  IP=$(hostname -I | awk '{print $1}')
else
  # Other
  IP="<your-ip-address>"
fi

echo -e "${GREEN}Detected IP Address: ${BLUE}${IP}${NC}"
echo -e "${GREEN}Starting server...${NC}"
echo -e "${GREEN}You can access the demo at:${NC}"
echo -e "  - ${BLUE}http://localhost:8080${NC} (if on the same machine)"
echo -e "  - ${BLUE}http://${IP}:8080${NC} (from other devices on the same network)"
echo -e ""
echo -e "${YELLOW}Press Ctrl+C to stop the server${NC}"
echo -e ""

# Run the demo
go run main.go