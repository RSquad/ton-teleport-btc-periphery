#!/bin/bash

# Function to output errors and exit the script
function error_exit {
  echo "Error: $1" >&2
  exit 1
}

# Function to display usage information
function usage {
  echo "Usage: $0 <image-path> <dns-host> <tag>"
  echo "  <image-path> - Path to the Docker image (e.g., your-docker-repo/app)"
  echo "  <dns-host>   - DNS host for Ingress (e.g., app.local)"
  echo "  <tag>        - Docker image tag (e.g., 31)"
  exit 1
}

# Check if Helm is installed
command -v helm >/dev/null 2>&1 || { error_exit "Helm is not installed. Please install Helm."; }

# Check the number of arguments
if [ "$#" -ne 3 ]; then
  usage
fi

# Assign arguments to variables
IMAGE_PATH="$1"
DNS_HOST="$2"
IMAGE_TAG="$3"

# Ensure that arguments are not empty
[ -z "$IMAGE_PATH" ] && { error_exit "Image path (image-path) is not provided."; }
[ -z "$DNS_HOST" ] && { error_exit "DNS host (dns-host) is not provided."; }
[ -z "$IMAGE_TAG" ] && { error_exit "Image tag (tag) is not provided."; }

# Variables for Image Pull Secret from the environment
# These should be set in the environment (e.g., via GitHub Actions)
DOCKER_REGISTRY_URL="${DOCKER_REGISTRY_URL}"
DOCKER_REGISTRY_USERNAME="${DOCKER_REGISTRY_USERNAME}"
DOCKER_REGISTRY_PASSWORD="${DOCKER_REGISTRY_PASSWORD}"
IMAGE_PULL_SECRET_NAME="${IMAGE_PULL_SECRET_NAME:-my-registry-secret}" # Default value if not set

# Formulate the Helm command
HELM_CMD="helm upgrade --install indexer ./ \
  --set image.repository=${IMAGE_PATH} \
  --set image.tag=${IMAGE_TAG} \
  --set ingress.hosts[0].host=${DNS_HOST}"

# If Image Pull Secret is enabled, add its parameters to the Helm command
# Check if all necessary variables for Image Pull Secret are set
if [[ -n "$DOCKER_REGISTRY_URL" && -n "$DOCKER_REGISTRY_USERNAME" && -n "$DOCKER_REGISTRY_PASSWORD" ]]; then
  HELM_CMD+=" \
  --set imagePullSecret.name=${IMAGE_PULL_SECRET_NAME} \
  --set imagePullSecret.registry=${DOCKER_REGISTRY_URL} \
  --set imagePullSecret.username=${DOCKER_REGISTRY_USERNAME} \
  --set imagePullSecret.password=${DOCKER_REGISTRY_PASSWORD}"
fi

# Execute the Helm command
echo "Executing Helm command:"
echo "$HELM_CMD"
eval $HELM_CMD

# Check if the Helm command was successful
if [ "$?" -eq 0 ]; then
  echo "Helm chart successfully installed/updated."
else
  error_exit "Helm chart installation/update failed."
fi
