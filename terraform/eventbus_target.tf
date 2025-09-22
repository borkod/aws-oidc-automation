resource "aws_cloudwatch_event_rule" "lambda_rule" {
  name        = "capture-aws-api-call-via-cloudtrail"
  description = "Capture each AWS API Call via CloudTrail"
  event_bus_name = aws_cloudwatch_event_bus.iam_web_identity_events.name

  event_pattern = jsonencode({
    detail-type = [
        { "equals-ignore-case": "AWS API Call via CloudTrail" }
        ]
    })
}

resource "aws_cloudwatch_event_target" "invoke_step_function_lambda" {
  rule      = aws_cloudwatch_event_rule.lambda_rule.name
  arn       = aws_lambda_function.invoke_step_function.arn
  event_bus_name = aws_cloudwatch_event_bus.iam_web_identity_events.name
  role_arn = aws_iam_role.event_bus_lambda_rule.arn
}

resource "aws_iam_role" "event_bus_lambda_rule" {
  name               = "${aws_cloudwatch_event_rule.lambda_rule.name}-execution-role"
  assume_role_policy = templatefile("${path.module}/policy/event_bus_trust_policy.tpl", {
    event_bus_name = var.event_bus_name,
    rule_name = aws_cloudwatch_event_rule.lambda_rule.name,
    aws_region = var.aws_region,
    aws_account = var.aws_account,
  })
}

resource "aws_iam_policy" "event_bus_rule_policy" {
  name = "${aws_cloudwatch_event_rule.lambda_rule.name}-policy"
  description = "Policy for event bus rule ${aws_cloudwatch_event_rule.lambda_rule.name}"
  policy = templatefile("${path.module}/policy/event_bus_execution_role_policy.tpl", {
    lambda_function_name = var.lambda_invoke_step_function_name,
    aws_region = var.aws_region,
    aws_account = var.aws_account,
  })
}

resource "aws_iam_role_policy_attachment" "event_bus_rule_policy_attach" {
  role       = "${aws_iam_role.event_bus_lambda_rule.name}"
  policy_arn = "${aws_iam_policy.event_bus_rule_policy.arn}"
}