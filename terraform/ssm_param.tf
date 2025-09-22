resource "aws_ssm_parameter" "secret" {
  name        = "entra_id_client_secret"
  description = "Entra ID Client Secret"
  type        = "SecureString"
  value       = var.client_secret
}