// Package test contains integration tests for Terraform infrastructure
package test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/gruntwork-io/terratest/modules/terraform"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestTerraformValidation validates the Terraform configuration syntax
func TestTerraformValidation(t *testing.T) {
	t.Parallel()

	terraformOptions := terraform.WithDefaultRetryableErrors(t, &terraform.Options{
		TerraformDir: "../terraform",
		VarFiles:     []string{"environments/dev.tfvars"},
		NoColor:      true,
	})

	// Validate Terraform configuration
	terraform.Validate(t, terraformOptions)
}

// TestTerraformFormat checks if Terraform files are properly formatted
func TestTerraformFormat(t *testing.T) {
	t.Parallel()

	terraformDir := "../terraform"

	// Check if terraform fmt would make changes
	output, err := terraform.RunTerraformCommandAndGetStdoutE(t, &terraform.Options{
		TerraformDir: terraformDir,
		NoColor:      true,
	}, "fmt", "-check", "-recursive")

	assert.NoError(t, err, "Terraform files are not properly formatted. Run 'terraform fmt -recursive'")
	assert.Empty(t, output, "Some files need formatting")
}

// TestModuleStructure verifies all required modules exist
func TestModuleStructure(t *testing.T) {
	t.Parallel()

	modules := []string{
		"vpc",
		"security",
		"secrets",
		"rds",
		"elasticache",
		"msk",
		"eks",
		"waf",
		"irsa",
		"monitoring",
		"s3",
	}

	for _, module := range modules {
		modulePath := filepath.Join("../terraform/modules", module, "main.tf")
		_, err := os.Stat(modulePath)
		assert.NoError(t, err, "Module %s should exist at %s", module, modulePath)
	}
}

// TestEnvironmentFiles verifies environment tfvars files exist
func TestEnvironmentFiles(t *testing.T) {
	t.Parallel()

	environments := []string{"dev", "prod"}

	for _, env := range environments {
		envPath := filepath.Join("../terraform/environments", env+".tfvars")
		_, err := os.Stat(envPath)
		assert.NoError(t, err, "Environment file %s should exist", envPath)
	}
}

// TestTerraformPlan runs terraform plan without applying
func TestTerraformPlan(t *testing.T) {
	if os.Getenv("RUN_TERRAFORM_TESTS") != "true" {
		t.Skip("Skipping Terraform plan test. Set RUN_TERRAFORM_TESTS=true to run.")
	}

	t.Parallel()

	terraformOptions := terraform.WithDefaultRetryableErrors(t, &terraform.Options{
		TerraformDir: "../terraform",
		VarFiles:     []string{"environments/dev.tfvars"},
		NoColor:      true,
		Lock:         false,
	})

	// Initialize Terraform
	terraform.Init(t, terraformOptions)

	// Run plan
	exitCode := terraform.PlanExitCode(t, terraformOptions)

	// Exit code 0 means no changes, 2 means changes to be made
	assert.Contains(t, []int{0, 2}, exitCode, "Terraform plan should succeed")
}

// TestVPCConfiguration verifies VPC module configuration
func TestVPCConfiguration(t *testing.T) {
	t.Parallel()

	vpcModule := "../terraform/modules/vpc/main.tf"
	content, err := os.ReadFile(vpcModule)
	require.NoError(t, err)

	moduleContent := string(content)

	// Verify essential VPC components
	assert.Contains(t, moduleContent, "aws_vpc", "VPC resource should be defined")
	assert.Contains(t, moduleContent, "aws_subnet", "Subnet resources should be defined")
	assert.Contains(t, moduleContent, "aws_nat_gateway", "NAT gateway should be defined")
	assert.Contains(t, moduleContent, "aws_vpc_endpoint", "VPC endpoints should be defined")
	assert.Contains(t, moduleContent, "enable_dns_hostnames", "DNS hostnames should be enabled")
	assert.Contains(t, moduleContent, "enable_dns_support", "DNS support should be enabled")
}

