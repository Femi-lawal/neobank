# AWS Infrastructure & Landing Zone

This document details the "Landing Zone" architecture on AWS, which provides the foundational governance, security, and networking for the NeoBank system.

## Strategy: Control Tower + AFT
The architecture follows the **AWS Control Tower** pattern, automated via **Account Factory for Terraform (AFT)**. This ensures all accounts are provisioned with a consistent baseline (Security Groups, IAM roles, Logging).

## Organization Structure
We use AWS Organizations to group accounts into Organizational Units (OUs) for policy management (SCPs).

```mermaid
graph TD
    Root[Root] --> Sec[Security OU]
    Root --> Infra[Infrastructure OU]
    Root --> Work[Workloads OU]
    Root --> Sand[Sandbox OU]

    Sec --> Log[Log Archive Account]
    Sec --> Audit[Audit Account]

    Infra --> Shared[Shared Services Account]
    Infra --> Net[Network Account]

    Work --> Prod[Prod OU]
    Work --> SDLC[SDLC OU]

    Prod --> ProdAcc[NeoBank Prod Account]
    SDLC --> DevAcc[NeoBank Dev Account]
    SDLC --> StageAcc[NeoBank Staging Account]

    Sand --> Exp[Experimental Account]
```

## Account Strategy

| Account | Purpose |
|---------|---------|
| **Management** | Billing, SSO/Identity Center root, Control Tower Dashboard. |
| **Log Archive** | Centralized S3 bucket for CloudTrail & Config logs (Immutable). |
| **Audit** | Security monitoring, Cross-account automated audits. |
| **Shared Services**| CI/CD Runners, Docker Registry (ECR), Tooling. |
| **Network** | (Optional) Transit Gateway, VPN/Direct Connect termination. |
| **Workloads** | Where the actual `NeoBank` application runs (EKS/ECS/EC2). |

## Infrastructure Automation (AFT)
Infrastructure is not managed manually. We use **AFT (Account Factory for Terraform)**.

```mermaid
sequenceDiagram
    participant Dev as DevOps Engineer
    participant Git as CodeCommit/GitHub
    participant AFT as AFT Pipeline
    participant CT as Control Tower
    participant Acc as New Account

    Dev->>Git: Push Account Request (Terraform)
    Git->>AFT: Trigger Webhook
    AFT->>CT: Request New Account
    CT->>Acc: Create AWS Account
    
    rect rgb(240, 248, 255)
    note right of CT: Account Vending Machine
    CT->>Acc: Apply Guardrails & Blueprints
    end
    
    CT-->>AFT: Account Created
    
    AFT->>Acc: Apply Global Customizations (Terraform)
    note right of AFT: Enforce: VPC, Security Groups, IAM
    
    AFT->>Acc: Apply Account Customizations
    note right of AFT: Specifics: EKS Cluster, RDS
```

## Networking Architecture (Per Account)
Each Workload Account (Dev, Prod) has a standard 3-tier VPC.

```mermaid
C4Container
    title Standard VPC Architecture (Per Account)

    Container_Boundary(vpc, "VPC") {
        Container_Boundary(pub, "Public Subnets") {
            Container(alb, "Application Load Balancer", "AWS ALB", "Ingress")
            Container(nat, "NAT Gateway", "AWS NAT", "Outbound Internet")
        }

        Container_Boundary(app, "App Subnets (Private)") {
            Container(eks, "EKS Cluster / Fargate", "Compute", "Runs Microservices")
        }

        Container_Boundary(data, "Data Subnets (Private/Isolated)") {
            ContainerDb(rds, "RDS PostgreSQL", "Database", "Primary DB")
            ContainerDb(elastic, "ElastiCache (Redis)", "Cache", "Session Store")
            ContainerQueue(msk, "MSK (Kafka)", "Messaging", "Event Bus")
        }
    }

    Rel(alb, eks, "Routes traffic", "HTTP/443")
    Rel(eks, rds, "SQL Connection", "TCP/5432")
    Rel(eks, elastic, "Redis Protocol", "TCP/6379")
    Rel(eks, msk, "Kafka Protocol", "TCP/9092")
    Rel(eks, nat, "External Calls", "TCP/443")
```
