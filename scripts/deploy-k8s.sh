#!/bin/bash
# =============================================================================
# NeoBank K8s Deployment Helper
# Substitutes environment variables in K8s manifests before applying
# =============================================================================

set -e

# Required environment variables
REQUIRED_VARS=(
    "AWS_ACCOUNT_ID"
    "AWS_REGION"
    "ENVIRONMENT"
    "IMAGE_TAG"
)

# Check required variables
echo "Checking required environment variables..."
for var in "${REQUIRED_VARS[@]}"; do
    if [ -z "${!var}" ]; then
        echo "ERROR: Required environment variable $var is not set"
        exit 1
    fi
done

echo "Environment configuration:"
echo "  AWS_ACCOUNT_ID: $AWS_ACCOUNT_ID"
echo "  AWS_REGION: $AWS_REGION"
echo "  ENVIRONMENT: $ENVIRONMENT"
echo "  IMAGE_TAG: $IMAGE_TAG"

# Build the kustomize output with variable substitution
echo ""
echo "Building K8s manifests with kustomize..."
OVERLAY_DIR="k8s/overlays/eks-${ENVIRONMENT}"

if [ ! -d "$OVERLAY_DIR" ]; then
    echo "ERROR: Overlay directory not found: $OVERLAY_DIR"
    exit 1
fi

# Create a temporary file for substituted manifests
TEMP_FILE=$(mktemp)
trap "rm -f $TEMP_FILE" EXIT

# Use kustomize build and envsubst for variable substitution
kustomize build "$OVERLAY_DIR" | envsubst > "$TEMP_FILE"

echo "Manifests generated successfully!"
echo ""

# Option to apply or just generate
if [ "$1" == "--apply" ]; then
    echo "Applying manifests to cluster..."
    kubectl apply -f "$TEMP_FILE"
    echo "Done!"
elif [ "$1" == "--dry-run" ]; then
    echo "Dry run - showing what would be applied:"
    kubectl apply -f "$TEMP_FILE" --dry-run=client
else
    echo "Generated manifests saved to: $TEMP_FILE"
    echo ""
    echo "Usage:"
    echo "  $0 --apply    # Apply manifests to cluster"
    echo "  $0 --dry-run  # Show what would be applied"
    echo ""
    echo "To see the generated manifests:"
    echo "  cat $TEMP_FILE"
fi
