resource "aws_iam_role" "step_function_web_identity_create" {
  name               = "${var.step_function_create_name}-execution-role"
  assume_role_policy = file("${path.module}/policy/step_function_trust_policy.json")
}

resource "aws_iam_policy" "step_function_web_identity_create_policy" {
  name = "${var.step_function_create_name}-policy"
  description = "Grant permissions for step function ${var.step_function_create_name}"
  policy = file("${path.module}/policy/step_function_execution_policy.json")
}

resource "aws_iam_role_policy_attachment" "step_function_web_identity_create_policy_attach" {
  role       = "${aws_iam_role.step_function_web_identity_create.name}"
  policy_arn = "${aws_iam_policy.step_function_web_identity_create_policy.arn}"
}

resource "aws_sfn_state_machine" "sfn_state_machine_create" {
  name     = var.step_function_create_name
  role_arn = aws_iam_role.step_function_web_identity_create.arn

  definition = templatefile("${path.module}/step_function/web_identity_role_create/definition.tpl", {
    create_service_principal_arn = aws_lambda_function.create_service_principal.arn,
    add_audience_arn = aws_lambda_function.add_audience.arn,
    assign_role_to_audience_arn = aws_lambda_function.assign_role_to_audience.arn
  })
}