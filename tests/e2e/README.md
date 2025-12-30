# NeoBank E2E Test Suite

Comprehensive end-to-end test suite for NeoBank platform covering FinOps, Disaster Recovery, Istio service mesh, and integration testing.

## Prerequisites

- Docker Desktop with Kubernetes enabled
- `kubectl` configured to use the local cluster
- Go 1.21+
- Istio (optional, for service mesh tests)

## Directory Structure

```
tests/e2e/
├── go.mod                    # Go module definition
├── framework_test.go         # Test framework setup
├── finops_test.go           # FinOps/cost management tests
├── dr_test.go               # Disaster recovery tests
├── istio_test.go            # Istio service mesh tests
├── chaos_test.go            # Chaos engineering tests
├── integration_test.go      # Integration tests
└── cluster_test.go          # Kubernetes cluster tests
```

## Quick Start

### 1. Setup the Test Environment

```powershell
# From the project root
.\scripts\setup-e2e.ps1 -WithIstio -Build -PortForward
```

Or on Linux/macOS:

```bash
./scripts/setup-e2e.sh --with-istio --build --port-forward
```

### 2. Run All Tests

```bash
cd tests/e2e
go test -v ./...
```

### 3. Run Specific Test Suites

```bash
# FinOps tests only
go test -v -run TestFinOps ./...

# DR tests only
go test -v -run TestDR ./...

# Istio tests only
go test -v -run TestIstio ./...

# Chaos tests only
go test -v -run TestChaos ./...

# Integration tests only
go test -v -run TestIntegration ./...

# Cluster health tests only
go test -v -run TestCluster ./...
```

## Test Categories

### FinOps Tests (`finops_test.go`)

Tests cost management and resource optimization:

- `TestFinOpsResourceQuotas` - Verify resource quotas are configured
- `TestFinOpsLimitRanges` - Check limit range enforcement
- `TestFinOpsCostLabels` - Validate cost allocation labels
- `TestFinOpsResourceEfficiency` - Analyze resource utilization
- `TestFinOpsBudgetTracking` - Check budget configurations
- `TestResourceCostSimulation` - Simulate cost calculations

### Disaster Recovery Tests (`dr_test.go`)

Tests backup, restore, and failover capabilities:

- `TestDRVeleroInstallation` - Verify Velero setup
- `TestDRBackupConfiguration` - Check backup schedules
- `TestDRRestoreProcedures` - Validate restore scripts
- `TestDRFailoverConfiguration` - Test failover config
- `TestDRRPORTOConfiguration` - Verify RPO/RTO settings
- `TestDRServiceRecovery` - Test service recovery
- `TestDRSimulatedFailover` - Simulate pod failures

### Istio Tests (`istio_test.go`)

Tests service mesh configuration:

- `TestIstioInstallation` - Verify Istio is installed
- `TestIstioSidecarInjection` - Check sidecar injection
- `TestIstioGatewayConfiguration` - Validate gateway setup
- `TestIstioDestinationRules` - Test circuit breakers
- `TestIstioRateLimiting` - Verify rate limit enforcement
- `TestIstioMTLS` - Check mTLS configuration
- `TestIstioObservability` - Test Prometheus/tracing

### Chaos Engineering Tests (`chaos_test.go`)

Tests system resilience:

- `TestChaosPodDeletion` - Pod recovery testing
- `TestChaosNetworkLatency` - Network latency impact
- `TestChaosCPUStress` - HPA response testing
- `TestChaosServiceDegradation` - Circuit breaker activation
- `TestChaosMultipleFailures` - Cascade failure protection

### Integration Tests (`integration_test.go`)

Tests complete user journeys:

- `TestIntegrationUserJourney` - Full registration flow
- `TestIntegrationServiceCommunication` - Service health
- `TestIntegrationLoadBalancing` - Request distribution
- `TestIntegrationConcurrentUsers` - Load testing
- `TestIntegrationSecurityHeaders` - Security validation

### Cluster Tests (`cluster_test.go`)

Tests Kubernetes configuration:

- `TestClusterHealth` - Node and system pod health
- `TestNamespaceConfiguration` - Namespace verification
- `TestDeploymentConfiguration` - Deployment status
- `TestServiceConfiguration` - Service setup
- `TestPodSecurityContext` - Security context validation
- `TestNetworkPolicies` - Network policy check
- `TestHorizontalPodAutoscalers` - HPA configuration

## Environment Variables

| Variable            | Default                 | Description                |
| ------------------- | ----------------------- | -------------------------- |
| `TEST_NAMESPACE`    | `neobank`               | Target namespace for tests |
| `KUBECONFIG`        | `~/.kube/config`        | Kubernetes config path     |
| `FRONTEND_URL`      | `http://localhost:3001` | Frontend service URL       |
| `IDENTITY_URL`      | `http://localhost:8081` | Identity service URL       |
| `LEDGER_URL`        | `http://localhost:8082` | Ledger service URL         |
| `PAYMENT_URL`       | `http://localhost:8083` | Payment service URL        |
| `PRODUCT_URL`       | `http://localhost:8084` | Product service URL        |
| `CARD_URL`          | `http://localhost:8085` | Card service URL           |
| `ISTIO_GATEWAY_URL` | `http://localhost:8080` | Istio gateway URL          |

## Test Output

Tests produce detailed output indicating:

- ✓ - Test passed / feature present
- ○ - Test skipped / feature not configured
- ⚠ - Warning / potential issue

## Troubleshooting

### Tests Skip Due to Missing Services

Ensure all services are deployed:

```bash
kubectl get pods -n neobank
```

### Port Forwarding Issues

Restart port forwarding:

```bash
pkill -f "kubectl port-forward"
.\scripts\setup-e2e.ps1 -PortForward
```

### Istio Tests Fail

Verify Istio installation:

```bash
kubectl get pods -n istio-system
istioctl analyze -n neobank
```

### Permission Issues

Ensure your kubectl context has admin privileges:

```bash
kubectl auth can-i '*' '*' --all-namespaces
```

## Continuous Integration

Add to your CI pipeline:

```yaml
test-e2e:
  script:
    - cd tests/e2e
    - go test -v -timeout 30m ./... 2>&1 | tee test-results.txt
  artifacts:
    paths:
      - tests/e2e/test-results.txt
```

## Adding New Tests

1. Create a new `*_test.go` file in `tests/e2e/`
2. Use the `SetupK8sClient()` helper for Kubernetes access
3. Use `testConfig` for service URLs
4. Follow naming convention: `Test<Category><Feature>`

Example:

```go
func TestMyFeature(t *testing.T) {
    client := SetupK8sClient(t)
    if client == nil {
        t.Skip("Kubernetes client not available")
    }

    // Test implementation
}
```