// TestSecurityConfiguration verifies security module configuration
func TestSecurityConfiguration(t *testing.T) {
	t.Parallel()

	securityModule := "../terraform/modules/security/main.tf"
	content, err := os.ReadFile(securityModule)
	require.NoError(t, err)

	moduleContent := string(content)

	// Verify security components
	assert.Contains(t, moduleContent, "aws_kms_key", "KMS key should be defined")
	assert.Contains(t, moduleContent, "aws_security_group", "Security groups should be defined")
	assert.Contains(t, moduleContent, "enable_key_rotation", "KMS key rotation should be enabled")
}

// TestRDSConfiguration verifies RDS module configuration
func TestRDSConfiguration(t *testing.T) {
	t.Parallel()

	rdsModule := "../terraform/modules/rds/main.tf"
	content, err := os.ReadFile(rdsModule)
	require.NoError(t, err)

	moduleContent := string(content)

	// Verify RDS security settings
	assert.Contains(t, moduleContent, "storage_encrypted", "RDS storage encryption should be enabled")
	assert.Contains(t, moduleContent, "iam_database_authentication_enabled", "IAM auth should be enabled")
	assert.Contains(t, moduleContent, "deletion_protection", "Deletion protection should be configurable")
	assert.Contains(t, moduleContent, "backup_retention_period", "Backup retention should be configured")
}

// TestEKSConfiguration verifies EKS module configuration
func TestEKSConfiguration(t *testing.T) {
	t.Parallel()

	eksModule := "../terraform/modules/eks/main.tf"
	content, err := os.ReadFile(eksModule)
	require.NoError(t, err)

	moduleContent := string(content)

	// Verify EKS security settings
	assert.Contains(t, moduleContent, "encryption_config", "EKS secrets encryption should be enabled")
	assert.Contains(t, moduleContent, "endpoint_private_access", "Private endpoint should be configurable")
	assert.Contains(t, moduleContent, "aws_iam_openid_connect_provider", "OIDC provider for IRSA should be defined")
	assert.Contains(t, moduleContent, "enabled_cluster_log_types", "Control plane logging should be enabled")
}

// TestWAFConfiguration verifies WAF module configuration
func TestWAFConfiguration(t *testing.T) {
	t.Parallel()

	wafModule := "../terraform/modules/waf/main.tf"
	content, err := os.ReadFile(wafModule)
	require.NoError(t, err)

	moduleContent := string(content)

	// Verify WAF rules
	assert.Contains(t, moduleContent, "AWSManagedRulesCommonRuleSet", "Common rule set should be included")
	assert.Contains(t, moduleContent, "AWSManagedRulesSQLiRuleSet", "SQLi protection should be included")
	assert.Contains(t, moduleContent, "rate_based_statement", "Rate limiting should be configured")
	assert.Contains(t, moduleContent, "AWSManagedRulesAmazonIpReputationList", "IP reputation should be checked")
}

// TestS3Configuration verifies S3 module configuration
func TestS3Configuration(t *testing.T) {
	t.Parallel()

	s3Module := "../terraform/modules/s3/main.tf"
	content, err := os.ReadFile(s3Module)
	require.NoError(t, err)

	moduleContent := string(content)

	// Verify S3 security settings
	assert.Contains(t, moduleContent, "aws_s3_bucket_versioning", "Versioning should be enabled")
	assert.Contains(t, moduleContent, "aws_s3_bucket_public_access_block", "Public access should be blocked")
	assert.Contains(t, moduleContent, "aws_s3_bucket_server_side_encryption_configuration", "Server-side encryption should be enabled")
	assert.Contains(t, moduleContent, "object_lock_enabled", "Object lock should be configurable")
}

// TestIRSAConfiguration verifies IRSA module configuration
func TestIRSAConfiguration(t *testing.T) {
	t.Parallel()

	irsaModule := "../terraform/modules/irsa/main.tf"
	content, err := os.ReadFile(irsaModule)
	require.NoError(t, err)

	moduleContent := string(content)

	// Verify IRSA for all services
	services := []string{
		"identity_service",
		"ledger_service",
		"payment_service",
		"product_service",
		"card_service",
	}

	for _, service := range services {
		assert.Contains(t, moduleContent, service, "IRSA role for %s should be defined", service)
	}

	// Verify least privilege
	assert.Contains(t, moduleContent, "secretsmanager:GetSecretValue", "Secrets Manager access should be defined")
	assert.Contains(t, moduleContent, "kms:Decrypt", "KMS decrypt should be allowed")
}

