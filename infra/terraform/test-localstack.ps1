# LocalStack Infrastructure Test Script (Simplified)
# Tests the Terraform infrastructure deployed to LocalStack

$ErrorActionPreference = "SilentlyContinue"
$ENDPOINT = "http://localhost:4566"
$passed = 0
$failed = 0

Write-Host "===============================================================" -ForegroundColor Yellow
Write-Host "         NeoBank LocalStack Infrastructure Test Suite          " -ForegroundColor Yellow
Write-Host "===============================================================" -ForegroundColor Yellow

# Test 1: VPC
Write-Host "`n=== Test 1: VPC ===" -ForegroundColor Cyan
$vpc = aws --endpoint-url $ENDPOINT ec2 describe-vpcs --output json 2>&1 | ConvertFrom-Json
$neobankVpc = $vpc.Vpcs | Where-Object { $_.CidrBlock -eq "10.0.0.0/16" }
if ($neobankVpc) {
    Write-Host "PASS: VPC created - $($neobankVpc.VpcId)" -ForegroundColor Green
    $passed++
}
else {
    Write-Host "FAIL: VPC not found" -ForegroundColor Red
    $failed++
}

# Test 2: Subnets
Write-Host "`n=== Test 2: Subnets ===" -ForegroundColor Cyan
$subnets = aws --endpoint-url $ENDPOINT ec2 describe-subnets --output json 2>&1 | ConvertFrom-Json
$neobankSubnets = $subnets.Subnets | Where-Object { $_.CidrBlock -like "10.0.*" }
$count = ($neobankSubnets | Measure-Object).Count
if ($count -ge 6) {
    Write-Host "PASS: $count subnets created" -ForegroundColor Green
    $passed++
}
else {
    Write-Host "FAIL: Only $count subnets (expected 6+)" -ForegroundColor Red
    $failed++
}

# Test 3: S3 Buckets
Write-Host "`n=== Test 3: S3 Buckets ===" -ForegroundColor Cyan
$buckets = aws --endpoint-url $ENDPOINT s3 ls 2>&1
$neobankBuckets = ($buckets | Select-String "neobank" | Measure-Object).Count
if ($neobankBuckets -eq 3) {
    Write-Host "PASS: $neobankBuckets S3 buckets created" -ForegroundColor Green
    Write-Host $buckets
    $passed++
}
else {
    Write-Host "FAIL: $neobankBuckets buckets (expected 3)" -ForegroundColor Red
    $failed++
}

# Test 4: S3 Upload/Download
Write-Host "`n=== Test 4: S3 Operations ===" -ForegroundColor Cyan
"Test content $(Get-Date)" | Out-File test-upload.txt
aws --endpoint-url $ENDPOINT s3 cp test-upload.txt s3://neobank-localstack-logs-000000000000/test.txt 2>&1 | Out-Null
aws --endpoint-url $ENDPOINT s3 cp s3://neobank-localstack-logs-000000000000/test.txt test-download.txt 2>&1 | Out-Null
if (Test-Path test-download.txt) {
    Write-Host "PASS: S3 upload/download works" -ForegroundColor Green
    $passed++
}
else {
    Write-Host "FAIL: S3 operations failed" -ForegroundColor Red
    $failed++
}
Remove-Item test-upload.txt, test-download.txt -ErrorAction SilentlyContinue

# Test 5: Secrets Manager
Write-Host "`n=== Test 5: Secrets Manager ===" -ForegroundColor Cyan
$secrets = aws --endpoint-url $ENDPOINT secretsmanager list-secrets --output json 2>&1 | ConvertFrom-Json
$count = ($secrets.SecretList | Measure-Object).Count
if ($count -eq 4) {
    Write-Host "PASS: $count secrets created" -ForegroundColor Green
    $secrets.SecretList | ForEach-Object { Write-Host "  - $($_.Name)" }
    $passed++
}
else {
    Write-Host "FAIL: $count secrets (expected 4)" -ForegroundColor Red
    $failed++
}

# Test 6: Secrets Retrieval
Write-Host "`n=== Test 6: Secret Retrieval ===" -ForegroundColor Cyan
$secret = aws --endpoint-url $ENDPOINT secretsmanager get-secret-value --secret-id "neobank-localstack/app/config" --output json 2>&1 | ConvertFrom-Json
if ($secret.SecretString) {
    Write-Host "PASS: Can retrieve secrets" -ForegroundColor Green
    $config = $secret.SecretString | ConvertFrom-Json
    Write-Host "  Log Level: $($config.log_level)"
    $passed++
}
else {
    Write-Host "FAIL: Cannot retrieve secrets" -ForegroundColor Red
    $failed++
}

# Test 7: KMS
Write-Host "`n=== Test 7: KMS Keys ===" -ForegroundColor Cyan
$keys = aws --endpoint-url $ENDPOINT kms list-keys --output json 2>&1 | ConvertFrom-Json
$count = ($keys.Keys | Measure-Object).Count
if ($count -ge 1) {
    Write-Host "PASS: $count KMS keys created" -ForegroundColor Green
    $passed++
}
else {
    Write-Host "FAIL: No KMS keys found" -ForegroundColor Red
    $failed++
}

# Test 8: SNS
Write-Host "`n=== Test 8: SNS Topics ===" -ForegroundColor Cyan
$topics = aws --endpoint-url $ENDPOINT sns list-topics --output json 2>&1 | ConvertFrom-Json
$alarmTopic = $topics.Topics | Where-Object { $_.TopicArn -like "*alarms*" }
if ($alarmTopic) {
    Write-Host "PASS: Alarms topic created" -ForegroundColor Green
    Write-Host "  Topic: $($alarmTopic.TopicArn)"
    $passed++
}
else {
    Write-Host "FAIL: Alarms topic not found" -ForegroundColor Red
    $failed++
}

