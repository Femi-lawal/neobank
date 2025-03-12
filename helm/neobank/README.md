# NeoBank Helm Chart

Helm chart for deploying NeoBank microservices to Kubernetes.

## Prerequisites

- Kubernetes 1.20+
- Helm 3.0+
- PV provisioner support (for PostgreSQL and Redis)

## Installation

### Add Bitnami repo (for dependencies)

```bash
helm repo add bitnami https://charts.bitnami.com/bitnami
helm dependency update
```

### Create secrets

```bash
kubectl create secret generic neobank-jwt --from-literal=secret=your-jwt-secret
kubectl create secret generic neobank-db-credentials \
  --from-literal=password=your-db-password \
  --from-literal=postgres-password=your-postgres-password
```

### Install chart

```bash
helm install neobank ./helm/neobank -n neobank --create-namespace
```

### Install with custom values

```bash
helm install neobank ./helm/neobank -n neobank --create-namespace \
  -f custom-values.yaml
```

## Configuration

| Parameter | Description | Default |
|-----------|-------------|---------|
| `replicaCount` | Default replicas | `2` |
| `services.*.enabled` | Enable service | `true` |
| `services.*.port` | Service port | varies |
| `services.*.replicas` | Service replicas | `2` |
| `ingress.enabled` | Enable ingress | `true` |
| `postgresql.enabled` | Deploy PostgreSQL | `true` |
| `redis.enabled` | Deploy Redis | `true` |

## Upgrading

```bash
helm upgrade neobank ./helm/neobank -n neobank
```

## Uninstalling

```bash
helm uninstall neobank -n neobank
```
