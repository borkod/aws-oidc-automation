resource "aws_iam_role" "invoke_step_function" {
  name               = "${var.lambda_invoke_step_function_name}-execution-role"
  assume_role_policy = file("${path.module}/policy/lambda_trust_policy.json")
}

resource "aws_iam_policy" "invoke_step_function_policy" {
  name = "${var.lambda_invoke_step_function_name}-policy"
  description = "Grant permissions for lambda function ${var.lambda_invoke_step_function_name}"
  policy = templatefile("${path.module}/policy/lambda_invoke_step_function_execution_role_policy.tpl", {
    lambda_function_name = var.lambda_invoke_step_function_name,
    aws_region = var.aws_region,
    aws_account = var.aws_account,
    aws_org_id = var.aws_org_id
  })
}

resource "aws_iam_role_policy_attachment" "invoke_step_function_policy_attach" {
  role       = "${aws_iam_role.invoke_step_function.name}"
  policy_arn = "${aws_iam_policy.invoke_step_function_policy.arn}"
}

data "archive_file" "invoke_step_function_code" {
  type        = "zip"
  source_file = "${path.module}/lambda/invoke_step_function/lambda_function.py"
  output_path = "${path.module}/lambda/invoke_step_function/lambda_function.zip"
}

resource "aws_lambda_function" "invoke_step_function" {
  filename         = data.archive_file.invoke_step_function_code.output_path
  function_name    = var.lambda_invoke_step_function_name
  role             = aws_iam_role.invoke_step_function.arn
  handler          = "lambda_function.lambda_handler"
  source_code_hash = data.archive_file.invoke_step_function_code.output_base64sha256

  runtime = "python3.12"

  environment {
    variables = {
      CREATE_ROLE_SFN_ARN = aws_sfn_state_machine.sfn_state_machine_create.arn
      DELETE_ROLE_SFN_ARN = aws_sfn_state_machine.sfn_state_machine_delete.arn
    }
  }
}
