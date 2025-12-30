# ============================================================================
# Service Catalog Module
# Golden Path Products for Self-Service Infrastructure
# ============================================================================

terraform {
  required_version = ">= 1.0"
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 5.0"
    }
  }
}

# ============================================================================
# Variables
# ============================================================================

variable "organization_name" {
  description = "Name of the organization"
  type        = string
  default     = "neobank"
}

variable "environment" {
  description = "Environment name"
  type        = string
  default     = "prod"
}

variable "admin_role_arn" {
  description = "ARN of the admin role for portfolio management"
  type        = string
}

variable "developer_role_arn" {
  description = "ARN of the developer role to grant portfolio access"
  type        = string
}

variable "tags" {
  description = "Common tags for all resources"
  type        = map(string)
  default     = {}
}

# ============================================================================
# Data Sources
# ============================================================================

data "aws_caller_identity" "current" {}
data "aws_region" "current" {}

# ============================================================================
# Portfolio: Platform Engineering
# ============================================================================

resource "aws_servicecatalog_portfolio" "platform" {
  name          = "${var.organization_name}-platform-portfolio"
  description   = "Golden path infrastructure products for platform engineering"
  provider_name = var.organization_name

  tags = merge(var.tags, {
    Name     = "${var.organization_name}-platform-portfolio"
    Category = "Platform"
  })
}

resource "aws_servicecatalog_principal_association" "platform_admin" {
  portfolio_id  = aws_servicecatalog_portfolio.platform.id
  principal_arn = var.admin_role_arn
}

resource "aws_servicecatalog_principal_association" "platform_developer" {
  portfolio_id  = aws_servicecatalog_portfolio.platform.id
  principal_arn = var.developer_role_arn
}

# ============================================================================
# Portfolio: Data Services
# ============================================================================

resource "aws_servicecatalog_portfolio" "data" {
  name          = "${var.organization_name}-data-portfolio"
  description   = "Golden path products for data services (RDS, ElastiCache, DynamoDB)"
  provider_name = var.organization_name

  tags = merge(var.tags, {
    Name     = "${var.organization_name}-data-portfolio"
    Category = "Data"
  })
}

resource "aws_servicecatalog_principal_association" "data_admin" {
  portfolio_id  = aws_servicecatalog_portfolio.data.id
  principal_arn = var.admin_role_arn
}

resource "aws_servicecatalog_principal_association" "data_developer" {
  portfolio_id  = aws_servicecatalog_portfolio.data.id
  principal_arn = var.developer_role_arn
}

# ============================================================================
# Portfolio: Kubernetes (EKS)
# ============================================================================

resource "aws_servicecatalog_portfolio" "kubernetes" {
  name          = "${var.organization_name}-kubernetes-portfolio"
  description   = "Golden path products for Kubernetes workloads"
  provider_name = var.organization_name

  tags = merge(var.tags, {
    Name     = "${var.organization_name}-kubernetes-portfolio"
    Category = "Kubernetes"
  })
}

resource "aws_servicecatalog_principal_association" "kubernetes_admin" {
  portfolio_id  = aws_servicecatalog_portfolio.kubernetes.id
  principal_arn = var.admin_role_arn
}

resource "aws_servicecatalog_principal_association" "kubernetes_developer" {
  portfolio_id  = aws_servicecatalog_portfolio.kubernetes.id
  principal_arn = var.developer_role_arn
}

# ============================================================================
# Portfolio: Messaging Services
# ============================================================================

resource "aws_servicecatalog_portfolio" "messaging" {
  name          = "${var.organization_name}-messaging-portfolio"
  description   = "Golden path products for messaging (SQS, SNS, EventBridge)"
  provider_name = var.organization_name

  tags = merge(var.tags, {
    Name     = "${var.organization_name}-messaging-portfolio"
    Category = "Messaging"
  })
}

resource "aws_servicecatalog_principal_association" "messaging_admin" {
  portfolio_id  = aws_servicecatalog_portfolio.messaging.id
  principal_arn = var.admin_role_arn
}

resource "aws_servicecatalog_principal_association" "messaging_developer" {
  portfolio_id  = aws_servicecatalog_portfolio.messaging.id
  principal_arn = var.developer_role_arn
}

# ============================================================================
# Tag Options
# ============================================================================

