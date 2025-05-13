#!/bin/bash
# Script to update New Relic license key in Kubernetes

# Check if a license key was provided
if [ -z "$1" ]; then
  echo "Error: No license key provided"
  echo "Usage: ./update-license-key.sh YOUR_LICENSE_KEY"
  exit 1
fi

NR_KEY=$1

# Update the Kubernetes secret
echo "Updating New Relic license key in Kubernetes..."
kubectl -n observability create secret generic newrelic-license \
  --from-literal=NEW_RELIC_LICENSE_KEY=$NR_KEY \
  --dry-run=client -o yaml | kubectl apply -f -

# Restart the collector pods
echo "Restarting collector pods..."
kubectl -n observability delete pod -l app=nrdot-collector-host

# Wait for pods to restart
echo "Waiting for pods to restart..."
sleep 5

# Run the connectivity test
echo "Running connectivity test..."
kubectl apply -f k8s/test-connectivity.yaml

# Wait for test to complete
echo "Waiting for test to complete..."
sleep 10

# Show test results
echo "Test results:"
kubectl -n observability logs -l job-name=test-nr-connectivity

echo ""
echo "If the test shows 200 responses, your setup is working correctly!"
echo "If you see 401/403 errors, please check your license key permissions."