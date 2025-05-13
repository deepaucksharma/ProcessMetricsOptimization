#!/bin/bash
# Script to update the nrdot-config ConfigMap with the latest config.yaml
# Simplifies the deployment of all-profiles-deployments.yaml

set -e

# Check if the config.yaml file exists
if [ ! -f "../config.yaml" ]; then
  echo "Error: config.yaml not found in parent directory"
  exit 1
fi

# Create or update the ConfigMap
echo "Updating nrdot-config ConfigMap with latest config.yaml..."
kubectl -n observability create configmap nrdot-config \
  --from-file=config.yaml=../config.yaml \
  --dry-run=client -o yaml | kubectl apply -f -

echo "ConfigMap updated successfully!"