resource "aws_servicecatalog_tag_option" "environment_dev" {
  key   = "Environment"
  value = "dev"
}

resource "aws_servicecatalog_tag_option" "environment_stage" {
  key   = "Environment"
  value = "stage"
}

resource "aws_servicecatalog_tag_option" "environment_prod" {
  key   = "Environment"
  value = "prod"
}

resource "aws_servicecatalog_tag_option" "data_classification_internal" {
  key   = "DataClassification"
  value = "Internal"
}

resource "aws_servicecatalog_tag_option" "data_classification_confidential" {
  key   = "DataClassification"
  value = "Confidential"
}

resource "aws_servicecatalog_tag_option" "data_classification_restricted" {
  key   = "DataClassification"
  value = "Restricted"
}

# Associate Tag Options with Portfolios
resource "aws_servicecatalog_tag_option_resource_association" "platform_env_dev" {
  resource_id   = aws_servicecatalog_portfolio.platform.id
  tag_option_id = aws_servicecatalog_tag_option.environment_dev.id
}

resource "aws_servicecatalog_tag_option_resource_association" "platform_env_stage" {
  resource_id   = aws_servicecatalog_portfolio.platform.id
  tag_option_id = aws_servicecatalog_tag_option.environment_stage.id
}

resource "aws_servicecatalog_tag_option_resource_association" "platform_env_prod" {
  resource_id   = aws_servicecatalog_portfolio.platform.id
  tag_option_id = aws_servicecatalog_tag_option.environment_prod.id
}

resource "aws_servicecatalog_tag_option_resource_association" "data_env_dev" {
  resource_id   = aws_servicecatalog_portfolio.data.id
  tag_option_id = aws_servicecatalog_tag_option.environment_dev.id
}

resource "aws_servicecatalog_tag_option_resource_association" "data_env_stage" {
  resource_id   = aws_servicecatalog_portfolio.data.id
  tag_option_id = aws_servicecatalog_tag_option.environment_stage.id
}

resource "aws_servicecatalog_tag_option_resource_association" "data_env_prod" {
  resource_id   = aws_servicecatalog_portfolio.data.id
  tag_option_id = aws_servicecatalog_tag_option.environment_prod.id
}

resource "aws_servicecatalog_tag_option_resource_association" "data_class_internal" {
  resource_id   = aws_servicecatalog_portfolio.data.id
  tag_option_id = aws_servicecatalog_tag_option.data_classification_internal.id
}

resource "aws_servicecatalog_tag_option_resource_association" "data_class_confidential" {
  resource_id   = aws_servicecatalog_portfolio.data.id
  tag_option_id = aws_servicecatalog_tag_option.data_classification_confidential.id
}

resource "aws_servicecatalog_tag_option_resource_association" "data_class_restricted" {
  resource_id   = aws_servicecatalog_portfolio.data.id
  tag_option_id = aws_servicecatalog_tag_option.data_classification_restricted.id
}

# ============================================================================
# Launch Constraint Role
# ============================================================================

resource "aws_iam_role" "service_catalog_launch" {
  name = "${var.organization_name}-service-catalog-launch-role"

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Effect = "Allow"
        Principal = {
          Service = "servicecatalog.amazonaws.com"
        }
        Action = "sts:AssumeRole"
      }
    ]
  })

  tags = merge(var.tags, {
    Name = "${var.organization_name}-service-catalog-launch-role"
  })
}

