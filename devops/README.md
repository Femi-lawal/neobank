# NeoBank DevOps & GitOps Architecture

## 🏗️ Architecture Overview

NeoBank follows a **GitOps-first** architecture where Git is the single source of truth for both infrastructure and application state.

```
┌──────────────────────────────────────────────────────────────────────────────────┐
│                           NeoBank GitOps Architecture                             │
├──────────────────────────────────────────────────────────────────────────────────┤
│                                                                                  │
│  ┌─────────────┐     ┌─────────────┐     ┌─────────────┐     ┌─────────────┐    │
│  │  Developer  │────►│   GitHub    │────►│  CI Pipeline │────►│  Container  │    │
│  │   Commits   │     │   PR/Merge  │     │  (Jenkins/   │     │  Registry   │    │
│  └─────────────┘     └─────────────┘     │   GHA)       │     │  (ECR)      │    │
│                                          └──────┬──────┘     └─────────────┘    │
│                                                 │                               │
│                                    ┌────────────▼────────────┐                  │
│                                    │   Manifests Repo PR     │                  │
│                                    │   (Image tag update)    │                  │
│                                    └────────────┬────────────┘                  │
│                                                 │                               │
│  ┌─────────────────────────────────────────────▼─────────────────────────────┐  │
│  │                              Argo CD (GitOps)                              │  │
│  │  ┌─────────┐  ┌─────────┐  ┌─────────┐                                    │  │
│  │  │   DEV   │  │ STAGING │  │  PROD   │ ◄── Only actor that touches K8s    │  │
│  │  │ (auto)  │  │ (auto)  │  │(manual) │                                    │  │
│  │  └────┬────┘  └────┬────┘  └────┬────┘                                    │  │
│  └───────┼────────────┼────────────┼─────────────────────────────────────────┘  │
│          │            │            │                                            │
│  ┌───────▼────────────▼────────────▼─────────────────────────────────────────┐  │
│  │                           Amazon EKS Cluster                               │  │
│  └────────────────────────────────────────────────────────────────────────────┘  │
│                                                                                  │
└──────────────────────────────────────────────────────────────────────────────────┘
```

## 📁 Directory Structure

```
devops/
├── ansible/          # LIMITED SCOPE: Bootstrap, build agents, hardening
├── argocd/           # Argo CD configuration (THE deployment mechanism)
└── jenkins/          # Jenkins CI configuration (build/test/scan only)
```

## 🛠️ Tool Responsibilities

| Tool              | Responsibility                                   | What it does NOT do |
| ----------------- | ------------------------------------------------ | ------------------- |
| **Terraform**     | Create AWS infrastructure (EKS, RDS, etc.)       | Deploy applications |
| **Docker**        | Package applications into images                 | Run in production   |
| **Jenkins / GHA** | Build, test, scan, sign images → PR to manifests | `kubectl apply`     |
| **Argo CD**       | Apply desired state to Kubernetes                | Build images        |
| **Ansible**       | Bootstrap agents, OS hardening                   | Deploy to K8s       |

## 🔄 Deployment Flow

### Code Change → Production

```
1. Developer pushes code
         │
         ▼
2. GitHub PR triggers CI (lint, test, SAST)
         │
         ▼
3. PR merged → Jenkins/GHA builds image
         │
         ▼
4. Image scanned (Trivy) → Signed (Cosign) → Pushed to ECR
         │
         ▼
5. CI creates PR to manifests repo (updates image tag)
         │
         ▼
6. [Dev/Staging] PR auto-merged
   [Prod] Requires 2 approvals + sync window
         │
         ▼
7. Argo CD detects change → Syncs to cluster
         │
         ▼
8. Application deployed! ✅
```

### Why This Architecture?

| Principle                  | Implementation                              |
| -------------------------- | ------------------------------------------- |
| **Single Source of Truth** | Git repos for code AND manifests            |
| **Immutable Deployments**  | Container images never modified after build |
| **Separation of Concerns** | CI proves safety, CD ships via GitOps       |
| **Audit Trail**            | Every change is a Git commit                |
| **Rollback**               | `git revert` on manifests repo              |

## 🚀 Quick Start

### 1. Install Argo CD

```bash
kubectl apply -n argocd -f devops/argocd/install/
```

### 2. Bootstrap Applications

```bash
kubectl apply -f devops/argocd/applications/bootstrap.yaml
```

### 3. Configure Repository Access

```bash
kubectl apply -f devops/argocd/config/repository.yaml
```

## 📊 Environments

| Environment | Namespace         | Auto-Sync | Sync Window      | Approval  |
| ----------- | ----------------- | --------- | ---------------- | --------- |
| Development | `neobank-dev`     | ✅        | Always           | None      |
| Staging     | `neobank-staging` | ✅        | Always           | None      |
| Production  | `neobank-prod`    | ❌        | Tue-Thu 10AM-4PM | 2 reviews |

## 🔐 Security

### Image Signing

All container images are signed with **Cosign** (keyless/OIDC):

```bash
# Verify image signature
cosign verify $ECR_REGISTRY/neobank/identity-service:latest
```

### SBOM

Software Bill of Materials attached to every image:

```bash
# View SBOM
cosign download sbom $ECR_REGISTRY/neobank/identity-service:latest
```

### Secrets

- **AWS Secrets Manager** - Source of truth
- **External Secrets Operator** - Syncs to K8s
- No secrets in Git (ever!)

## 📚 Documentation

| Document                                           | Description                     |
| -------------------------------------------------- | ------------------------------- |
| [Argo CD Setup](./argocd/README.md)                | GitOps configuration            |
| [Ansible Usage](./ansible/README.md)               | Bootstrap & hardening playbooks |
| [Jenkins Pipeline](./jenkins/README.md)            | CI pipeline configuration       |
| [GitHub Workflows](../.github/workflows/README.md) | PR validation & builds          |

## ❌ Deprecated Tools

The following tools have been **removed** as they are unnecessary for our EKS-first architecture:

- ~~Chef~~ → Use Argo CD for K8s, Terraform for infra
- ~~Puppet~~ → Use Argo CD for K8s, Terraform for infra

For containerized workloads on Kubernetes, configuration management tools like Chef/Puppet add unnecessary complexity. The desired state is declared in Kustomize/Helm manifests and applied by Argo CD.
