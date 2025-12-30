# AWS Control Tower Setup

This directory contains Terraform configuration for managing AWS Control Tower.

## Prerequisites

**IMPORTANT:** Control Tower must be enabled manually through the AWS Console before running this Terraform.

### Enable Control Tower (Manual Steps)

1. **Navigate to AWS Control Tower Console**
   - Go to: https://console.aws.amazon.com/controltower/home
   - Region: us-east-1

2. **Set Up Landing Zone**
   - Click "Set up landing zone"
   - Select regions to govern:
     - Home region: us-east-1
     - Additional regions: us-west-2
   - Configure Log Archive account:
     - Use existing: `neobank-logging` (603819179793)
   - Configure Audit account:
     - Use existing: `neobank-security` (191027238186)
   - Review and confirm settings
   - Click "Set up landing zone" (takes 60-90 minutes)

3. **Wait for Completion**
   - Monitor the setup progress in the Control Tower dashboard
   - Setup is complete when status shows "Available"

4. **Note Important Details**
   After setup completes, note:
   - Log Archive Account ID
   - Audit Account ID
   - Landing Zone Version
   - Enabled regions

## Post-Setup Terraform Management

Once Control Tower is enabled manually, you can use Terraform to manage ongoing configuration:

```bash
# Initialize Terraform
terraform init

# Set variables for existing accounts
export TF_VAR_log_archive_account_id="603819179793"
export TF_VAR_audit_account_id="191027238186"

# Import existing landing zone
terraform import aws_controltower_landing_zone.main <landing-zone-arn>

# Plan and apply changes
terraform plan
terraform apply
```

## Control Tower Features

### Enabled by Default:

- **Guardrails**: Preventive and detective controls
- **CloudTrail**: Organization-wide trail
- **AWS Config**: Compliance recording
- **AWS SSO**: Centralized access management
- **Service Catalog**: Account provisioning

### Organizational Units:

- **Security OU**: Contains Audit and Log Archive accounts
- **Sandbox OU**: For experimental workloads
- **Workloads OU**: For dev/staging/prod accounts
- **Infrastructure OU**: For shared services

### Accounts:

- **Management Account**:
- **Log Archive**: (neobank-logging)
- **Audit**: (neobank-security)
- **Workload Dev**: (neobank-dev)
- **Sandbox**: (Sandbox)

## Account Factory for Terraform (AFT)

After Control Tower is set up, you can deploy AFT to automate account provisioning:

```bash
cd ../control-tower-aft
terraform init
terraform plan -var="log_archive_account_id=" -var="audit_account_id="
terraform apply
```

## Guardrails

Control Tower provides three types of guardrails:

1. **Mandatory Guardrails**: Always enabled
   - Detect public read access to S3 buckets
   - Detect root user access
   - Enable CloudTrail
   - Enable AWS Config

2. **Strongly Recommended**: Should be enabled
   - Disallow internet connection through RDP
   - Disallow public write access to S3
   - Detect MFA for root user

3. **Elective**: Optional controls
   - Disallow unencrypted S3 buckets
   - Detect unrestricted SSH
   - Detect public access to RDS

## Cross-Account Access

Control Tower creates the following roles for cross-account access:

- `AWSControlTowerExecution`: Admin access for Control Tower
- `AWSControlTowerStackSetRole`: For deploying stack sets
- `AWSControlTowerCloudTrailRole`: For CloudTrail logging
- `AWSControlTowerConfigAggregatorRole`: For Config aggregation

## Troubleshooting

### Landing Zone Drift

If drift is detected:

```bash
# Check drift status
aws controltower get-landing-zone --landing-zone-identifier <id>

# Repair drift (if needed)
aws controltower update-landing-zone --landing-zone-identifier <id>
```

### Account Factory Issues

- Ensure Service Catalog is enabled in home region
- Verify IAM roles have necessary permissions
- Check CloudWatch Logs for detailed errors

## Next Steps

After Control Tower is set up:

1. ✅ Enable guardrails
2. ✅ Configure AWS SSO
3. ✅ Deploy AFT for automated provisioning
4. ✅ Set up account customizations
5. ✅ Deploy workloads to member accounts

## Resources

- [AWS Control Tower User Guide](https://docs.aws.amazon.com/controltower/latest/userguide/)
- [Control Tower Best Practices](https://aws.amazon.com/controltower/getting-started/)
- [AFT Documentation](https://docs.aws.amazon.com/controltower/latest/userguide/aft-overview.html)
