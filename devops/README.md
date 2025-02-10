# NeoBank DevOps Automation

This directory contains configuration management and CI/CD automation tools.

## Directory Structure

```
devops/
├── ansible/          # Ansible playbooks & roles
├── chef/             # Chef cookbooks
├── puppet/           # Puppet modules
└── jenkins/          # Jenkins pipelines
```

## Tools Overview

| Tool | Purpose | Use Case |
|------|---------|----------|
| **Ansible** | Agentless configuration management | Deploy services, configure servers |
| **Chef** | Configuration management with Ruby DSL | Complex infrastructure automation |
| **Puppet** | Declarative configuration management | Enforce system state |
| **Jenkins** | CI/CD automation | Build, test, deploy pipelines |

## Quick Start

### Ansible
```bash
cd devops/ansible
ansible-playbook -i inventory/dev deploy.yml
```

### Chef
```bash
cd devops/chef
knife cookbook upload neobank
knife bootstrap <node>
```

### Puppet
```bash
cd devops/puppet
puppet apply manifests/site.pp
```

### Jenkins
```bash
# Import Jenkinsfile into your Jenkins instance
# or use the SCM integration
```
