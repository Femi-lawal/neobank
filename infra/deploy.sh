#!/bin/bash
# ============================================================================
# NeoBank AWS Deployment Script
# ============================================================================
set -euo pipefail

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Default values
ENVIRONMENT="${ENVIRONMENT:-dev}"
AWS_REGION="${AWS_REGION:-us-east-1}"
ACTION="${1:-plan}"

# Script directory
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
TERRAFORM_DIR="${SCRIPT_DIR}/terraform"
K8S_DIR="${SCRIPT_DIR}/../k8s/overlays/aws"

log_info() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

log_warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

check_prerequisites() {
    log_info "Checking prerequisites..."
    
    local missing=()
    
    if ! command -v terraform &> /dev/null; then
        missing+=("terraform")
    fi
    
    if ! command -v aws &> /dev/null; then
        missing+=("aws-cli")
    fi
    
    if ! command -v kubectl &> /dev/null; then
        missing+=("kubectl")
    fi
    
    if ! command -v kustomize &> /dev/null; then
        missing+=("kustomize")
    fi
    
    if [ ${#missing[@]} -gt 0 ]; then
        log_error "Missing required tools: ${missing[*]}"
        exit 1
    fi
    
    # Check AWS credentials
    if ! aws sts get-caller-identity &> /dev/null; then
        log_error "AWS credentials not configured. Please run 'aws configure' or set environment variables."
        exit 1
    fi
    
    log_info "All prerequisites met."
}

init_terraform() {
    log_info "Initializing Terraform..."
    cd "${TERRAFORM_DIR}"
    
    terraform init -upgrade
    
    log_info "Terraform initialized."
}

validate_terraform() {
    log_info "Validating Terraform configuration..."
    cd "${TERRAFORM_DIR}"
    
    terraform validate
    terraform fmt -check -recursive
    
    log_info "Terraform configuration is valid."
}

plan_terraform() {
    log_info "Planning Terraform changes for ${ENVIRONMENT}..."
    cd "${TERRAFORM_DIR}"
    
    terraform plan \
        -var-file="environments/${ENVIRONMENT}.tfvars" \
        -out="tfplan-${ENVIRONMENT}"
    
    log_info "Terraform plan saved to tfplan-${ENVIRONMENT}"
}

apply_terraform() {
    log_info "Applying Terraform changes for ${ENVIRONMENT}..."
    cd "${TERRAFORM_DIR}"
    
    if [ ! -f "tfplan-${ENVIRONMENT}" ]; then
        log_error "No plan file found. Run 'plan' first."
        exit 1
    fi
    
    terraform apply "tfplan-${ENVIRONMENT}"
    
    # Save outputs for later use
    terraform output -json > "outputs-${ENVIRONMENT}.json"
    
    log_info "Terraform apply completed."
}

destroy_terraform() {
    log_warn "Destroying Terraform infrastructure for ${ENVIRONMENT}..."
    cd "${TERRAFORM_DIR}"
    
    read -p "Are you sure you want to destroy all resources? (yes/no): " confirm
    if [ "$confirm" != "yes" ]; then
        log_info "Destroy cancelled."
        exit 0
    fi
    
    terraform destroy \
        -var-file="environments/${ENVIRONMENT}.tfvars" \
        -auto-approve
    
    log_info "Terraform destroy completed."
}

configure_kubectl() {
    log_info "Configuring kubectl for EKS cluster..."
    cd "${TERRAFORM_DIR}"
    
    local cluster_name
    cluster_name=$(terraform output -raw eks_cluster_name 2>/dev/null || echo "")
    
    if [ -z "$cluster_name" ]; then
        log_error "Could not get EKS cluster name from Terraform outputs."
        exit 1
    fi
    
    aws eks update-kubeconfig \
        --region "${AWS_REGION}" \
        --name "${cluster_name}"
    
    log_info "kubectl configured for cluster: ${cluster_name}"
}

export_terraform_outputs() {
    log_info "Exporting Terraform outputs for Kubernetes..."
    cd "${TERRAFORM_DIR}"
    
    # Export outputs as environment variables
    export AWS_REGION="${AWS_REGION}"
    export RDS_ENDPOINT=$(terraform output -raw rds_endpoint)
    export RDS_PORT=$(terraform output -raw rds_port)
    export DB_NAME=$(terraform output -raw rds_database_name)
    export REDIS_ENDPOINT=$(terraform output -raw redis_endpoint)
    export REDIS_PORT=$(terraform output -raw redis_port)
    export DB_SECRET_ARN=$(terraform output -raw db_secret_arn)
    export REDIS_SECRET_ARN=$(terraform output -raw redis_secret_arn)
    export JWT_SECRET_ARN=$(terraform output -raw jwt_secret_arn)
    export IDENTITY_SERVICE_ROLE_ARN=$(terraform output -raw identity_service_role_arn)
    export LEDGER_SERVICE_ROLE_ARN=$(terraform output -raw ledger_service_role_arn)
    export PAYMENT_SERVICE_ROLE_ARN=$(terraform output -raw payment_service_role_arn)
    export PRODUCT_SERVICE_ROLE_ARN=$(terraform output -raw product_service_role_arn)
    export CARD_SERVICE_ROLE_ARN=$(terraform output -raw card_service_role_arn)
    
    # MSK (optional)
    export MSK_BOOTSTRAP_BROKERS=$(terraform output -raw msk_bootstrap_brokers 2>/dev/null || echo "")
    
    log_info "Terraform outputs exported."
}

deploy_kubernetes() {
    log_info "Deploying Kubernetes resources..."
    
    export_terraform_outputs
    configure_kubectl
    
    # Create namespace if not exists
    kubectl create namespace neobank --dry-run=client -o yaml | kubectl apply -f -
    
    # Apply AWS config
    cd "${K8S_DIR}"
    
    # Substitute environment variables in configmap
    envsubst < configmap.yaml | kubectl apply -f -
    envsubst < service-accounts.yaml | kubectl apply -f -
    
    # Apply kustomization
    kustomize build . | kubectl apply -f -
    
    # Wait for deployments
    log_info "Waiting for deployments to be ready..."
    kubectl rollout status deployment -n neobank --timeout=300s
    
    log_info "Kubernetes deployment completed."
}

show_status() {
    log_info "Cluster Status:"
    
    configure_kubectl 2>/dev/null || true
    
    echo ""
    echo "=== Namespaces ==="
    kubectl get namespaces | grep -E "neobank|NAME" || echo "Namespace not found"
    
    echo ""
    echo "=== Deployments ==="
    kubectl get deployments -n neobank 2>/dev/null || echo "No deployments found"
    
    echo ""
    echo "=== Pods ==="
    kubectl get pods -n neobank 2>/dev/null || echo "No pods found"
    
    echo ""
    echo "=== Services ==="
    kubectl get services -n neobank 2>/dev/null || echo "No services found"
    
    echo ""
    echo "=== Ingress ==="
    kubectl get ingress -n neobank 2>/dev/null || echo "No ingress found"
}

print_usage() {
    echo "Usage: $0 <action> [environment]"
    echo ""
    echo "Actions:"
    echo "  init      - Initialize Terraform"
    echo "  validate  - Validate Terraform configuration"
    echo "  plan      - Plan Terraform changes"
    echo "  apply     - Apply Terraform changes"
    echo "  destroy   - Destroy all infrastructure"
    echo "  deploy-k8s - Deploy Kubernetes resources"
    echo "  status    - Show deployment status"
    echo "  all       - Run init, plan, apply, and deploy-k8s"
    echo ""
    echo "Environment variables:"
    echo "  ENVIRONMENT - Target environment (dev, staging, prod). Default: dev"
    echo "  AWS_REGION  - AWS region. Default: us-east-1"
    echo ""
    echo "Examples:"
    echo "  $0 plan"
    echo "  ENVIRONMENT=prod $0 apply"
    echo "  $0 all"
}

main() {
    case "${ACTION}" in
        init)
            check_prerequisites
            init_terraform
            ;;
        validate)
            check_prerequisites
            validate_terraform
            ;;
        plan)
            check_prerequisites
            init_terraform
            validate_terraform
            plan_terraform
            ;;
        apply)
            check_prerequisites
            apply_terraform
            ;;
        destroy)
            check_prerequisites
            destroy_terraform
            ;;
        deploy-k8s)
            check_prerequisites
            deploy_kubernetes
            ;;
        status)
            show_status
            ;;
        all)
            check_prerequisites
            init_terraform
            validate_terraform
            plan_terraform
            apply_terraform
            deploy_kubernetes
            show_status
            ;;
        help|--help|-h)
            print_usage
            ;;
        *)
            log_error "Unknown action: ${ACTION}"
            print_usage
            exit 1
            ;;
    esac
}

main
