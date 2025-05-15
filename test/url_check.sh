#!/bin/bash
set -e

# Define color codes for better readability
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Function to test a URL
test_url() {
  local url="$1"
  local description="$2"
  local expected_status="$3"
  local timeout="$4"

  echo -e "${YELLOW}Testing ${description} at ${BLUE}${url}${NC}..."
  
  # Use curl to test the URL with a timeout
  status_code=$(curl --write-out '%{http_code}' --silent --output /dev/null --max-time "$timeout" "$url")
  
  if [ "$status_code" = "$expected_status" ]; then
    echo -e "${GREEN}✓ Success: ${description} is accessible (Status code: ${status_code})${NC}"
    return 0
  else
    echo -e "${RED}✗ Failed: ${description} returned status code ${status_code} (Expected: ${expected_status})${NC}"
    return 1
  fi
}

# Main function
main() {
  echo -e "${YELLOW}Starting URL verification tests${NC}"
  echo -e "${BLUE}===============================${NC}"
  
  # Start Docker services if not already running
  echo -e "${YELLOW}Ensuring Docker services are running...${NC}"
  $(dirname "$0")/../run.sh up > /dev/null
  
  # Give services time to initialize
  echo -e "${YELLOW}Waiting for services to initialize (5 seconds)...${NC}"
  sleep 5
  
  local failures=0
  
  # Test services individually to avoid word splitting issues
  test_url "http://localhost:15679/debug/servicez" "zPages" "200" "5" || ((failures++))
  test_url "http://localhost:19090" "Prometheus" "302" "5" || ((failures++))
  test_url "http://localhost:19090/graph" "Prometheus Graph" "302" "5" || ((failures++))
  test_url "http://localhost:13000" "Grafana" "302" "5" || ((failures++))
  test_url "http://localhost:13000/login" "Grafana Login" "200" "5" || ((failures++))
  test_url "http://localhost:18080" "Mock New Relic" "200" "5" || ((failures++))
  
  # Stop Docker services and test demo mode
  echo -e "\n${YELLOW}Testing demo mode...${NC}"
  $(dirname "$0")/../run.sh down > /dev/null
  
  # Start demo in background
  echo -e "${YELLOW}Starting demo mode in background...${NC}"
  (cd $(dirname "$0")/.. && ./run.sh demo > /dev/null 2>&1 &)
  demo_pid=$!
  
  # Give demo time to initialize
  echo -e "${YELLOW}Waiting for demo to initialize (3 seconds)...${NC}"
  sleep 3
  
  # Test demo endpoint
  test_url "http://localhost:8080" "Demo mode" "200" "5" || ((failures++))
  
  # Kill demo process
  echo -e "${YELLOW}Stopping demo mode...${NC}"
  kill $demo_pid 2>/dev/null || true
  
  # Print summary
  echo -e "\n${BLUE}===============================${NC}"
  if [ $failures -eq 0 ]; then
    echo -e "${GREEN}✓ All URL tests passed!${NC}"
    return 0
  else
    echo -e "${RED}✗ ${failures} URL tests failed!${NC}"
    return 1
  fi
}

# Run the main function
main "$@"