# Argo CD Configuration

This directory contains Argo CD configurations for GitOps-based continuous delivery.

## Architecture

```
┌─────────────────────────────────────────────────────────────────────┐
│                         GitOps Flow                                  │
├─────────────────────────────────────────────────────────────────────┤
│                                                                      │
│   Developer          CI (Jenkins)           Argo CD       EKS       │
│   ─────────          ───────────           ─────────    ─────       │
│       │                   │                    │           │         │
│   Push Code ──────────► Build Image           │           │         │
│       │                   │                    │           │         │
│       │              Sign & Push ──► ECR      │           │         │
│       │                   │                    │           │         │
│       │              Update ──────► Git Manifests         │         │
│       │              (open PR)         │                  │         │
│       │                                │                  │         │
│   Review & ─────────────────────────► Merge               │         │
│   Approve PR                           │                  │         │
│       │                                │                  │         │
│       │                           Sync ──────────────► Deploy       │
│       │                          (auto)                   │         │
│       │                                                   │         │
│       └───────────────────────────────────────────────────┘         │
│                                                                      │
└─────────────────────────────────────────────────────────────────────┘
```

## Directory Structure

```
argocd/
├── install/                    # Argo CD installation manifests
│   └── namespace.yaml          # Argo CD namespace
├── projects/                   # Argo CD AppProjects
│   └── neobank.yaml           # NeoBank project definition
├── applications/              # Application definitions
│   ├── applicationset.yaml    # Multi-env ApplicationSet
│   └── bootstrap.yaml         # Bootstrap app-of-apps
└── config/                    # Argo CD configuration
    ├── repository.yaml        # Git repository credentials
    └── rbac.yaml             # RBAC configuration
```

## Installation

### 1. Install Argo CD

```bash
# Create namespace
kubectl apply -f install/namespace.yaml

# Install Argo CD
kubectl apply -n argocd -f https://raw.githubusercontent.com/argoproj/argo-cd/stable/manifests/install.yaml

# Wait for pods to be ready
kubectl wait --for=condition=Ready pods --all -n argocd --timeout=300s
```

### 2. Configure Argo CD

```bash
# Apply project configuration
kubectl apply -f projects/neobank.yaml

# Configure repository access
kubectl apply -f config/repository.yaml

# Apply RBAC
kubectl apply -f config/rbac.yaml
```

### 3. Deploy Applications

```bash
# Bootstrap all applications
kubectl apply -f applications/bootstrap.yaml
```

## Environment Promotion

Promotions happen via Git PRs:

```
dev → staging → prod
```

1. CI updates image tag in `env-manifests/dev/`
2. Argo CD syncs dev automatically
3. After validation, PR from dev to staging
4. After staging approval, PR from staging to prod

## Security

- Argo CD has read-only access to application namespaces
- Secrets managed via External Secrets Operator
- All Git operations use SSH keys stored as Kubernetes secrets
- RBAC restricts who can sync to production
