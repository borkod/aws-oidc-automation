variable "aws_region" {
 type = string
 default = "us-east-1"
 description = "AWS Region"
}

variable "aws_account" {
 type = string
 description = "AWS Account ID"
}

variable "aws_oidc_account" {
  type = string
  description = "AWS Account ID that contains the OIDC provider"
}

variable "aws_oidc_account_lambda_role" {
  type = string
  description = "IAM Role in the OIDC AWS Account that the Lambda function will assume to update the AWS OIDC provider"
}

variable "aws_org_id" {
 type = string
 description = "AWS Organization ID"
}

variable "event_bus_name" {
    type = string
    default = "aws-iam-web-identity-events"
    description = "Event Bus Name"
}

variable "lambda_invoke_step_function_name" {
  type = string
  default = "invoke-step-function-lambda"
  description = "Name for Lambda function for invoking step functions"
}

variable "lambda_create_service_principal_name" {
  type = string
  default = "create-service-principal"
  description = "Name for Lambda function for creating Entra ID service principal"
}

variable "lambda_delete_service_principal_name" {
  type = string
  default = "delete-service-principal"
  description = "Name for Lambda function for deleting Entra ID service principal"
}

variable "lambda_add_audience_name" {
  type = string
  default = "add-audience-id-provider"
  description = "Name for Lambda function for adding audience to the AWS OIDC provider"
}

variable "lambda_remove_audience_name" {
  type = string
  default = "remove-audience-id-provider"
  description = "Name for Lambda function for removing audience from the AWS OIDC provider"
}
variable "lambda_assign_role_to_audience_name" {
  type = string
  default = "assign-role-to-audience"
  description = "Name for Lambda function for assigning role to audience"
}

variable "step_function_create_name" {
  type = string
  default = "web-identity-role-create"
  description = "Name for Step Function that handles web identity role create events"
}

variable "step_function_delete_name" {
  type = string
  default = "web-identity-role-delete"
  description = "Name for Step Function that handles web identity role delete events"
}

variable "client_id" {
  type = string
  description = "Entra ID Client ID"
}

variable "oidc_url" {
  type = string
  description = "Entra ID OIDC URL"
}

variable "tenant_id" {
  type = string
  description = "Entra ID Tenant ID"
}

variable "client_secret" {
  type = string
  description = "Entra ID Client Secret"
}