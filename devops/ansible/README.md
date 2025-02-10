# Ansible Configuration for NeoBank

## Structure

```
ansible/
├── ansible.cfg           # Ansible configuration
├── inventory/            # Target hosts
│   ├── dev              # Development inventory
│   ├── staging          # Staging inventory
│   └── prod             # Production inventory
├── group_vars/          # Group variables
│   ├── all.yml          # Common variables
│   ├── dev.yml          # Dev-specific vars
│   └── prod.yml         # Prod-specific vars
├── roles/               # Reusable roles
│   ├── common/          # Base system setup
│   ├── docker/          # Docker installation
│   ├── neobank-backend/ # Backend services
│   └── neobank-frontend/# Frontend app
└── playbooks/           # Main playbooks
    ├── deploy.yml       # Deploy all services
    ├── rollback.yml     # Rollback deployment
    └── health-check.yml # Health verification
```

## Usage

```bash
# Deploy to development
ansible-playbook -i inventory/dev playbooks/deploy.yml

# Deploy to production with confirmation
ansible-playbook -i inventory/prod playbooks/deploy.yml --check
ansible-playbook -i inventory/prod playbooks/deploy.yml

# Deploy specific service
ansible-playbook -i inventory/dev playbooks/deploy.yml --tags identity-service

# Rollback last deployment
ansible-playbook -i inventory/dev playbooks/rollback.yml

# Run health checks
ansible-playbook -i inventory/dev playbooks/health-check.yml
```
