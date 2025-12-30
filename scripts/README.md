# Deployment Scripts Guide

This directory contains scripts to deploy and test NeoBank on Docker Desktop Kubernetes.

## Quick Start

```powershell
# 1. Deploy everything
.\scripts\deploy.ps1

# 2. Verify deployment
.\scripts\verify-deployment.ps1

# 3. Run tests
.\scripts\run-e2e-tests.ps1 -Cluster
```

## Available Scripts

### 🚀 Deployment Scripts

#### `deploy.ps1` (Recommended)

Complete deployment with validation and error handling.

```powershell
# Full deployment
.\scripts\deploy.ps1

# Clean install
.\scripts\deploy.ps1 -Clean

# Skip image building (use existing images)
.\scripts\deploy.ps1 -SkipBuild

# Skip validation (faster)
.\scripts\deploy.ps1 -SkipValidation
```

**What it does:**

- ✅ Validates manifests
- ✅ Builds Docker images
- ✅ Creates namespaces
- ✅ Applies FinOps configurations
- ✅ Deploys application
- ✅ Waits for pods to be ready
- ✅ Shows deployment status

---

#### `simple-deploy.ps1`

Minimal deployment script without extra features.

```powershell
.\scripts\simple-deploy.ps1
```

**Use when:** You just want basic deployment without Istio/Velero.

---

#### `deploy-local.ps1`

Advanced deployment with Istio and Velero support.

```powershell
# With Istio
.\scripts\deploy-local.ps1 -InstallIstio -BuildImages

# With Velero DR
.\scripts\deploy-local.ps1 -InstallVelero -BuildImages

# Full stack
.\scripts\deploy-local.ps1 -InstallIstio -InstallVelero -BuildImages

# Clean and redeploy
.\scripts\deploy-local.ps1 -Clean -InstallIstio -BuildImages
```

**What it does:**

- ✅ Optionally installs Istio
- ✅ Optionally installs Velero
- ✅ Builds images
- ✅ Applies all configurations
- ✅ Sets up port forwarding

---

### 🔍 Validation Scripts

#### `validate-manifests.ps1`

Validates all Kubernetes manifests before deployment.

```powershell
.\scripts\validate-manifests.ps1
```

**Checks:**

- Kustomize build validity
- FinOps manifest syntax
- DR manifest syntax
- Istio configuration
- Common issues

---

#### `verify-deployment.ps1`

Verifies what's currently deployed in the cluster.

```powershell
.\scripts\verify-deployment.ps1
```

**Checks:**

- ✅ Cluster connectivity
- ✅ Namespace existence
- ✅ Istio installation
- ✅ Velero installation
- ✅ FinOps resources
- ✅ Application deployments
- ✅ Istio configuration
- ✅ Pod status
- ✅ Services

**Output:** Shows what's working and what needs attention.

---

### 🧪 Testing Scripts

#### `run-e2e-tests.ps1`

Runs comprehensive E2E tests.

```powershell
# All tests
.\scripts\run-e2e-tests.ps1 -All

# Specific test suites
.\scripts\run-e2e-tests.ps1 -FinOps
.\scripts\run-e2e-tests.ps1 -DR
.\scripts\run-e2e-tests.ps1 -Istio
.\scripts\run-e2e-tests.ps1 -Chaos
.\scripts\run-e2e-tests.ps1 -Integration
.\scripts\run-e2e-tests.ps1 -Cluster

# With environment setup
.\scripts\run-e2e-tests.ps1 -Setup -All

# Verbose output
.\scripts\run-e2e-tests.ps1 -All -Verbose
```

---

#### `test-deployment.ps1`

Tests the deployment process without actually deploying.

```powershell
# Dry run
.\scripts\test-deployment.ps1 -DryRun

# Test without building images
.\scripts\test-deployment.ps1 -SkipBuild
```

---

#### `setup-e2e.sh` / `setup-e2e.ps1`

Sets up E2E test environment.