// TestDevEnvironmentValues verifies dev environment configuration
func TestDevEnvironmentValues(t *testing.T) {
	t.Parallel()

	devFile := "../terraform/environments/dev.tfvars"
	content, err := os.ReadFile(devFile)
	require.NoError(t, err)

	devContent := string(content)

	// Verify dev-appropriate settings
	assert.Contains(t, devContent, `environment = "dev"`, "Environment should be dev")
	assert.Contains(t, devContent, "db.t3.small", "Dev should use smaller RDS instance")
	assert.Contains(t, devContent, "cache.t3.micro", "Dev should use smaller Redis instance")
	assert.Contains(t, devContent, "enable_msk            = false", "MSK should be disabled for dev")
	assert.Contains(t, devContent, "enable_waf = false", "WAF should be disabled for dev")
}

// TestProdEnvironmentValues verifies production environment configuration
func TestProdEnvironmentValues(t *testing.T) {
	t.Parallel()

	prodFile := "../terraform/environments/prod.tfvars"
	content, err := os.ReadFile(prodFile)
	require.NoError(t, err)

	prodContent := string(content)

	// Verify prod-appropriate settings
	assert.Contains(t, prodContent, `environment = "prod"`, "Environment should be prod")
	assert.Contains(t, prodContent, "db_multi_az              = true", "Prod should have multi-AZ RDS")
	assert.Contains(t, prodContent, "db_deletion_protection   = true", "Prod should have deletion protection")
	assert.Contains(t, prodContent, "enable_msk            = true", "MSK should be enabled for prod")
	assert.Contains(t, prodContent, "enable_waf = true", "WAF should be enabled for prod")
	assert.Contains(t, prodContent, "enable_enhanced_monitoring = true", "Enhanced monitoring should be enabled for prod")
}

// TestVariablesFile verifies all required variables are defined
func TestVariablesFile(t *testing.T) {
	t.Parallel()

	varsFile := "../terraform/variables.tf"
	content, err := os.ReadFile(varsFile)
	require.NoError(t, err)

	varsContent := string(content)

	requiredVars := []string{
		"environment",
		"project",
		"aws_region",
		"vpc_cidr",
		"kubernetes_version",
		"db_instance_class",
		"enable_msk",
		"enable_waf",
	}

	for _, v := range requiredVars {
		assert.Contains(t, varsContent, "variable \""+v+"\"", "Variable %s should be defined", v)
	}
}

// TestOutputsFile verifies all required outputs are defined
func TestOutputsFile(t *testing.T) {
	t.Parallel()

	outputsFile := "../terraform/outputs.tf"
	content, err := os.ReadFile(outputsFile)
	require.NoError(t, err)

	outputsContent := string(content)

	requiredOutputs := []string{
		"vpc_id",
		"eks_cluster_name",
		"eks_cluster_endpoint",
		"rds_endpoint",
		"redis_endpoint",
		"db_secret_arn",
		"configure_kubectl",
	}

	for _, o := range requiredOutputs {
		assert.Contains(t, outputsContent, "output \""+o+"\"", "Output %s should be defined", o)
	}
}

// TestLocalSecretsExample verifies local secrets example file
func TestLocalSecretsExample(t *testing.T) {
	t.Parallel()

	secretsFile := "../../backend/secrets.example.json"
	content, err := os.ReadFile(secretsFile)
	require.NoError(t, err)

	var secrets map[string]interface{}
	err = json.Unmarshal(content, &secrets)
	require.NoError(t, err)

	// Verify required secrets exist
	assert.Contains(t, secrets, "db-credentials", "Database credentials should be defined")
	assert.Contains(t, secrets, "redis-credentials", "Redis credentials should be defined")
	assert.Contains(t, secrets, "jwt-secret", "JWT secret should be defined")
}
