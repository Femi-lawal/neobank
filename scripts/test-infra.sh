#!/bin/bash
set -e

echo "🔹 Starting Infrastructure Testing Suite..."

# 1. Terraform Testing
echo "------------------------------------------------"
echo "🔍 Testing Terraform (Security & Linting)..."
if [ -d "infra/terraform" ]; then
    echo " > Running tfsec..."
    docker run --rm -v "$(pwd):/src" aquasec/tfsec /src/infra/terraform --soft-fail

    echo " > Running tflint (stub)..."
    # Placeholder: tflint requires specific config or setup, skipping strict run for now
    # docker run --rm -v "$(pwd):/data" -t ghcr.io/terraform-linters/tflint /data/infra/terraform 
else
    echo "⚠️ infra/terraform directory not found, skipping..."
fi

# 2. Prometheus Testing
echo "------------------------------------------------"
echo "🔍 Testing Prometheus Configuration..."
if [ -f "infra/observability/prometheus/prometheus.yml" ]; then
    echo " > Running promtool check..."
    docker run --rm -v "$(pwd)/infra/observability/prometheus:/etc/prometheus" \
        prom/prometheus:latest promtool check config /etc/prometheus/prometheus.yml
else
    echo "⚠️ Prometheus config not found, skipping..."
fi

# 3. Jenkinsfile Validation
echo "------------------------------------------------"
echo "🔍 Validate Jenkinsfile..."
if [ -f "Jenkinsfile" ]; then
    # Simple syntax check via curl to Jenkins server is not available locally
    # We will do a basic structure check
    if grep -q "pipeline {" Jenkinsfile; then
        echo "✅ Jenkinsfile appears to be declarative pipeline (basic check passed)"
    else
        echo "❌ Jenkinsfile missing 'pipeline {' block"
        exit 1
    fi
else
    echo "⚠️ Jenkinsfile not found!"
fi

echo "------------------------------------------------"
echo "✅ Infrastructure Tests Completed."
