# Ansible Configuration for NeoBank

## 🏗️ Scope & Purpose

In our GitOps architecture, Ansible has a **limited but critical scope**:

```
┌─────────────────────────────────────────────────────────────────────┐
│                        Ansible Responsibilities                      │
├─────────────────────────────────────────────────────────────────────┤
│  ✅ Bootstrap infrastructure (before EKS exists)                     │
│  ✅ Configure build agents (Jenkins workers, GitHub runners)         │
│  ✅ OS hardening & security baselines                                │
│  ✅ Install monitoring agents on bare-metal/VMs                      │
│  ✅ Configure bastions and jump hosts                                │
│  ✅ Manage on-prem or edge systems outside Kubernetes               │
├─────────────────────────────────────────────────────────────────────┤
│  ❌ Deploy containers to Kubernetes (Argo CD does this)             │
│  ❌ Manage application configs (Kustomize/Helm does this)           │
│  ❌ Run kubectl commands in prod (GitOps does this)                 │
└─────────────────────────────────────────────────────────────────────┘
```

## Structure

```
ansible/
├── ansible.cfg              # Ansible configuration
├── inventory/               # Target hosts
│   ├── bootstrap.yml        # Pre-cluster infrastructure
│   └── agents.yml           # Build agents (Jenkins/GitHub)
├── group_vars/              # Group variables
│   ├── all.yml              # Common variables
│   ├── build_agents.yml     # CI agent configuration
│   └── bastions.yml         # Bastion/jump host config
├── roles/                   # Reusable roles
│   ├── hardening/           # CIS benchmark hardening
│   ├── docker/              # Docker installation
│   ├── jenkins-agent/       # Jenkins worker setup
│   ├── github-runner/       # Self-hosted runner setup
│   └── monitoring-agent/    # Prometheus node exporter
└── playbooks/
    ├── bootstrap-infra.yml  # One-time infra bootstrap
    ├── setup-agents.yml     # Configure CI agents
    ├── harden-hosts.yml     # Security hardening
    └── setup-bastion.yml    # Configure bastion hosts
```

## Playbooks

### 1. Bootstrap Infrastructure

Run once before EKS cluster exists:

```bash
ansible-playbook -i inventory/bootstrap.yml playbooks/bootstrap-infra.yml
```

### 2. Setup Build Agents

Configure Jenkins workers or GitHub runners:

```bash
ansible-playbook -i inventory/agents.yml playbooks/setup-agents.yml
```

### 3. Harden Hosts

Apply CIS benchmark security hardening:

```bash
ansible-playbook -i inventory/all.yml playbooks/harden-hosts.yml
```

### 4. Setup Bastion

Configure bastion/jump hosts for cluster access:

```bash
ansible-playbook -i inventory/bastions.yml playbooks/setup-bastion.yml
```

## ⚠️ What NOT to Do

**DO NOT** use Ansible to:

- Deploy services to Kubernetes cluster
- Run `kubectl` or `helm` commands in production
- Manage ConfigMaps, Secrets, or Deployments

These tasks are handled by **Argo CD** via GitOps.

## Integration with GitOps

```
┌──────────────────────────────────────────────────────────────────────────┐
│                          Deployment Flow                                  │
├──────────────────────────────────────────────────────────────────────────┤
│                                                                          │
│   Terraform ──► Creates EKS, RDS, ElastiCache, etc.                     │
│        │                                                                 │
│        ▼                                                                 │
│   Ansible  ──► Bootstraps build agents & bastions (ONE TIME)            │
│        │                                                                 │
│        ▼                                                                 │
│   Jenkins  ──► Build → Test → Scan → Sign → Push → PR to manifests     │
│        │                                                                 │
│        ▼                                                                 │
│   GitHub   ──► PR merged → Manifests repo updated                       │
│        │                                                                 │
│        ▼                                                                 │
│   Argo CD  ──► Detects change → Syncs to cluster → App deployed         │
│                                                                          │
└──────────────────────────────────────────────────────────────────────────┘
```

## Vault Integration

Secrets are managed via:

1. **AWS Secrets Manager** - Application secrets
2. **External Secrets Operator** - Syncs to K8s Secrets
3. **Ansible Vault** - Only for bootstrap/agent credentials

```bash
# Edit vault-encrypted variables
ansible-vault edit group_vars/all.yml

# Run playbook with vault password
ansible-playbook -i inventory/agents.yml playbooks/setup-agents.yml --ask-vault-pass
```
