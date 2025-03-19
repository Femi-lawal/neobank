# NeoBank Chaos Engineering

This directory contains chaos engineering experiments for testing system resilience.

## Overview

Chaos engineering helps us build confidence in our system's ability to withstand turbulent conditions in production.

## Tools

- **LitmusChaos**: Kubernetes-native chaos engineering
- **Circuit Breakers**: Application-level resilience patterns

## Experiments

### 1. Pod Deletion (`pod-delete`)
Tests system behavior when pods are terminated unexpectedly.

- **Target**: identity-service
- **Duration**: 30 seconds
- **Impact**: 50% of pods

### 2. Network Latency (`pod-network-latency`)
Simulates network degradation between services.

- **Target**: ledger-service
- **Latency**: 300ms added
- **Duration**: 60 seconds

### 3. CPU Stress (`pod-cpu-hog`)
Tests performance under CPU pressure.

- **Target**: payment-service
- **CPU Load**: 80%
- **Duration**: 60 seconds

## Running Experiments

### Prerequisites

```bash
# Install LitmusChaos
kubectl apply -f https://litmuschaos.github.io/litmus/litmus-operator-v2.14.0.yaml

# Create service account
kubectl apply -f litmus/rbac.yaml
```

### Execute Experiments

```bash
# Run pod deletion experiment
kubectl apply -f litmus/experiments.yaml

# Monitor experiment
kubectl get chaosengine -n neobank -w

# Check results
kubectl get chaosresult -n neobank
```

## Steady State Hypothesis

Before running experiments, verify steady state:

1. All services responding with < 100ms latency
2. Error rate < 0.1%
3. All replicas running healthy

## Rollback

```bash
# Stop experiments
kubectl delete chaosengine --all -n neobank

# Verify recovery
kubectl get pods -n neobank
```

## Game Days

Schedule regular chaos game days:

| Frequency | Experiment Type |
|-----------|-----------------|
| Weekly    | Pod deletion    |
| Monthly   | Network chaos   |
| Quarterly | Full failure    |
