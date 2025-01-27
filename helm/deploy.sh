#!/bin/bash

##############################################################################
# Helper functions
##############################################################################

function error_exit {
  echo "Error: $1" >&2
  exit 1
}

function usage {
  echo "Usage: $0 image-path=<image-path> dns-host=<dns-host> tag=<tag> [ingress-path=<ingress-path>] [namespace=<namespace>]"
  echo
  echo "Description:"
  echo "  Deploys (or upgrades) the 'indexer' Helm chart with the specified options."
  echo
  echo "Required arguments (in any order):"
  echo "  image-path=<image-path>  Path to the Docker image (e.g., docker-repo/app)"
  echo "  dns-host=<dns-host>      DNS host for Ingress (e.g., app.local)"
  echo "  tag=<tag>                Docker image tag (e.g., 31)"
  echo
  echo "Optional arguments (in any order):"
  echo "  ingress-path=<ingress-path>  Ingress path (default: /)"
  echo "  namespace=<namespace>        Kubernetes namespace (default: default)"
  echo
  echo "Environment variables (optional for image pull secrets):"
  echo "  DOCKER_REGISTRY_URL, DOCKER_REGISTRY_USERNAME, DOCKER_REGISTRY_PASSWORD, IMAGE_PULL_SECRET_NAME"
  echo
  echo "Examples:"
  echo "  # Minimal usage with required arguments only:"
  echo "  $0 image-path=docker-repo/app dns-host=app.local tag=latest"
  echo
  echo "  # Custom ingress path, custom namespace:"
  echo "  $0 dns-host=app.local image-path=docker-repo/app tag=latest ingress-path=/api namespace=my-namespace"
  exit 1
}

##############################################################################
# Parse key-value arguments
##############################################################################

# Default values for optional arguments
INGRESS_PATH="/"
NAMESPACE="default"

# Track which required arguments we've seen
IMAGE_PATH=""
DNS_HOST=""
IMAGE_TAG=""

# Iterate over each argument, parse key=value pairs
for arg in "$@"; do
  case $arg in
    image-path=*)
      IMAGE_PATH="${arg#*=}"
      ;;
    dns-host=*)
      DNS_HOST="${arg#*=}"
      ;;
    tag=*)
      IMAGE_TAG="${arg#*=}"
      ;;
    ingress-path=*)
      INGRESS_PATH="${arg#*=}"
      ;;
    namespace=*)
      NAMESPACE="${arg#*=}"
      ;;
    *)
      echo "Invalid argument: $arg"
      usage
      ;;
  esac
done

##############################################################################
# Validate required arguments
##############################################################################

# If any of the required variables are empty, show usage and exit
[ -z "$IMAGE_PATH" ] && { echo "Missing required argument: image-path"; usage; }
[ -z "$DNS_HOST" ] && { echo "Missing required argument: dns-host"; usage; }
[ -z "$IMAGE_TAG" ] && { echo "Missing required argument: tag"; usage; }

# Check if Helm is installed
command -v helm >/dev/null 2>&1 || { error_exit "Helm is not installed. Please install Helm."; }

##############################################################################
# Image Pull Secret variables from the environment
##############################################################################

DOCKER_REGISTRY_URL="${DOCKER_REGISTRY_URL}"
DOCKER_REGISTRY_USERNAME="${DOCKER_REGISTRY_USERNAME}"
DOCKER_REGISTRY_PASSWORD="${DOCKER_REGISTRY_PASSWORD}"
IMAGE_PULL_SECRET_NAME="${IMAGE_PULL_SECRET_NAME:-my-registry-secret}" # default if not set

##############################################################################
# Build the Helm command
##############################################################################

HELM_CMD="helm upgrade --install indexer ./ \
  --namespace ${NAMESPACE} \
  --create-namespace \
  --set image.repository=${IMAGE_PATH} \
  --set image.tag=${IMAGE_TAG} \
  --set ingress.enabled=true \
  --set ingress.hosts[0].host=${DNS_HOST} \
  --set ingress.hosts[0].paths[0]=${INGRESS_PATH}"

# If all the registry credentials are present, configure imagePullSecret
if [[ -n "$DOCKER_REGISTRY_URL" && -n "$DOCKER_REGISTRY_USERNAME" && -n "$DOCKER_REGISTRY_PASSWORD" ]]; then
  HELM_CMD+=" \
  --set imagePullSecret.name=${IMAGE_PULL_SECRET_NAME} \
  --set imagePullSecret.registry=${DOCKER_REGISTRY_URL} \
  --set imagePullSecret.username=${DOCKER_REGISTRY_USERNAME} \
  --set imagePullSecret.password=${DOCKER_REGISTRY_PASSWORD}"
fi

##############################################################################
# Execute Helm command
##############################################################################

echo "Executing Helm command:"
echo "$HELM_CMD"
eval "$HELM_CMD"

# Check success
if [ "$?" -eq 0 ]; then
  echo "Helm chart successfully installed/updated."
else
  error_exit "Helm chart installation/update failed."
fi
