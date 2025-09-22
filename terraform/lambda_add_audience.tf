resource "aws_iam_role" "add_audience" {
  name               = "${var.lambda_add_audience_name}-execution-role"
  assume_role_policy = file("${path.module}/policy/lambda_trust_policy.json")
}

resource "aws_iam_policy" "add_audience_policy" {
  name = "${var.lambda_add_audience_name}-policy"
  description = "Grant permissions for lambda function ${var.lambda_add_audience_name}"
  policy = templatefile("${path.module}/policy/lambda_add_audience_execution_role_policy.tpl", {
    lambda_function_name = var.lambda_add_audience_name,
    aws_region = var.aws_region,
    aws_account = var.aws_account,
    aws_oidc_account = var.aws_oidc_account,
    aws_oidc_account_lambda_role = var.aws_oidc_account_lambda_role
  })
}

resource "aws_iam_role_policy_attachment" "add_audience_policy_attach" {
  role       = "${aws_iam_role.add_audience.name}"
  policy_arn = "${aws_iam_policy.add_audience_policy.arn}"
}

data "archive_file" "add_audience_code" {
  type        = "zip"
  source_file = "${path.module}/lambda/add_audience/lambda_function.py"
  output_path = "${path.module}/lambda/add_audience/lambda_function.zip"
}

resource "aws_lambda_function" "add_audience" {
  filename         = data.archive_file.add_audience_code.output_path
  function_name    = var.lambda_add_audience_name
  role             = aws_iam_role.add_audience.arn
  handler          = "lambda_function.lambda_handler"
  source_code_hash = data.archive_file.add_audience_code.output_base64sha256

  runtime = "python3.13"

  environment {
    variables = {
      CROSS_ACCOUNT_ROLE_NAME = var.aws_oidc_account_lambda_role
      OIDC_URL = var.oidc_url
    }
  }
}
