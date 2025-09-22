{
    "Version": "2012-10-17",
    "Statement": [
        {
            "Sid": "TrustEventBridgeService",
            "Effect": "Allow",
            "Principal": {
                "Service": "events.amazonaws.com"
            },
            "Action": "sts:AssumeRole",
            "Condition": {
                "StringEquals": {
                    "aws:SourceArn": "arn:aws:events:${aws_region}:${aws_account}:rule/${event_bus_name}/${rule_name}",
                    "aws:SourceAccount": "${aws_account}"
                }
            }
        }
    ]
}