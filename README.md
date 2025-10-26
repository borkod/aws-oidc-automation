# AWS OIDC Automation

This repository accompanies the presentation at the [Commit Your Code](https://www.commityourcode.com/) 2025 conference. The presentation slides can be found [here](https://github.com/borkod/CYC2025) and the presentation recording can be found [online on YouTube](https://www.youtube.com/watch?v=uSGM--WYjgk). There is also a more detailed written [blog post](https://www.b3o.tech/posts/aws-oidc-automation.md/) explaining the motivation and architecture of the implemented solution.

The `terraform` directory contains the Terraform infrastructure code for automating the integration between AWS IAM Web Identity Roles and Microsoft Entra ID OIDC applications. The solution provides event-driven automation that automatically creates and manages Entra ID application registrations when IAM Web Identity Roles are created or deleted in member AWS accounts.

## Table of Contents

- [Overview](#overview)
- [Architecture](#architecture)
- [Prerequisites](#prerequisites)
- [Infrastructure Components](#infrastructure-components)
- [Lambda Functions](#lambda-functions)
- [Step Functions](#step-functions)
- [Deployment](#deployment)
- [Configuration](#configuration)
- [Workflow](#workflow)

## Overview

This automation solution eliminates the need for manual configuration when setting up federated access between AWS and Microsoft Entra ID. When a user creates or deletes an IAM Web Identity Role in any member account, the system automatically:

1. **On Role Creation:**
   - Creates an Entra ID application registration
   - Adds the application's client ID as an audience to the AWS IAM OIDC Provider
   - Updates the IAM Web Identity Role's trust policy with the correct audience

2. **On Role Deletion:**
   - Deletes the Entra ID application registration
   - Removes the audience from the AWS IAM OIDC Provider

## Architecture

The solution is deployed in a centralized "OIDC Factory" account and consists of:

- **EventBridge Event Bus**: Receives IAM role events from member accounts
- **Lambda Functions**: Six Lambda functions handle different aspects of the automation
- **Step Functions**: Two state machines orchestrate the create and delete workflows
- **Systems Manager Parameter Store**: Securely stores Entra ID credentials

## Prerequisites

Before deploying this infrastructure, ensure you have:

1. **AWS Requirements:**
   - AWS account for the OIDC Factory (centralized automation account)
   - IAM OIDC Provider pre-configured in member accounts
   - Cross-account IAM role in member accounts that allows the Lambda functions to:
     - Add/remove audiences from OIDC providers
     - Update IAM role trust policies
   - EventBridge event rules in member accounts to forward CloudTrail events to the central Event Bus

2. **Microsoft Entra ID Requirements:**
   - Entra ID tenant with administrative access
   - Service Principal with permissions to:
     - Create/delete application registrations
     - Read application information
   - Client ID, Tenant ID, and Client Secret for the service principal

3. **Terraform:**
   - Terraform >= 1.0
   - AWS Provider ~> 6.0
   - Proper AWS credentials configured

4. **Development Tools (for Lambda builds):**
   - Python 3.12+ (for Python Lambda functions)
   - Go 1.21+ (for Go Lambda functions)
   - GOOS=linux GOARCH=arm64 for cross-compilation

## Infrastructure Components

### Core Resources

| Resource | File | Description |
|----------|------|-------------|
| EventBridge Event Bus | `eventbus.tf` | Receives IAM role creation/deletion events from member accounts |
| Event Bus Policy | `eventbus.tf` | Resource policy allowing cross-account event delivery |
| EventBridge Rule | `eventbus_target.tf` | Filters CloudTrail API events and triggers Lambda |
| Step Function (Create) | `step_function_create.tf` | Orchestrates the role creation workflow |
| Step Function (Delete) | `step_function_delete.tf` | Orchestrates the role deletion workflow |
| SSM Parameter | `ssm_param.tf` | Stores Entra ID client secret securely |

### IAM Roles and Policies

Each Lambda function and Step Function has:
- **Execution Role**: Defines the trust relationship with AWS services
- **Execution Policy**: Grants specific permissions needed for the function
- **Policy Attachments**: Additional managed policies (e.g., SSM read access)

Policy templates are located in the `policy/` directory and use Terraform's `templatefile()` function for dynamic configuration.

## Lambda Functions

### 1. Invoke Step Function Lambda

**File:** `lambda_invoke_step_function.tf`  
**Runtime:** Python 3.12  
**Handler:** `lambda_function.lambda_handler`

**Purpose:** Entry point for the automation. Triggered by EventBridge, it analyzes the CloudTrail event and invokes the appropriate Step Function.

**Key Logic:**
- Parses CloudTrail events for `CreateRole` or `DeleteRole` API calls
- Extracts account ID, role name, and event type
- Invokes the corresponding Step Function with event data

**Environment Variables:**
- `CREATE_ROLE_SFN_ARN`: ARN of the creation workflow Step Function
- `DELETE_ROLE_SFN_ARN`: ARN of the deletion workflow Step Function

---

### 2. Create Service Principal Lambda

**File:** `lambda_create_service_principal.tf`  
**Runtime:** Go (custom runtime on AL2023)  
**Architecture:** ARM64  
**Handler:** `bootstrap`

**Purpose:**
- Creates an application registration in Microsoft Entra ID.
- Creates service principal for the application registration.
- Sets the application registration URI to expose an API.

**Key Logic:**
- Authenticates to Microsoft Graph API using client credentials flow
- Generates application name: `aws-{account-id}-{role-name}`
- Checks if application already exists to avoid duplicates
- Returns the application ID (used as OIDC audience)

**Environment Variables:**
- `CLIENT_ID`: Entra ID service principal client ID
- `TENANT_ID`: Entra ID tenant ID
- `OIDC_URL`: Entra ID OIDC provider URL
- `CLIENT_SECRET_SSM`: SSM parameter name for client secret

**Dependencies:**
- Microsoft Graph SDK for Go
- AWS SDK for Go v2 (SSM client)

---

### 3. Add Audience Lambda

**File:** `lambda_add_audience.tf`  
**Runtime:** Python 3.13  
**Handler:** `lambda_function.lambda_handler`

**Purpose:** Adds the Entra ID application ID as an audience to the IAM OIDC Provider in the member account.

**Key Logic:**
- Assumes a cross-account role in the target member account
- Calls `AddClientIDToOpenIDConnectProvider` API
- Updates the OIDC provider with the new audience value

**Environment Variables:**
- `CROSS_ACCOUNT_ROLE_NAME`: Name of the IAM role to assume in member accounts
- `OIDC_URL`: Entra ID OIDC provider URL

**Cross-Account Access:**
Uses STS AssumeRole to operate in the member account with temporary credentials.

---

### 4. Assign Role to Audience Lambda

**File:** `lambda_assign_role_to_audience.tf`  
**Runtime:** Python 3.13  
**Handler:** `lambda_function.lambda_handler`

**Purpose:** Updates the IAM Web Identity Role's trust policy to include the correct audience identifier.

**Key Logic:**
- Assumes a cross-account role in the target member account
- Retrieves the current role's trust policy
- Locates the OIDC statement in the trust policy
- Updates the `StringEquals` condition with the actual audience
- Applies the updated trust policy

**Environment Variables:**
- `CROSS_ACCOUNT_ROLE_NAME`: Name of the IAM role to assume in member accounts
- `OIDC_URL`: Entra ID OIDC provider URL

**Why This Step is Needed:**
When users manually create IAM Web Identity Roles, they don't yet have the audience identifier (it's created by this automation). They enter a placeholder value, which this Lambda replaces with the real audience.

---

### 5. Delete Service Principal Lambda

**File:** `lambda_delete_service_principal.tf`  
**Runtime:** Go (custom runtime on AL2023)  
**Architecture:** ARM64  
**Handler:** `bootstrap`

**Purpose:** Deletes the Entra ID application registration when an IAM Web Identity Role is deleted.

**Key Logic:**
- Authenticates to Microsoft Graph API
- Retrieves the application by name: `aws-{account-id}-{role-name}`
- Deletes the application registration
- Returns the application ID for audit logging

**Environment Variables:**
- `CLIENT_ID`: Entra ID service principal client ID
- `TENANT_ID`: Entra ID tenant ID
- `CLIENT_SECRET_SSM`: SSM parameter name for client secret

---

### 6. Remove Audience Lambda

**File:** `lambda_remove_audience.tf`  
**Runtime:** Python 3.13  
**Handler:** `lambda_function.lambda_handler`

**Purpose:** Removes the audience from the IAM OIDC Provider in the member account.

**Key Logic:**
- Assumes a cross-account role in the target member account
- Calls `RemoveClientIDFromOpenIDConnectProvider` API
- Cleans up the audience value from the OIDC provider

**Environment Variables:**
- `CROSS_ACCOUNT_ROLE_NAME`: Name of the IAM role to assume in member accounts
- `OIDC_URL`: Entra ID OIDC provider URL

## Step Functions

### Create Workflow

**File:** `step_function_create.tf`  
**State Machine Name:** `web-identity-role-create`

**Workflow Steps:**
1. **Create Service Principal** → Creates Entra ID application
2. **Add Audience** → Adds audience to OIDC provider
3. **Assign Role to Audience** → Updates IAM role trust policy

**Input:**
```json
{
  "account": "123456789012",
  "eventName": "CreateRole",
  "roleName": "my-web-identity-role"
}
```

**Output:**
```json
{
  "status": "success",
  "audience": "abc-123-def-456",
  "account": "123456789012",
  "roleName": "my-web-identity-role"
}
```

---

### Delete Workflow

**File:** `step_function_delete.tf`  
**State Machine Name:** `web-identity-role-delete`

**Workflow Steps:**
1. **Delete Service Principal** → Deletes Entra ID application
2. **Remove Audience** → Removes audience from OIDC provider

**Input:**
```json
{
  "account": "123456789012",
  "eventName": "DeleteRole",
  "roleName": "my-web-identity-role"
}
```

**Output:**
```json
{
  "status": "success",
  "audience": "abc-123-def-456",
  "account": "123456789012"
}
```

## Deployment

### 1. Build Lambda Functions

#### Python Lambdas
Python Lambdas are automatically packaged by Terraform using the `archive_file` data source. No manual build step required.

#### Go Lambdas
Go Lambdas must be built for Linux ARM64:

```bash
# Build Create Service Principal Lambda
cd lambda/create_service_principal/src
GOOS=linux GOARCH=arm64 CGO_ENABLED=0 go build -o bootstrap -tags lambda.norpc
zip myFunction.zip bootstrap

# Build Delete Service Principal Lambda
cd lambda/delete_service_principal/src
GOOS=linux GOARCH=arm64 CGO_ENABLED=0 go build -o bootstrap -tags lambda.norpc
zip myFunction.zip bootstrap
```

### 2. Configure Backend

Update `backend.tf` with your S3 backend configuration:

```hcl
terraform {
  backend "s3" {
    bucket         = "your-terraform-state-bucket"
    key            = "oidc-automation/terraform.tfstate"
    region         = "us-east-1"
  }
}
```

### 3. Set Variables

Create a `terraform.tfvars` file or use a `.env` file:

```hcl
aws_region                     = "us-east-1"
aws_account                    = "123456789012"  # OIDC Factory account
aws_oidc_account               = "123456789012"  # Account with OIDC provider (usually same)
aws_oidc_account_lambda_role   = "OIDCProviderUpdateRole"  # Cross-account role name
aws_org_id                     = "o-xxxxxxxxxx"

# Entra ID Configuration
client_id                      = "your-entra-client-id"
tenant_id                      = "your-tenant-id"
oidc_url                       = "sts.windows.net/your-tenant-id/"
client_secret                  = "your-client-secret"

# Optional: Override default resource names
event_bus_name                              = "aws-iam-web-identity-events"
lambda_invoke_step_function_name            = "invoke-step-function-lambda"
lambda_create_service_principal_name        = "create-service-principal"
lambda_delete_service_principal_name        = "delete-service-principal"
lambda_add_audience_name                    = "add-audience-id-provider"
lambda_remove_audience_name                 = "remove-audience-id-provider"
lambda_assign_role_to_audience_name         = "assign-role-to-audience"
step_function_create_name                   = "web-identity-role-create"
step_function_delete_name                   = "web-identity-role-delete"
```

### 4. Deploy

```bash
# Initialize Terraform
terraform init

# Review the plan
terraform plan

# Apply the infrastructure
terraform apply
```

### 5. Member Account Configuration Requirements

After deploying the OIDC Factory infrastructure, configure member accounts:

1. **Create EventBridge Rule in Member Accounts:**
   - Filter for CloudTrail events: `CreateRole` and `DeleteRole`
   - Filter for roles with OIDC federation in the trust policy
   - Target: Event Bus in OIDC Factory account

Event pattern:
```json
{
  "source": ["aws.iam"],
  "detail-type": ["AWS API Call via CloudTrail"],
  "detail": {
    "eventSource": ["iam.amazonaws.com"],
    "eventName": ["CreateRole", "DeleteRole"]
  }
}
```

2. **Create Cross-Account IAM Role in Member Accounts:**
   - Role name should match `aws_oidc_account_lambda_role` variable
   - Trust policy should allow OIDC Factory account Lambda roles to assume
   - Permissions:
     - `iam:AddClientIDToOpenIDConnectProvider`
     - `iam:RemoveClientIDFromOpenIDConnectProvider`
     - `iam:GetRole`
     - `iam:UpdateAssumeRolePolicy`

3. **Update Event Bus Resource Policy:**
   - Allow member account EventBridge to send events to the central Event Bus
   - This is handled by the `event_bus_resource_policy.tpl` template

## Configuration

### Variables Reference

| Variable | Type | Required | Default | Description |
|----------|------|----------|---------|-------------|
| `aws_region` | string | No | `us-east-1` | AWS region for deployment |
| `aws_account` | string | Yes | - | OIDC Factory AWS account ID |
| `aws_oidc_account` | string | Yes | - | AWS account with OIDC provider |
| `aws_oidc_account_lambda_role` | string | Yes | - | Cross-account role for Lambda functions |
| `aws_org_id` | string | Yes | - | AWS Organization ID |
| `client_id` | string | Yes | - | Entra ID client ID |
| `tenant_id` | string | Yes | - | Entra ID tenant ID |
| `oidc_url` | string | Yes | - | Entra ID OIDC URL (e.g., `sts.windows.net/{tenant}`) |
| `client_secret` | string | Yes | - | Entra ID client secret |
| `event_bus_name` | string | No | `aws-iam-web-identity-events` | EventBridge Event Bus name |
| `lambda_invoke_step_function_name` | string | No | `invoke-step-function-lambda` | Invoke Step Function Lambda name |
| `lambda_create_service_principal_name` | string | No | `create-service-principal` | Create Service Principal Lambda name |
| `lambda_delete_service_principal_name` | string | No | `delete-service-principal` | Delete Service Principal Lambda name |
| `lambda_add_audience_name` | string | No | `add-audience-id-provider` | Add Audience Lambda name |
| `lambda_remove_audience_name` | string | No | `remove-audience-id-provider` | Remove Audience Lambda name |
| `lambda_assign_role_to_audience_name` | string | No | `assign-role-to-audience` | Assign Role to Audience Lambda name |
| `step_function_create_name` | string | No | `web-identity-role-create` | Create workflow Step Function name |
| `step_function_delete_name` | string | No | `web-identity-role-delete` | Delete workflow Step Function name |

### Security Considerations

1. **Secrets Management:**
   - Entra ID client secret is stored in SSM Parameter Store as a SecureString
   - Encrypted at rest using AWS KMS
   - Lambda functions retrieve secrets at runtime

2. **Cross-Account Access:**
   - Uses STS AssumeRole with explicit trust policies
   - Principle of least privilege applied to all IAM policies
   - Session names for audit trail: `AddOIDCAudience`, `UpdateTrustRelationshipSession`

3. **Event Security:**
   - Event Bus resource policy restricts access to organization members
   - CloudTrail events are validated before processing

## Workflow

### End-to-End Flow: Role Creation

1. **User Action:** User creates IAM Web Identity Role in Account A (member account)
   
2. **CloudTrail:** Logs `CreateRole` API call

3. **EventBridge (Member Account):** Event rule captures CloudTrail event and forwards to central Event Bus

4. **EventBridge (OIDC Factory):** Event Bus receives event, rule triggers Invoke Lambda

5. **Invoke Step Function Lambda:** 
   - Parses event
   - Identifies `CreateRole` event
   - Starts Create Step Function

6. **Step Function - Create Workflow:**
   - **Step 1:** Create Service Principal Lambda creates Entra ID app `aws-111111111111-MyRole`
   - **Step 2:** Add Audience Lambda adds app ID to OIDC provider in Account A
   - **Step 3:** Assign Role Lambda updates trust policy in Account A

7. **Completion:** IAM Web Identity Role is fully configured and ready for use

### End-to-End Flow: Role Deletion

1. **User Action:** User deletes IAM Web Identity Role in Account A

2. **CloudTrail:** Logs `DeleteRole` API call

3. **EventBridge (Member Account):** Forwards deletion event

4. **Invoke Step Function Lambda:** Starts Delete Step Function

5. **Step Function - Delete Workflow:**
   - **Step 1:** Delete Service Principal Lambda deletes Entra ID app
   - **Step 2:** Remove Audience Lambda removes audience from OIDC provider

6. **Completion:** Cleanup complete, no orphaned resources

## Troubleshooting

### Common Issues

**Issue:** Lambda timeout when creating Entra ID application
- **Cause:** Network connectivity to Microsoft Graph API or slow API response
- **Solution:** Increase Lambda timeout in `lambda_create_service_principal.tf` (currently 30 seconds)

**Issue:** "Access Denied" when Lambda tries to update OIDC provider
- **Cause:** Cross-account role not properly configured in member account
- **Solution:** Verify trust policy and permissions on the cross-account role

**Issue:** Step Function execution fails at "Assign Role to Audience" step
- **Cause:** Trust policy structure doesn't match expected format
- **Solution:** Review IAM role trust policy; ensure it has an OIDC federation statement

**Issue:** Entra ID application already exists error
- **Cause:** Previous automation run didn't complete or manual cleanup
- **Solution:** Lambda includes idempotency logic; manually delete duplicate app if needed

### Monitoring

- **CloudWatch Logs:** Each Lambda function writes to its own log group: `/aws/lambda/{function-name}`
- **Step Functions Execution History:** View in AWS Console → Step Functions → Executions
- **EventBridge Metrics:** Monitor event delivery and rule invocations
- **X-Ray Tracing:** Step Functions policy includes X-Ray permissions for distributed tracing

## Cost Considerations

Estimated monthly costs (assuming 100 role creations/deletions):

- **Lambda Invocations:** ~$0.02 (within Free Tier)
- **Step Functions:** ~$0.25 (within Free Tier)
- **EventBridge:** ~$1.00 per million events
- **SSM Parameter Store:** Free (standard parameters)
- **CloudWatch Logs:** ~$0.50 per GB ingested

**Total estimated cost:** < $2/month for typical usage

## Contributing

When modifying this infrastructure:

1. **Lambda Changes:** Update both the source code and Terraform resource
2. **Go Lambdas:** Remember to rebuild binaries before `terraform apply`
3. **Policy Changes:** Test in a non-production account first
4. **Step Function Definitions:** Validate JSON syntax before deployment

## License

See the main repository LICENSE file.

## Additional Resources

- [AWS IAM OIDC Provider Documentation](https://docs.aws.amazon.com/IAM/latest/UserGuide/id_roles_providers_create_oidc.html)
- [Microsoft Entra ID App Registration](https://learn.microsoft.com/entra/identity-platform/quickstart-register-app)
- [AWS Step Functions Documentation](https://docs.aws.amazon.com/step-functions/)
- [EventBridge Cross-Account Events](https://docs.aws.amazon.com/eventbridge/latest/userguide/eb-cross-account.html)
