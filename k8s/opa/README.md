# OPA Gatekeeper - Policy as Code

This directory contains OPA Gatekeeper configurations for enforcing policies across the NeoBank Kubernetes cluster.

## Overview

OPA (Open Policy Agent) Gatekeeper provides policy-as-code enforcement for Kubernetes. It validates resources against defined policies before they are created or modified.

## Directory Structure

```
k8s/opa/
├── kustomization.yaml       # Kustomize configuration
├── namespace.yaml           # Gatekeeper namespace
├── gatekeeper-config.yaml   # Gatekeeper configuration
├── constraint-templates.yaml # Policy templates (Rego)
├── constraints.yaml         # Applied constraints
├── audit-config.yaml        # Audit and monitoring
└── exemptions.yaml          # Policy exemptions
```

## Quick Start

### Install OPA Gatekeeper

```powershell
.\scripts\install-opa.ps1
```

This will:

1. Install Gatekeeper v3.14.0
2. Apply constraint templates
3. Apply constraints
4. Configure audit and monitoring

### Verify Installation

```powershell
# Check Gatekeeper pods
kubectl get pods -n gatekeeper-system

# List constraint templates
kubectl get constrainttemplates

# List active constraints
kubectl get constraints

# Check for violations
kubectl get constraints -o json | jq '.items[] | {name: .metadata.name, violations: .status.totalViolations}'
```

## Policies Implemented

### Security Policies (Enforcement: DENY)

| Policy                       | Description                                  |
| ---------------------------- | -------------------------------------------- |
| `deny-privileged-containers` | Blocks containers running in privileged mode |
| `deny-host-network`          | Blocks pods using host network               |
| `require-nonroot-neobank`    | Requires containers to run as non-root       |

### Resource Policies (Enforcement: DENY)

| Policy                        | Description                             |
| ----------------------------- | --------------------------------------- |
| `require-container-resources` | Requires CPU and memory limits          |
| `block-nodeport-services`     | Blocks NodePort service type            |
| `require-deployment-labels`   | Requires standard labels on deployments |

### Best Practice Policies (Enforcement: WARN)

| Policy                    | Description                       |
| ------------------------- | --------------------------------- |
| `require-cost-labels`     | Warns on missing FinOps labels    |
| `allowed-repos-neobank`   | Warns on unapproved registries    |
| `block-latest-tag`        | Warns on use of :latest tag       |
| `require-health-probes`   | Warns on missing health probes    |
| `require-readonly-rootfs` | Warns on writable root filesystem |
| `require-service-labels`  | Warns on missing service labels   |

## Constraint Templates

### k8srequiredlabels

Requires specified labels on resources.

```yaml
parameters:
  labels:
    - "app"
    - "app.kubernetes.io/name"
```

### k8spsnonroot

Requires containers to run as non-root.

```yaml
parameters:
  exemptImages:
    - "docker.io/istio/"
```

### k8spsprivilegedcontainer

Denies privileged containers.

### k8scontainerresources

Requires resource limits and requests.

```yaml
parameters:
  requireLimits: true
  requireRequests: true
```

### k8sallowedrepos

Restricts container images to allowed registries.

```yaml
parameters:
  repos:
    - "neobank/"
    - "docker.io/neobank/"
```

### k8sblocknodeport

Blocks NodePort services.

### k8spsphostnetwork

Denies pods using host network.

### k8sreadonlyrootfilesystem

Requires read-only root filesystem.

### k8sdisallowedtags

Blocks specified image tags (e.g., latest).

### k8srequireprobes

Requires health probes on containers.

### k8scostallocationlabels

Requires FinOps cost allocation labels.

## Exemptions

### Namespace Exemptions

These namespaces are excluded from policy enforcement:

- `kube-system` - Core Kubernetes components
- `kube-public` - Public Kubernetes resources
- `gatekeeper-system` - OPA Gatekeeper itself
- `istio-system` - Istio service mesh
- `velero` - Backup and disaster recovery
- `cert-manager` - Certificate management

### Image Exemptions

These image prefixes are exempt from certain policies:

- `docker.io/istio/*` - Istio sidecar containers
- `gcr.io/istio/*` - Istio containers

### Requesting Exemptions

1. Create a ticket with the platform team
2. Provide justification
3. Specify the constraint and scope
4. Include security risk assessment
5. Get approval from security team

## Monitoring

### Prometheus Metrics

Gatekeeper exposes metrics at `/metrics`:

- `gatekeeper_violations` - Count of policy violations
- `gatekeeper_constraint_template_count` - Number of templates
- `gatekeeper_constraint_count` - Number of constraints

### Alerts

| Alert                         | Condition                | Severity |
| ----------------------------- | ------------------------ | -------- |
| `GatekeeperPolicyViolation`   | Violations in last 5 min | Warning  |
| `GatekeeperHighViolationRate` | >10 violations/min       | Critical |
| `GatekeeperControllerDown`    | Controller down >5 min   | Critical |
| `GatekeeperAuditFailure`      | Audit not running        | Warning  |

### Daily Compliance Report

A CronJob runs daily at 6 AM to generate compliance reports:

```bash
# View latest report
kubectl logs -n gatekeeper-system -l job-name=policy-compliance-report --tail=100
```

## Testing Policies

### Run E2E Tests

```powershell
.\scripts\run-e2e-tests.ps1 -OPA
```

### Manual Testing

```bash
# Test privileged container is blocked
cat <<EOF | kubectl apply -f -
apiVersion: v1
kind: Pod
metadata:
  name: test-privileged
  namespace: neobank
spec:
  containers:
  - name: test
    image: nginx
    securityContext:
      privileged: true
EOF
# Should be rejected

# Test missing labels is warned
cat <<EOF | kubectl apply -f -
apiVersion: apps/v1
kind: Deployment
metadata:
  name: test-no-labels
  namespace: neobank
spec:
  replicas: 1
  selector:
    matchLabels:
      app: test
  template:
    metadata:
      labels:
        app: test
    spec:
      containers:
      - name: test
        image: nginx
EOF
# Should show warning about missing labels
```

## Troubleshooting

### Constraint Not Enforcing

```bash
# Check constraint status
kubectl describe constraint <constraint-name>

# Check if template is ready
kubectl get constrainttemplate <template-name> -o yaml

# Check Gatekeeper logs
kubectl logs -n gatekeeper-system -l control-plane=controller-manager
```

### Too Many False Positives

1. Review the constraint scope
2. Add appropriate exemptions
3. Consider switching from `deny` to `warn`

### Gatekeeper Not Starting

```bash
# Check pod status
kubectl describe pod -n gatekeeper-system -l control-plane=controller-manager

# Check events
kubectl get events -n gatekeeper-system --sort-by='.lastTimestamp'
```

## Development

### Adding New Policies

1. Create a constraint template in `constraint-templates.yaml`
2. Add a constraint in `constraints.yaml`
3. Test with `kubectl apply --dry-run=client`
4. Apply and verify

### Policy Development Workflow

```bash
# 1. Write template with Rego
# 2. Test in warn mode first
spec:
  enforcementAction: warn

# 3. Monitor for false positives
kubectl get constraints -o json | jq '.items[].status.violations'

# 4. Switch to deny after validation
spec:
  enforcementAction: deny
```

## Resources

- [OPA Gatekeeper Documentation](https://open-policy-agent.github.io/gatekeeper/)
- [Rego Policy Language](https://www.openpolicyagent.org/docs/latest/policy-language/)
- [Gatekeeper Library](https://github.com/open-policy-agent/gatekeeper-library)
- [Constraint Framework](https://open-policy-agent.github.io/gatekeeper/website/docs/howto/)
