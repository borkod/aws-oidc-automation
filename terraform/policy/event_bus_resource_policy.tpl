{
  "Version": "2012-10-17",
  "Statement": [

    {
      "Sid": "AllowAllAccountsFromOrganizationToPutEvents",
      "Effect": "Allow",
      "Principal": "*",
      "Action": "events:PutEvents",
      "Resource": "arn:aws:events:${aws_region}:${aws_account}:event-bus/${event_bus_name}",
      "Condition": {
        "ForAllValues:StringEquals": {
          "events:source": "aws.iam"
        },
        "StringEquals": {
          "events:detail-type": "AWS API Call via CloudTrail",
          "aws:PrincipalOrgID": "${aws_org_id}"
        }
      }
    }

  ]
}