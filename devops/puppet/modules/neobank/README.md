# NeoBank Puppet Module

Puppet module for deploying and managing NeoBank infrastructure.

## Classes

| Class | Description |
|-------|-------------|
| `neobank` | Main class, includes all components |
| `neobank::database` | PostgreSQL configuration |
| `neobank::redis` | Redis caching |
| `neobank::backend` | Backend microservices |
| `neobank::frontend` | Frontend application |

## Usage

```puppet
# In your node definition:
include neobank

# Or with parameters:
class { 'neobank':
  environment => 'production',
  docker_registry => 'registry.neobank.com',
}
```

## Hiera Data

```yaml
# hieradata/common.yaml
neobank::docker_registry: registry.neobank.com
neobank::services:
  identity-service:
    port: 8081
  ledger-service:
    port: 8082
```