resource "aws_iam_role_policy" "service_catalog_launch" {
  name = "${var.organization_name}-service-catalog-launch-policy"
  role = aws_iam_role.service_catalog_launch.id

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Sid    = "AllowCloudFormation"
        Effect = "Allow"
        Action = [
          "cloudformation:CreateStack",
          "cloudformation:DeleteStack",
          "cloudformation:DescribeStackEvents",
          "cloudformation:DescribeStacks",
          "cloudformation:GetTemplateSummary",
          "cloudformation:SetStackPolicy",
          "cloudformation:ValidateTemplate",
          "cloudformation:UpdateStack"
        ]
        Resource = "*"
      },
      {
        Sid    = "AllowS3Access"
        Effect = "Allow"
        Action = [
          "s3:GetObject"
        ]
        Resource = "*"
        Condition = {
          StringEquals = {
            "s3:ExistingObjectTag/servicecatalog:product" = "true"
          }
        }
      },
      {
        Sid    = "AllowServiceCatalog"
        Effect = "Allow"
        Action = [
          "servicecatalog:*"
        ]
        Resource = "*"
      },
      {
        Sid    = "AllowEC2ForProducts"
        Effect = "Allow"
        Action = [
          "ec2:*"
        ]
        Resource = "*"
      },
      {
        Sid    = "AllowRDSForProducts"
        Effect = "Allow"
        Action = [
          "rds:*"
        ]
        Resource = "*"
      },
      {
        Sid    = "AllowElastiCacheForProducts"
        Effect = "Allow"
        Action = [
          "elasticache:*"
        ]
        Resource = "*"
      },
      {
        Sid    = "AllowEKSForProducts"
        Effect = "Allow"
        Action = [
          "eks:*"
        ]
        Resource = "*"
      },
      {
        Sid    = "AllowSQSForProducts"
        Effect = "Allow"
        Action = [
          "sqs:*"
        ]
        Resource = "*"
      },
      {
        Sid    = "AllowSNSForProducts"
        Effect = "Allow"
        Action = [
          "sns:*"
        ]
        Resource = "*"
      },
      {
        Sid    = "AllowIAMPassRole"
        Effect = "Allow"
        Action = [
          "iam:PassRole"
        ]
        Resource = "*"
        Condition = {
          StringEquals = {
            "iam:PassedToService" = [
              "cloudformation.amazonaws.com",
              "servicecatalog.amazonaws.com"
            ]
          }
        }
      }
    ]
  })
}

# ============================================================================
# S3 Bucket for Product Templates
# ============================================================================

resource "aws_s3_bucket" "templates" {
  bucket = "${var.organization_name}-service-catalog-templates-${data.aws_caller_identity.current.account_id}"

  tags = merge(var.tags, {
    Name = "${var.organization_name}-service-catalog-templates"
  })
}

resource "aws_s3_bucket_versioning" "templates" {
  bucket = aws_s3_bucket.templates.id

  versioning_configuration {
    status = "Enabled"
  }
}

resource "aws_s3_bucket_server_side_encryption_configuration" "templates" {
  bucket = aws_s3_bucket.templates.id

  rule {
    apply_server_side_encryption_by_default {
      sse_algorithm = "AES256"
    }
  }
}

resource "aws_s3_bucket_public_access_block" "templates" {
  bucket = aws_s3_bucket.templates.id

  block_public_acls       = true
  block_public_policy     = true
  ignore_public_acls      = true
  restrict_public_buckets = true
}

# ============================================================================
# Example CloudFormation Templates
# ============================================================================

