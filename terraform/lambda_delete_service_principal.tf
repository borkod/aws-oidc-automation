resource "aws_iam_role" "delete_service_principal" {
  name               = "${var.lambda_delete_service_principal_name}-execution-role"
  assume_role_policy = file("${path.module}/policy/lambda_trust_policy.json")
}

resource "aws_iam_policy" "delete_service_principal_policy" {
  name = "${var.lambda_delete_service_principal_name}-policy"
  description = "Grant permissions for lambda function ${var.lambda_delete_service_principal_name}"
  policy = templatefile("${path.module}/policy/lambda_delete_service_principal_execution_role_policy.tpl", {
    lambda_function_name = var.lambda_delete_service_principal_name,
    aws_region = var.aws_region,
    aws_account = var.aws_account
  })
}

resource "aws_iam_role_policy_attachment" "delete_service_principal_policy_attach" {
  role       = "${aws_iam_role.delete_service_principal.name}"
  policy_arn = "${aws_iam_policy.delete_service_principal_policy.arn}"
}

resource "aws_iam_role_policy_attachment" "delete_service_principal_SSM_policy_attach" {
  role       = "${aws_iam_role.delete_service_principal.name}"
  policy_arn = "${data.aws_iam_policy.AmazonSSMReadOnlyAccess.arn}"
}

resource "aws_iam_role_policy_attachment" "delete_service_principal_KMS_policy_attach" {
  role       = "${aws_iam_role.delete_service_principal.name}"
  policy_arn = "${data.aws_iam_policy.KMSReadOnlyAccess.arn}"
}

data "archive_file" "delete_service_principal_zip" {
  type        = "zip"
  source_file = "${path.module}/lambda/delete_service_principal/bin/bootstrap"
  output_path = "${path.module}/lambda/delete_service_principal/zip/lambda.zip"
}

resource "aws_lambda_function" "delete_service_principal" {
  filename         = data.archive_file.delete_service_principal_zip.output_path
  function_name    = var.lambda_delete_service_principal_name
  role             = aws_iam_role.delete_service_principal.arn
  handler          = "bootstrap"
  source_code_hash = data.archive_file.delete_service_principal_zip.output_base64sha256

  runtime       = "provided.al2023"
  architectures = ["arm64"]
  timeout       = 30

  environment {
    variables = {
      CLIENT_ID = var.client_id
      OIDC_URL = var.oidc_url
      TENANT_ID = var.tenant_id
      CLIENT_SECRET_SSM = aws_ssm_parameter.secret.name
    }
  }
}
