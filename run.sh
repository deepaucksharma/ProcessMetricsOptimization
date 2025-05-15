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
  echo -e "Usage: ./run.sh [mode] [command] [options]"
  echo
  echo -e "Modes:"
  echo -e "  docker    Run using Docker containers (default)"
  echo -e "  demo      Run simple demo app directly"
  echo
  echo -e "Docker Commands:"
  echo -e "  up        Start all services"
  echo -e "  down      Stop all services"
  echo -e "  logs      Follow logs from all services"
  echo -e "  restart   Restart all services"
  echo
  echo -e "Options:"
  echo -e "  --no-browser   Don't automatically open URLs in the browser"
  echo -e "  --open-urls    Open URLs in the browser (default)"
  echo
  echo -e "Examples:"
  echo -e "  ./run.sh docker up                Start services with Docker"
  echo -e "  ./run.sh demo                     Run the simple demo"
  echo -e "  ./run.sh up                       Start services with Docker (default mode)"
  echo -e "  ./run.sh up --no-browser          Start services without opening URLs"
  echo -e "  ./run.sh demo --no-browser        Run demo without opening browser"
  echo
  echo -e "Access Points (Docker mode):"
  echo -e "  zPages: http://localhost:15679"
  echo -e "  Prometheus: http://localhost:19090"
  echo -e "  Grafana: http://localhost:13000 (admin/admin)"
  echo -e "  Mock New Relic: http://localhost:18080"
  echo
}

# Check Docker is running (for Docker mode)
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

# Check if NEW_RELIC_LICENSE_KEY is set (for Docker mode)
check_license_key() {
  if [ -z "$NEW_RELIC_LICENSE_KEY" ]; then
    echo -e "${YELLOW}Warning: NEW_RELIC_LICENSE_KEY environment variable is not set.${NC}"
    echo -e "Data will not be sent to New Relic. For production use, set this key in .env file."
  fi
}

# Get the machine's IP address (for demo mode)
get_ip_address() {
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
  
  echo $IP
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

# Run the simple demo
run_demo() {
  echo -e "${YELLOW}Starting NRDOT Hello World Processor Demo...${NC}"
  
  IP=$(get_ip_address)
  
  echo -e "${GREEN}Detected IP Address: ${BLUE}${IP}${NC}"
  echo -e "${GREEN}Starting server...${NC}"
  echo -e "${GREEN}You can access the demo at:${NC}"
  echo -e "  - ${BLUE}http://localhost:8080${NC} (if on the same machine)"
  echo -e "  - ${BLUE}http://${IP}:8080${NC} (from other devices on the same network)"
  echo -e ""
  echo -e "${YELLOW}Press Ctrl+C to stop the server${NC}"
  echo -e ""
  
  # Open demo URL in browser
  if [ "${OPEN_URLS:-true}" = "true" ]; then
    # Open in background to allow server to start
    (
      # Wait a moment for the server to start
      sleep 1
      echo -e "${GREEN}Opening demo in your browser...${NC}"
      open_url "http://localhost:8080" || true
    ) &
  fi
  
  # Run the demo
  go run main.go
}

# Docker mode operations
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

# Determine the mode and command
if [ $# -eq 0 ]; then
  display_usage
  exit 1
fi

# Set default mode to docker and default options
MODE="docker"
CMD=""
OPEN_URLS="true"

# Check for help flag first
if [ "$1" == "--help" ] || [ "$1" == "-h" ] || [ "$1" == "help" ]; then
  display_usage
  exit 0
fi

# Parse arguments
if [ "$1" == "docker" ] || [ "$1" == "demo" ]; then
  MODE="$1"
  shift
  CMD="$1"
  if [ -n "$CMD" ]; then
    shift
  fi
else
  CMD="$1"
  shift
fi

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

# Process based on mode
case "$MODE" in
  docker)
    # Load environment variables and check Docker is running
    load_env
    check_docker
    check_license_key
    
    # Process Docker commands
    case "$CMD" in
      up|"")
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
        echo -e "${RED}Unknown command for docker mode: $CMD${NC}"
        display_usage
        exit 1
        ;;
    esac
    ;;
  
  demo)
    # Run the demo directly
    run_demo
    ;;
  
  *)
    echo -e "${RED}Unknown mode: $MODE${NC}"
    display_usage
    exit 1
    ;;
esac