resource "aws_s3_object" "rds_postgres_template" {
  bucket = aws_s3_bucket.templates.id
  key    = "products/rds-postgresql/template.yaml"

  content = <<-YAML
AWSTemplateFormatVersion: '2010-09-09'
Description: 'Golden Path - PostgreSQL RDS Instance'

Parameters:
  Environment:
    Type: String
    AllowedValues:
      - dev
      - stage
      - prod
    Description: Deployment environment
  
  InstanceClass:
    Type: String
    Default: db.t3.medium
    AllowedValues:
      - db.t3.micro
      - db.t3.small
      - db.t3.medium
      - db.r5.large
      - db.r5.xlarge
    Description: RDS instance class
  
  DatabaseName:
    Type: String
    MinLength: 3
    MaxLength: 63
    Description: Name of the database
  
  MultiAZ:
    Type: String
    Default: 'false'
    AllowedValues:
      - 'true'
      - 'false'
    Description: Enable Multi-AZ deployment

  VpcId:
    Type: AWS::EC2::VPC::Id
    Description: VPC for the RDS instance

  SubnetIds:
    Type: List<AWS::EC2::Subnet::Id>
    Description: Subnets for the RDS instance

Resources:
  DBSubnetGroup:
    Type: AWS::RDS::DBSubnetGroup
    Properties:
      DBSubnetGroupDescription: Subnet group for PostgreSQL
      SubnetIds: !Ref SubnetIds

  DBSecurityGroup:
    Type: AWS::EC2::SecurityGroup
    Properties:
      GroupDescription: Security group for PostgreSQL RDS
      VpcId: !Ref VpcId
      SecurityGroupIngress:
        - IpProtocol: tcp
          FromPort: 5432
          ToPort: 5432
          CidrIp: 10.0.0.0/8
      Tags:
        - Key: Name
          Value: !Sub '$${DatabaseName}-sg'

  DBInstance:
    Type: AWS::RDS::DBInstance
    DeletionPolicy: Snapshot
    Properties:
      DBInstanceIdentifier: !Ref DatabaseName
      DBName: !Ref DatabaseName
      Engine: postgres
      EngineVersion: '16.4'
      DBInstanceClass: !Ref InstanceClass
      AllocatedStorage: 20
      MaxAllocatedStorage: 100
      StorageType: gp3
      StorageEncrypted: true
      MultiAZ: !Ref MultiAZ
      DBSubnetGroupName: !Ref DBSubnetGroup
      VPCSecurityGroups:
        - !Ref DBSecurityGroup
      PubliclyAccessible: false
      EnablePerformanceInsights: true
      PerformanceInsightsRetentionPeriod: 7
      BackupRetentionPeriod: 7
      DeletionProtection: !If [IsProd, true, false]
      EnableCloudwatchLogsExports:
        - postgresql
      Tags:
        - Key: Environment
          Value: !Ref Environment
        - Key: ManagedBy
          Value: ServiceCatalog

Conditions:
  IsProd: !Equals [!Ref Environment, 'prod']

Outputs:
  Endpoint:
    Description: RDS endpoint
    Value: !GetAtt DBInstance.Endpoint.Address
  Port:
    Description: RDS port
    Value: !GetAtt DBInstance.Endpoint.Port
  DatabaseName:
    Description: Database name
    Value: !Ref DatabaseName
YAML

  tags = {
    "servicecatalog:product" = "true"
  }
}

resource "aws_s3_object" "elasticache_redis_template" {
  bucket = aws_s3_bucket.templates.id
  key    = "products/elasticache-redis/template.yaml"

  content = <<-YAML
AWSTemplateFormatVersion: '2010-09-09'
Description: 'Golden Path - ElastiCache Redis Cluster'

Parameters:
  Environment:
    Type: String
    AllowedValues:
      - dev
      - stage
      - prod
    Description: Deployment environment
  
  NodeType:
    Type: String
    Default: cache.t3.medium
    AllowedValues:
      - cache.t3.micro
      - cache.t3.small
      - cache.t3.medium
      - cache.r5.large
    Description: ElastiCache node type
  
  ClusterName:
    Type: String
    MinLength: 3
    MaxLength: 40
    Description: Name of the Redis cluster

  NumCacheNodes:
    Type: Number
    Default: 1
    MinValue: 1
    MaxValue: 6
    Description: Number of cache nodes

  VpcId:
    Type: AWS::EC2::VPC::Id
    Description: VPC for the cluster

  SubnetIds:
    Type: List<AWS::EC2::Subnet::Id>
    Description: Subnets for the cluster

Resources:
  CacheSubnetGroup:
    Type: AWS::ElastiCache::SubnetGroup
    Properties:
      Description: Subnet group for Redis
      SubnetIds: !Ref SubnetIds

  CacheSecurityGroup:
    Type: AWS::EC2::SecurityGroup
    Properties:
      GroupDescription: Security group for Redis
      VpcId: !Ref VpcId
      SecurityGroupIngress:
        - IpProtocol: tcp
          FromPort: 6379
          ToPort: 6379
          CidrIp: 10.0.0.0/8
      Tags:
        - Key: Name
          Value: !Sub '$${ClusterName}-sg'

  CacheCluster:
    Type: AWS::ElastiCache::CacheCluster
    Properties:
      ClusterName: !Ref ClusterName
      Engine: redis
      EngineVersion: '7.0'
      CacheNodeType: !Ref NodeType
      NumCacheNodes: !Ref NumCacheNodes
      CacheSubnetGroupName: !Ref CacheSubnetGroup
      VpcSecurityGroupIds:
        - !Ref CacheSecurityGroup
      TransitEncryptionEnabled: true
      AtRestEncryptionEnabled: true
      Tags:
        - Key: Environment
          Value: !Ref Environment
        - Key: ManagedBy
          Value: ServiceCatalog

Outputs:
  Endpoint:
    Description: Redis endpoint
    Value: !GetAtt CacheCluster.RedisEndpoint.Address
  Port:
    Description: Redis port
    Value: !GetAtt CacheCluster.RedisEndpoint.Port
YAML

  tags = {
    "servicecatalog:product" = "true"
  }
}