```powershell
# Windows
.\scripts\setup-e2e.ps1 -WithIstio -Build -PortForward

# Linux/macOS
./scripts/setup-e2e.sh --with-istio --build --port-forward
```

---

## Typical Workflows

### First Time Setup

```powershell
# 1. Validate everything first
.\scripts\validate-manifests.ps1

# 2. Deploy with Istio (recommended)
.\scripts\deploy-local.ps1 -InstallIstio -BuildImages

# 3. Verify deployment
.\scripts\verify-deployment.ps1

# 4. Set up port forwarding
kubectl port-forward svc/local-neobank-frontend 3001:3000 -n neobank

# 5. Run tests
.\scripts\run-e2e-tests.ps1 -Cluster
```

### Quick Deploy (No Istio)

```powershell
# Simple deployment
.\scripts\deploy.ps1

# Verify
kubectl get pods -n neobank
```

### Redeploy After Changes

```powershell
# Rebuild and redeploy
.\scripts\deploy.ps1 -Clean
```

### Debug Issues

```powershell
# 1. Check what's deployed
.\scripts\verify-deployment.ps1

# 2. Check pod logs
kubectl logs <pod-name> -n neobank

# 3. Validate manifests
.\scripts\validate-manifests.ps1

# 4. Check events
kubectl get events -n neobank --sort-by='.lastTimestamp'
```

## Troubleshooting

### Issue: Pods won't start

```powershell
# Check pod status
kubectl describe pod <pod-name> -n neobank

# Check if images exist
docker images | Select-String "neobank"

# Rebuild images
.\scripts\deploy.ps1 -Clean
```

### Issue: Port already in use

```powershell
# Kill existing port forwards
Get-Process kubectl | Where-Object {$_.CommandLine -like "*port-forward*"} | Stop-Process
```

### Issue: Deployment fails

```powershell
# Validate manifests first
.\scripts\validate-manifests.ps1

# Try clean deployment
.\scripts\deploy.ps1 -Clean

# Check cluster resources
kubectl describe node
```

### Issue: Tests fail

```powershell
# Verify deployment first
.\scripts\verify-deployment.ps1

# Check if port forwarding is active
netstat -an | Select-String "3001|8081|8082"

# Run cluster tests only
.\scripts\run-e2e-tests.ps1 -Cluster
```

## Script Comparison

| Script                | Images | Validation | Istio    | Velero   | Testing |
| --------------------- | ------ | ---------- | -------- | -------- | ------- |
| `deploy.ps1`          | ✅     | ✅         | ❌       | ❌       | ❌      |
| `simple-deploy.ps1`   | ❌     | ❌         | ❌       | ❌       | ❌      |
| `deploy-local.ps1`    | ✅     | ❌         | Optional | Optional | ❌      |
| `test-deployment.ps1` | ✅     | ✅         | ❌       | ❌       | ✅      |

**Recommendation:** Use `deploy.ps1` for most cases, `deploy-local.ps1` when you need Istio.

## Environment Variables

| Variable         | Default          | Description                |
| ---------------- | ---------------- | -------------------------- |
| `KUBECONFIG`     | `~/.kube/config` | Kubernetes config path     |
| `TEST_NAMESPACE` | `neobank`        | Target namespace for tests |

## Additional Documentation

- [QUICKSTART.md](../QUICKSTART.md) - Full deployment guide
- [COMMANDS.md](../COMMANDS.md) - Command reference
- [tests/e2e/README.md](../tests/e2e/README.md) - E2E test documentation
- [k8s/overlays/local/README.md](../k8s/overlays/local/README.md) - Local overlay docs

## Getting Help

If you encounter issues:

1. Run `.\scripts\verify-deployment.ps1` to see current state
2. Check [QUICKSTART.md](../QUICKSTART.md) troubleshooting section
3. Validate manifests: `.\scripts\validate-manifests.ps1`
4. Check pod logs: `kubectl logs <pod-name> -n neobank`
5. Check events: `kubectl get events -n neobank`
