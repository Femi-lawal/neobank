# NeoBank Kubernetes Deployment

This directory contains Kubernetes manifests for deploying NeoBank.

## Directory Structure

```
k8s/
├── base/                    # Base manifests
│   ├── namespace.yaml       # Namespace definition
│   ├── secrets.yaml         # Secrets (JWT, DB credentials)
│   ├── configmap.yaml       # Configuration
│   ├── services/            # ClusterIP services
│   └── deployments/         # Deployment manifests
└── overlays/                # Kustomize overlays
    ├── dev/                 # Development configuration
    └── prod/                # Production configuration
```

## Quick Start

```bash
# Deploy to development
kubectl apply -k k8s/overlays/dev

# Deploy to production
kubectl apply -k k8s/overlays/prod

# Or apply base manifests directly
kubectl apply -f k8s/base/

# Check deployment status
kubectl get pods -n neobank
kubectl get svc -n neobank

# View logs
kubectl logs -n neobank -l app=identity-service
```

## Prerequisites

- Kubernetes cluster (minikube, kind, GKE, EKS, AKS)
- kubectl configured
- Docker images built and pushed to registry

## Building Images

```bash
# Build all images
docker build -t neobank/identity-service:latest -f backend/identity-service/Dockerfile ./backend
docker build -t neobank/ledger-service:latest -f backend/ledger-service/Dockerfile ./backend
docker build -t neobank/payment-service:latest -f backend/payment-service/Dockerfile ./backend
docker build -t neobank/product-service:latest -f backend/product-service/Dockerfile ./backend
docker build -t neobank/card-service:latest -f backend/card-service/Dockerfile ./backend
docker build -t neobank/frontend:latest -f frontend/Dockerfile ./frontend
```

## Services

| Service | Internal Port | Description |
|---------|--------------|-------------|
| frontend | 3000 | Next.js web app |
| identity-service | 8081 | Authentication |
| ledger-service | 8082 | Accounts & Transactions |
| payment-service | 8083 | Transfers |
| product-service | 8084 | Banking products |
| card-service | 8085 | Card management |
| postgres | 5432 | Database |
| redis | 6379 | Cache |