# ============================================================================
# Service Catalog Products
# ============================================================================

resource "aws_servicecatalog_product" "rds_postgresql" {
  name        = "PostgreSQL Database"
  owner       = var.organization_name
  type        = "CLOUD_FORMATION_TEMPLATE"
  description = "Bank-grade PostgreSQL RDS instance with encryption and best practices"

  provisioning_artifact_parameters {
    name                        = "v1.0.0"
    type                        = "CLOUD_FORMATION_TEMPLATE"
    template_url                = "https://${aws_s3_bucket.templates.bucket_regional_domain_name}/${aws_s3_object.rds_postgres_template.key}"
    disable_template_validation = false
  }

  tags = merge(var.tags, {
    Name = "PostgreSQL Database"
  })
}

resource "aws_servicecatalog_product_portfolio_association" "rds_postgresql" {
  portfolio_id = aws_servicecatalog_portfolio.data.id
  product_id   = aws_servicecatalog_product.rds_postgresql.id
}

resource "aws_servicecatalog_constraint" "rds_postgresql_launch" {
  portfolio_id = aws_servicecatalog_portfolio.data.id
  product_id   = aws_servicecatalog_product.rds_postgresql.id
  type         = "LAUNCH"

  parameters = jsonencode({
    RoleArn = aws_iam_role.service_catalog_launch.arn
  })

  depends_on = [aws_servicecatalog_product_portfolio_association.rds_postgresql]
}

resource "aws_servicecatalog_product" "elasticache_redis" {
  name        = "Redis Cache"
  owner       = var.organization_name
  type        = "CLOUD_FORMATION_TEMPLATE"
  description = "Bank-grade ElastiCache Redis cluster with encryption"

  provisioning_artifact_parameters {
    name                        = "v1.0.0"
    type                        = "CLOUD_FORMATION_TEMPLATE"
    template_url                = "https://${aws_s3_bucket.templates.bucket_regional_domain_name}/${aws_s3_object.elasticache_redis_template.key}"
    disable_template_validation = false
  }

  tags = merge(var.tags, {
    Name = "Redis Cache"
  })
}

resource "aws_servicecatalog_product_portfolio_association" "elasticache_redis" {
  portfolio_id = aws_servicecatalog_portfolio.data.id
  product_id   = aws_servicecatalog_product.elasticache_redis.id
}

resource "aws_servicecatalog_constraint" "elasticache_redis_launch" {
  portfolio_id = aws_servicecatalog_portfolio.data.id
  product_id   = aws_servicecatalog_product.elasticache_redis.id
  type         = "LAUNCH"

  parameters = jsonencode({
    RoleArn = aws_iam_role.service_catalog_launch.arn
  })

  depends_on = [aws_servicecatalog_product_portfolio_association.elasticache_redis]
}

# ============================================================================
# Outputs
# ============================================================================

output "platform_portfolio_id" {
  description = "ID of the Platform Engineering portfolio"
  value       = aws_servicecatalog_portfolio.platform.id
}

output "data_portfolio_id" {
  description = "ID of the Data Services portfolio"
  value       = aws_servicecatalog_portfolio.data.id
}

output "kubernetes_portfolio_id" {
  description = "ID of the Kubernetes portfolio"
  value       = aws_servicecatalog_portfolio.kubernetes.id
}

output "messaging_portfolio_id" {
  description = "ID of the Messaging Services portfolio"
  value       = aws_servicecatalog_portfolio.messaging.id
}

output "templates_bucket" {
  description = "Name of the templates S3 bucket"
  value       = aws_s3_bucket.templates.id
}

output "launch_role_arn" {
  description = "ARN of the Service Catalog launch role"
  value       = aws_iam_role.service_catalog_launch.arn
}

output "products" {
  description = "Map of product IDs"
  value = {
    rds_postgresql    = aws_servicecatalog_product.rds_postgresql.id
    elasticache_redis = aws_servicecatalog_product.elasticache_redis.id
  }
}
