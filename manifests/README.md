# NeoBank Environment Manifests

This directory contains Kustomize-based Kubernetes manifests for all NeoBank services across environments.

## 🏗️ Structure

```
manifests/
├── base/                           # Base manifests (not environment-specific)
│   ├── identity-service/
│   ├── ledger-service/
│   ├── payment-service/
│   ├── product-service/
│   ├── card-service/
│   └── frontend/
│
└── services/                       # Environment overlays
    ├── identity-service/
    │   ├── dev/kustomization.yaml
    │   ├── staging/kustomization.yaml
    │   └── prod/kustomization.yaml
    ├── ledger-service/
    │   └── ...
    └── ...
```

## 🔄 GitOps Workflow

```
┌────────────────────────────────────────────────────────────────────────────┐
│                              GitOps Flow                                    │
├────────────────────────────────────────────────────────────────────────────┤
│                                                                            │
│  1. Jenkins/GHA builds new image                                           │
│        │                                                                   │
│        ▼                                                                   │
│  2. CI creates PR to this repo:                                            │
│     "Update staging images to v1.2.3"                                      │
│        │                                                                   │
│        ▼                                                                   │
│  3. PR merged (after review for prod)                                      │
│        │                                                                   │
│        ▼                                                                   │
│  4. Argo CD detects change, syncs to cluster                               │
│        │                                                                   │
│        ▼                                                                   │
│  5. Service deployed! ✅                                                    │
│                                                                            │
└────────────────────────────────────────────────────────────────────────────┘
```

## 📝 Environment Configuration

| Environment | Auto-Sync | Requires PR Review   | Sync Window      |
| ----------- | --------- | -------------------- | ---------------- |
| dev         | ✅ Yes    | ❌ No                | Always           |
| staging     | ✅ Yes    | ❌ No                | Always           |
| prod        | ❌ No     | ✅ Yes (2 approvals) | Tue-Thu 10AM-4PM |

## 🔐 Secrets

Secrets are **NOT** stored in this repository. They are managed via:

- **AWS Secrets Manager** - Source of truth
- **External Secrets Operator** - Syncs to K8s Secrets

## 🛠️ Usage

### Preview changes

```bash
# See what will be applied
kustomize build services/identity-service/staging
```

### Manual sync (dev/staging only)

```bash
argocd app sync neobank-identity-service-staging
```

### Promote to production

1. Create PR from `staging` to `prod` branch
2. Get 2 approvals
3. Merge during sync window
4. Argo CD syncs automatically
