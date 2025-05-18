#!/bin/bash
set -e

# Define color codes for better readability
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Function to display usage
display_usage() {
  echo -e "${YELLOW}NRDOT Process-Metrics Optimization - Runner${NC}"
  echo
  echo -e "Usage: ./run.sh [command] [options]"
  echo
  echo -e "Commands:"
  echo -e "  up        Start all services"
  echo -e "  down      Stop all services"
  echo -e "  logs      Follow logs from all services"
  echo -e "  restart   Restart all services"
  echo
  echo -e "Options:"
  echo -e "  --no-browser   Don't automatically open URLs in the browser"
  echo -e "  --open-urls    Open URLs in the browser (default)"
  echo
  echo -e "Environment Variables:"
  echo -e "  COLLECTOR_CONFIG  Collector config file (default: opt-plus.yaml)"
  echo
  echo -e "Examples:"
  echo -e "  ./run.sh up                     Start services with Docker"
  echo -e "  ./run.sh up --no-browser        Start services without opening URLs"
  echo
  echo -e "Access Points:"
  echo -e "  zPages: http://localhost:15679"
  echo -e "  Prometheus: http://localhost:19090"
  echo -e "  Grafana: http://localhost:13000 (admin/admin)"
  echo -e "  Mock New Relic: http://localhost:18080"
  echo
}

# Check Docker is running
check_docker() {
  if ! docker info >/dev/null 2>&1; then
    echo -e "${RED}Error: Docker is not running or not installed.${NC}"
    echo -e "Please start Docker and try again."
    exit 1
  fi
}

# Load environment variables from .env file
load_env() {
  if [ -f ".env" ]; then
    echo -e "${GREEN}Loading environment variables from .env file${NC}"
    export $(grep -v '^#' ".env" | xargs)
  fi
}


# Function to open a URL in the default browser
open_url() {
  local url=$1

  # Check if URL is provided
  if [ -z "$url" ]; then
    echo -e "${RED}Error: No URL provided to open_url function${NC}"
    return 1
  fi

  echo -e "${GREEN}Opening URL: ${BLUE}${url}${NC}"

  # Open URL based on operating system
  if [[ "$OSTYPE" == "darwin"* ]]; then
    # macOS
    open "$url" &>/dev/null
  elif [[ "$OSTYPE" == "linux-gnu"* ]]; then
    # Linux - try various browsers
    if command -v xdg-open &>/dev/null; then
      xdg-open "$url" &>/dev/null
    elif command -v gnome-open &>/dev/null; then
      gnome-open "$url" &>/dev/null
    elif command -v sensible-browser &>/dev/null; then
      sensible-browser "$url" &>/dev/null
    else
      echo -e "${YELLOW}Warning: Could not determine browser to open URL${NC}"
      echo -e "Please open ${BLUE}${url}${NC} manually in your browser"
      return 1
    fi
  else
    # Other OS - just display the URL
    echo -e "${YELLOW}Warning: Automatic URL opening not supported on this OS${NC}"
    echo -e "Please open ${BLUE}${url}${NC} manually in your browser"
    return 1
  fi

  return 0
}

# Docker operations
docker_up() {
  echo -e "${GREEN}Starting services with Docker Compose...${NC}"
  docker-compose -f build/docker-compose.yaml up -d
  echo -e "${GREEN}Services started successfully${NC}"
  echo -e "${BLUE}Access points:${NC}"
  echo -e "  zPages: http://localhost:15679"
  echo -e "  Prometheus: http://localhost:19090"
  echo -e "  Grafana: http://localhost:13000 (admin/admin)"
  echo -e "  Mock New Relic: http://localhost:18080"

  # Give services a moment to initialize
  echo -e "${YELLOW}Waiting for services to initialize...${NC}"
  sleep 2

  # Open main URLs in browser
  if [ "${OPEN_URLS:-true}" = "true" ]; then
    echo -e "${GREEN}Opening service interfaces in your browser...${NC}"
    # Open Grafana dashboard
    open_url "http://localhost:13000" || true
    # Open Prometheus
    open_url "http://localhost:19090" || true
    # Open zPages
    open_url "http://localhost:15679" || true
  else
    echo -e "${YELLOW}URL auto-opening is disabled. Set OPEN_URLS=true to enable.${NC}"
  fi
}

docker_down() {
  echo -e "${YELLOW}Stopping services...${NC}"
  docker-compose -f build/docker-compose.yaml down
  echo -e "${GREEN}Services stopped successfully${NC}"
}

docker_logs() {
  echo -e "${YELLOW}Following logs...${NC}"
  docker-compose -f build/docker-compose.yaml logs -f
}

docker_restart() {
  echo -e "${YELLOW}Restarting services...${NC}"
  docker-compose -f build/docker-compose.yaml restart
  echo -e "${GREEN}Services restarted successfully${NC}"
}

# Determine the command
if [ $# -eq 0 ]; then
  display_usage
  exit 1
fi

# Set default options
CMD=""
OPEN_URLS="true"

# Check for help flag first
if [ "$1" == "--help" ] || [ "$1" == "-h" ] || [ "$1" == "help" ]; then
  display_usage
  exit 0
fi

# Parse arguments
CMD="$1"
shift

# Parse options
while [ $# -gt 0 ]; do
  case "$1" in
    --no-browser)
      OPEN_URLS="false"
      shift
      ;;
    --open-urls)
      OPEN_URLS="true"
      shift
      ;;
    *)
      echo -e "${RED}Unknown option: $1${NC}"
      display_usage
      exit 1
      ;;
  esac
done

export OPEN_URLS

echo -e "${YELLOW}Starting NRDOT Process-Metrics Optimization...${NC}"

# Load environment variables and check Docker is running
load_env
check_docker

# Process commands
case "$CMD" in
  up)
    docker_up
    ;;
  down)
    docker_down
    ;;
  logs)
    docker_logs
    ;;
  restart)
    docker_restart
    ;;
  *)
    echo -e "${RED}Unknown command: $CMD${NC}"
    display_usage
    exit 1
    ;;
esac