# Local Development Overlay

This overlay is designed for local Kubernetes development using Docker Desktop.

## Features

- Reduced resource requests/limits for local development
- Single replica deployments
- Debug logging enabled
- FinOps and DR configurations included
- Istio service mesh support

## Prerequisites

- Docker Desktop with Kubernetes enabled
- kubectl configured
- Optional: Istio installed

## Usage

### Deploy with Kustomize

```bash
# From the k8s directory
kubectl apply -k overlays/local

# Or using kustomize directly
kustomize build overlays/local | kubectl apply -f -
```

### With Istio

First install Istio:

```bash
istioctl install --set profile=demo -y
kubectl label namespace neobank istio-injection=enabled
```

Then apply the Istio configurations:

```bash
kubectl apply -f ../istio/
```

### Verify Deployment

```bash
kubectl get pods -n neobank
kubectl get svc -n neobank
```

## Included Resources

### Base Resources

- All service deployments (identity, ledger, payment, product, card, frontend)
- Service accounts and RBAC
- Network policies
- HPA configurations
- Ingress

### FinOps Resources

- Resource quotas
- Limit ranges
- Cost allocation labels
- Budget alerts (simulated)

### DR Resources

- Velero backup schedules (requires Velero installation)
- Restore procedures
- DR testing CronJob
- Failover configuration

## Local Testing

### Port Forwarding

```bash
# Frontend
kubectl port-forward svc/local-neobank-frontend 3001:3000 -n neobank

# Identity Service
kubectl port-forward svc/local-neobank-identity-service 8081:8081 -n neobank

# All services via Istio Gateway
kubectl port-forward svc/istio-ingressgateway 8080:80 -n istio-system
```

### Access Services

- Frontend: http://localhost:3001
- Identity Service: http://localhost:8081/health
- Via Istio Gateway: http://localhost:8080

## Configuration

### Environment Variables

The local overlay sets:

- `LOG_LEVEL=debug`
- `ENVIRONMENT=local`
- `ENABLE_FINOPS=true`
- `ENABLE_DR_TESTING=true`

### Resource Limits

| Resource | Request | Limit |
| -------- | ------- | ----- |
| CPU      | 10m     | 200m  |
| Memory   | 64Mi    | 256Mi |

## Troubleshooting

### Pods Not Starting

Check events:

```bash
kubectl describe pod <pod-name> -n neobank
```

### Resource Issues

If pods are being evicted due to resource pressure:

```bash
# Increase Docker Desktop resources
# Settings > Resources > Memory: 8GB+ recommended
```

### Istio Issues

Verify Istio installation:

```bash
istioctl analyze -n neobank
kubectl get pods -n istio-system
```
