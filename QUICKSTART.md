# Quick Start: Deploy to Docker Desktop Kubernetes

This guide will help you deploy the NeoBank platform with FinOps, DR, and Istio to your local Docker Desktop Kubernetes cluster.

## Prerequisites

1. **Docker Desktop** with Kubernetes enabled
   - Settings → Kubernetes → Enable Kubernetes
   - Allocate at least 8GB RAM and 4 CPUs

2. **kubectl** installed and configured

   ```powershell
   kubectl version --client
   kubectl cluster-info
   ```

3. **Optional but recommended:**
   - [Istio](https://istio.io/latest/docs/setup/getting-started/#download) for service mesh
   - [Velero](https://velero.io/docs/v1.12/basic-install/) for disaster recovery

## Quick Deploy (Recommended)

### Option 1: Full Stack with Istio

```powershell
# Build images and deploy everything with Istio
.\scripts\deploy-local.ps1 -InstallIstio -BuildImages
```

### Option 2: Basic Deployment (No Istio/Velero)

```powershell
# Just deploy the application with FinOps
.\scripts\deploy-local.ps1 -BuildImages
```

### Option 3: Clean Install

```powershell
# Remove existing deployment and start fresh
.\scripts\deploy-local.ps1 -Clean -InstallIstio -BuildImages
```

## Step-by-Step Deployment

### Step 1: Verify Prerequisites

```powershell
.\scripts\verify-deployment.ps1
```

This shows what's currently deployed and what's missing.

### Step 2: Build Docker Images

```powershell
# Build all service images
docker build -t neobank/identity-service:local backend/identity-service
docker build -t neobank/ledger-service:local backend/ledger-service
docker build -t neobank/payment-service:local backend/payment-service
docker build -t neobank/product-service:local backend/product-service
docker build -t neobank/card-service:local backend/card-service
docker build -t neobank/frontend:local frontend
```

Or use the deploy script with `-BuildImages` flag.

### Step 3: Install Istio (Optional)

```powershell
# Download istioctl
# https://istio.io/latest/docs/setup/getting-started/#download

# Install Istio
istioctl install --set profile=demo -y

# Verify installation
kubectl get pods -n istio-system
```

### Step 4: Deploy NeoBank

```powershell
# Deploy with local overlay (includes FinOps and DR configs)
kubectl apply -k k8s/overlays/local/
```

Or use the deploy script:

```powershell
.\scripts\deploy-local.ps1
```

### Step 5: Verify Deployment

```powershell
# Check all resources
.\scripts\verify-deployment.ps1

# Check pods
kubectl get pods -n neobank

# Check services
kubectl get services -n neobank

# Check FinOps resources
kubectl get resourcequotas,limitranges -n neobank
```

### Step 6: Access Services

Set up port forwarding:

```powershell
# Automated port forwarding
.\scripts\setup-e2e.ps1 -PortForward

# Or manually:
kubectl port-forward svc/local-neobank-frontend 3001:3000 -n neobank
kubectl port-forward svc/istio-ingressgateway 8080:80 -n istio-system
```

Access:

- **Frontend:** http://localhost:3001
- **Istio Gateway:** http://localhost:8080
- **Identity Service:** http://localhost:8081/health

## What Gets Deployed

### FinOps Features ✓

- Resource quotas per namespace
- Limit ranges for cost control
- Cost allocation labels
- Budget monitoring (Prometheus alerts)
- Resource utilization tracking

### Disaster Recovery ✓

- Velero backup schedules (hourly tier1, 6h tier2, daily full)
- Automated restore procedures
- Monthly DR testing CronJob
- Failover configurations

### Istio Service Mesh ✓

- mTLS between services
- Circuit breakers and connection pooling
- Rate limiting per service
- Authorization policies
- Traffic management (retries, timeouts)
- Observability (metrics, tracing)

### Application Services ✓

- Identity Service (authentication)
- Ledger Service (accounts, transactions)
- Payment Service (payments, transfers)
- Product Service (products, offers)
- Card Service (card management)
- Frontend (Next.js web app)

## Verification Commands

```powershell
# Overall status
.\scripts\verify-deployment.ps1

# Pod status
kubectl get pods -n neobank -o wide

# FinOps resources
kubectl describe resourcequota -n neobank
kubectl describe limitrange -n neobank

# Istio configuration
kubectl get gateway,virtualservices,destinationrules -n neobank

# Velero backups
kubectl get schedules -n velero
kubectl get backups -n velero

# Check resource usage
kubectl top pods -n neobank
kubectl top nodes
```

## Running E2E Tests

```powershell
# Run all tests
.\scripts\run-e2e-tests.ps1 -All

# Run specific test suites
.\scripts\run-e2e-tests.ps1 -FinOps
.\scripts\run-e2e-tests.ps1 -DR
.\scripts\run-e2e-tests.ps1 -Istio
.\scripts\run-e2e-tests.ps1 -Chaos
```

## Troubleshooting

### Pods Not Starting

```powershell
# Check pod status
kubectl get pods -n neobank

# Check pod details
kubectl describe pod <pod-name> -n neobank

# Check logs
kubectl logs <pod-name> -n neobank

# Check events
kubectl get events -n neobank --sort-by='.lastTimestamp'
```

### Images Not Found

Make sure images are built with correct tags:

```powershell
# List images
docker images | Select-String "neobank"

# Rebuild if needed
.\scripts\deploy-local.ps1 -BuildImages
```

### Resource Constraints

Check if Docker Desktop has enough resources:

```powershell
# Check node resources
kubectl describe node

# Increase Docker Desktop resources:
# Settings → Resources → Memory: 8GB+, CPUs: 4+
```

### Port Already in Use

```powershell
# Kill existing port forwards
Get-Process -Name kubectl | Where-Object {$_.CommandLine -like "*port-forward*"} | Stop-Process

# Or restart Docker Desktop
```

### Istio Not Working

```powershell
# Verify Istio installation
istioctl analyze -n neobank

# Check sidecar injection
kubectl get pods -n neobank -o jsonpath='{.items[*].spec.containers[*].name}'
# Should show "istio-proxy" alongside app containers

# Re-label namespace
kubectl label namespace neobank istio-injection=enabled --overwrite

# Restart pods
kubectl rollout restart deployment -n neobank
```

## Cleanup

```powershell
# Delete application
kubectl delete namespace neobank

# Delete FinOps resources
kubectl delete namespace finops

# Uninstall Istio
istioctl uninstall --purge -y
kubectl delete namespace istio-system

# Uninstall Velero
kubectl delete namespace velero
```

## Next Steps

1. ✅ Verify deployment: `.\scripts\verify-deployment.ps1`
2. ✅ Set up port forwarding: `.\scripts\setup-e2e.ps1 -PortForward`
3. ✅ Run E2E tests: `.\scripts\run-e2e-tests.ps1 -All`
4. ✅ Monitor resources: `kubectl get all -n neobank`
5. ✅ Check FinOps metrics: `kubectl describe resourcequota -n neobank`
6. ✅ View Istio dashboard: `istioctl dashboard kiali`

## Additional Resources

- [Full E2E Test Documentation](../tests/e2e/README.md)
- [FinOps Configuration Details](../k8s/finops/README.md)
- [DR Configuration Details](../k8s/dr/README.md)
- [Istio Setup Guide](../k8s/istio/README.md)
- [Local Overlay Configuration](../k8s/overlays/local/README.md)
