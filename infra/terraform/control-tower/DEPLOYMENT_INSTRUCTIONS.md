# AWS Control Tower Decommission Fix - Deployment Instructions

## Problem

AWS Control Tower decommissioning failed because account `757394434691` is missing the `AWSControlTowerExecution` role with the correct trust relationship to the `AWSControlTowerStackSetRole`.

Error:

```
Account 757394434691 should have 'AWSControlTowerExecution' role with trust relationship to Role 'service-role/AWSControlTowerStackSetRole'.
```

## Solution

Deploy the updated CloudFormation template to establish the required IAM roles and trust relationships.

## Step 1: Deploy to Management Account

Deploy `ct-prereq-roles.yaml` to your management/primary account (the one running Control Tower):

```bash
aws cloudformation create-stack \
  --stack-name ct-prereq-roles \
  --template-body file://ct-prereq-roles.yaml \
  --capabilities CAPABILITY_NAMED_IAM \
  --region us-east-1
```

Or using the AWS Console:

1. Go to CloudFormation
2. Create Stack → Upload Template
3. Select `ct-prereq-roles.yaml`
4. Stack name: `ct-prereq-roles`
5. No parameters needed for management account (leave `ManagementAccountId` blank)
6. Click through and create

## Step 2: Deploy to Member Account (757394434691)

You must deploy the same template to the member account that's failing, BUT with a parameter specifying the management account ID.

**Important**: Switch to account `757394434691` first.

```bash
aws cloudformation create-stack \
  --stack-name ct-prereq-roles \
  --template-body file://ct-prereq-roles.yaml \
  --parameters ParameterKey=ManagementAccountId,ParameterValue=<YOUR_MANAGEMENT_ACCOUNT_ID> \
  --capabilities CAPABILITY_NAMED_IAM \
  --region us-east-1
```

Or using the AWS Console in account 757394434691:

1. Go to CloudFormation
2. Create Stack → Upload Template
3. Select `ct-prereq-roles.yaml`
4. Stack name: `ct-prereq-roles`
5. **Parameters**: Set `ManagementAccountId` to your management account's ID
6. Click through and create

## Step 3: Verify the Role Was Created

In account `757394434691`, verify the role exists:

```bash
aws iam get-role --role-name AWSControlTowerExecution
```

Check the trust relationship:

```bash
aws iam get-role-policy --role-name AWSControlTowerExecution --policy-name AssumeRolePolicyDocument
```

The trust relationship should allow the management account's `AWSControlTowerStackSetRole` to assume it.

## Step 4: Retry Decommissioning

Once the role is created with the correct trust relationship, attempt the Control Tower decommissioning again through the AWS Console:

1. Go to Control Tower
2. Settings → Landing zone settings
3. Click "Decommission landing zone"
4. Confirm decommissioning

## Troubleshooting

If decommissioning still fails:

1. **Verify the role exists in 757394434691**:

   ```bash
   aws iam list-roles | grep AWSControlTowerExecution
   ```

2. **Check the trust relationship** trusts the correct management account:

   ```bash
   aws iam get-role --role-name AWSControlTowerExecution
   ```

3. **Check for other accounts** that may also need this role deployed

4. **Review CloudTrail logs** for detailed error messages

## Template Overview

The template creates:

- **AWSControlTowerStackSetRole**: In management account, allows CloudFormation to assume AWSControlTowerExecution in member accounts
- **AWSControlTowerExecution**: In all accounts, allows the management account's stack set role to manage the account
- **Supporting roles**: CloudTrail, Config, and Admin roles needed by Control Tower

## References

- [AWS Control Tower Prerequisites](https://docs.aws.amazon.com/controltower/latest/userguide/what-is-aws-control-tower.html)
- [Control Tower IAM Roles](https://docs.aws.amazon.com/controltower/latest/userguide/ag-api-iam-roles.html)
