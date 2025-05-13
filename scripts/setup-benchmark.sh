#!/bin/bash
set -e

# Configuration
IMAGE_TAG=${1:-"bench"}
DURATION=${2:-"5m"}
PROFILE=${3:-"max-throughput-traces"}

echo "Setting up benchmark environment with:"
echo "Image Tag: $IMAGE_TAG"
echo "Duration: $DURATION"
echo "Profile: $PROFILE"

# Step 1: Build the Docker image
echo "Building Docker image..."
docker build -t ghcr.io/deepaucksharma/nrdot-reservoir:$IMAGE_TAG -f build/docker/Dockerfile.bench .

# Step 2: Check if KinD cluster exists, create if not
if ! kind get clusters | grep -q kind; then
  echo "Creating KinD cluster..."
  kind create cluster --config infra/kind/bench-config.yaml
else
  echo "Using existing KinD cluster"
fi

# Step 3: Load the image into KinD
echo "Loading image into KinD..."
kind load docker-image ghcr.io/deepaucksharma/nrdot-reservoir:$IMAGE_TAG

# Step 4: Run a manual test using kubectl
echo "Running a manual test deployment..."
cat <<EOF | kubectl apply -f -
apiVersion: apps/v1
kind: Deployment
metadata:
  name: reservoir-test
  namespace: default
spec:
  replicas: 1
  selector:
    matchLabels:
      app: reservoir-test
  template:
    metadata:
      labels:
        app: reservoir-test
    spec:
      containers:
      - name: reservoir
        image: ghcr.io/deepaucksharma/nrdot-reservoir:$IMAGE_TAG
        resources:
          limits:
            cpu: 500m
            memory: 512Mi
          requests:
            cpu: 100m
            memory: 128Mi
        env:
        - name: RESERVOIR_SIZE
          value: "5000"
        - name: WINDOW_DURATION
          value: "60s"
        - name: TRACE_AWARE
          value: "true"
EOF

echo "Waiting for pod to be ready..."
kubectl wait --for=condition=ready pod -l app=reservoir-test --timeout=60s

echo "Checking container logs:"
POD_NAME=$(kubectl get pods -l app=reservoir-test -o jsonpath="{.items[0].metadata.name}")
kubectl logs $POD_NAME

echo "Benchmark setup complete. To clean up, run:"
echo "kubectl delete deployment reservoir-test"
echo "kind delete cluster"