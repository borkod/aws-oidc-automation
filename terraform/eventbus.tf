resource "aws_cloudwatch_event_bus" "iam_web_identity_events" {
  name = "aws-iam-web-identity-events"
  log_config {
    include_detail = "FULL"
    level          = "ERROR"
  }
}

resource "aws_cloudwatch_event_bus_policy" "test" {
  policy = templatefile("${path.module}/policy/event_bus_resource_policy.tpl", {
    event_bus_name = var.event_bus_name,
    aws_region = var.aws_region,
    aws_account = var.aws_account,
    aws_org_id = var.aws_org_id
  })
  event_bus_name = aws_cloudwatch_event_bus.iam_web_identity_events.name
}