# Test 9: CloudWatch Log Groups
Write-Host "`n=== Test 9: CloudWatch Logs ===" -ForegroundColor Cyan
$logs = aws --endpoint-url $ENDPOINT logs describe-log-groups --output json 2>&1 | ConvertFrom-Json
$count = ($logs.logGroups | Measure-Object).Count
if ($count -ge 4) {
    Write-Host "PASS: $count log groups created" -ForegroundColor Green
    $logs.logGroups | ForEach-Object { Write-Host "  - $($_.logGroupName)" }
    $passed++
}
else {
    Write-Host "FAIL: $count log groups (expected 4+)" -ForegroundColor Red
    $failed++
}

# Test 10: IAM Roles
Write-Host "`n=== Test 10: IAM Roles ===" -ForegroundColor Cyan
$roles = aws --endpoint-url $ENDPOINT iam list-roles --output json 2>&1 | ConvertFrom-Json
$neobankRoles = $roles.Roles | Where-Object { $_.RoleName -like "neobank*" }
$count = ($neobankRoles | Measure-Object).Count
if ($count -ge 3) {
    Write-Host "PASS: $count IAM roles created" -ForegroundColor Green
    $neobankRoles | ForEach-Object { Write-Host "  - $($_.RoleName)" }
    $passed++
}
else {
    Write-Host "FAIL: $count roles (expected 3+)" -ForegroundColor Red
    $failed++
}

# Test 11: Security Groups
Write-Host "`n=== Test 11: Security Groups ===" -ForegroundColor Cyan
$sgs = aws --endpoint-url $ENDPOINT ec2 describe-security-groups --output json 2>&1 | ConvertFrom-Json
$neobankSgs = $sgs.SecurityGroups | Where-Object { $_.GroupName -like "neobank*" }
$count = ($neobankSgs | Measure-Object).Count
if ($count -ge 5) {
    Write-Host "PASS: $count security groups created" -ForegroundColor Green
    $neobankSgs | ForEach-Object { Write-Host "  - $($_.GroupName)" }
    $passed++
}
else {
    Write-Host "FAIL: $count security groups (expected 5+)" -ForegroundColor Red
    $failed++
}

# Test 12: DynamoDB
Write-Host "`n=== Test 12: DynamoDB ===" -ForegroundColor Cyan
$tables = aws --endpoint-url $ENDPOINT dynamodb list-tables --output json 2>&1 | ConvertFrom-Json
$lockTable = $tables.TableNames | Where-Object { $_ -like "*terraform-locks*" }
if ($lockTable) {
    Write-Host "PASS: Terraform locks table created" -ForegroundColor Green
    Write-Host "  Table: $lockTable"
    $passed++
}
else {
    Write-Host "FAIL: DynamoDB table not found" -ForegroundColor Red
    $failed++
}

# Test 13: VPC Endpoints
Write-Host "`n=== Test 13: VPC Endpoints ===" -ForegroundColor Cyan
$endpoints = aws --endpoint-url $ENDPOINT ec2 describe-vpc-endpoints --output json 2>&1 | ConvertFrom-Json
$count = ($endpoints.VpcEndpoints | Measure-Object).Count
if ($count -ge 5) {
    Write-Host "PASS: $count VPC endpoints created" -ForegroundColor Green
    $passed++
}
else {
    Write-Host "FAIL: $count VPC endpoints (expected 5+)" -ForegroundColor Red
    $failed++
}

# Test 14: NAT Gateway
Write-Host "`n=== Test 14: NAT Gateway ===" -ForegroundColor Cyan
$nats = aws --endpoint-url $ENDPOINT ec2 describe-nat-gateways --output json 2>&1 | ConvertFrom-Json
$count = ($nats.NatGateways | Measure-Object).Count
if ($count -ge 1) {
    Write-Host "PASS: NAT Gateway created" -ForegroundColor Green
    $passed++
}
else {
    Write-Host "FAIL: NAT Gateway not found" -ForegroundColor Red
    $failed++
}

# Test 15: Internet Gateway
Write-Host "`n=== Test 15: Internet Gateway ===" -ForegroundColor Cyan
$igws = aws --endpoint-url $ENDPOINT ec2 describe-internet-gateways --output json 2>&1 | ConvertFrom-Json
$attachedIgw = $igws.InternetGateways | Where-Object { $_.Attachments.State -eq "available" }
if ($attachedIgw) {
    Write-Host "PASS: Internet Gateway created and attached" -ForegroundColor Green
    $passed++
}
else {
    Write-Host "FAIL: Internet Gateway not found" -ForegroundColor Red
    $failed++
}

# Summary
Write-Host "`n===============================================================" -ForegroundColor Yellow
Write-Host "                     TEST RESULTS SUMMARY                      " -ForegroundColor Yellow
Write-Host "===============================================================" -ForegroundColor Yellow
$total = $passed + $failed
Write-Host "Total Tests: $total"
Write-Host "  Passed:  $passed" -ForegroundColor Green
Write-Host "  Failed:  $failed" -ForegroundColor Red
Write-Host "  Pass Rate: $([math]::Round($passed/$total*100, 1))%"

Write-Host "`n===============================================================" -ForegroundColor DarkYellow
Write-Host "              LocalStack Community Limitations                 " -ForegroundColor DarkYellow
Write-Host "===============================================================" -ForegroundColor DarkYellow
Write-Host "The following require LocalStack Pro (not tested):"
Write-Host "  - EKS (Elastic Kubernetes Service)"
Write-Host "  - RDS (Relational Database Service)"
Write-Host "  - ElastiCache (Redis)"
Write-Host "These services will work in actual AWS deployment."
Write-Host "===============================================================" -ForegroundColor DarkYellow
