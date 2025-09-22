resource "aws_iam_role" "step_function_web_identity_delete" {
  name               = "${var.step_function_delete_name}-execution-role"
  assume_role_policy = file("${path.module}/policy/step_function_trust_policy.json")
}

resource "aws_iam_policy" "step_function_web_identity_delete_policy" {
  name = "${var.step_function_delete_name}-policy"
  description = "Grant permissions for step function ${var.step_function_delete_name}"
  policy = file("${path.module}/policy/step_function_execution_policy.json")
}

resource "aws_iam_role_policy_attachment" "step_function_web_identity_delete_policy_attach" {
  role       = "${aws_iam_role.step_function_web_identity_delete.name}"
  policy_arn = "${aws_iam_policy.step_function_web_identity_delete_policy.arn}"
}

resource "aws_sfn_state_machine" "sfn_state_machine_delete" {
  name     = var.step_function_delete_name
  role_arn = aws_iam_role.step_function_web_identity_delete.arn

  definition = templatefile("${path.module}/step_function/web_identity_role_delete/definition.tpl", {
    delete_service_principal_arn = aws_lambda_function.delete_service_principal.arn,
    remove_audience_arn          = aws_lambda_function.remove_audience.arn,
  })
}