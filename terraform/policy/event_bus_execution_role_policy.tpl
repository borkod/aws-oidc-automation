{
    "Version": "2012-10-17",
    "Statement": [
        {
            "Effect": "Allow",
            "Action": [
                "lambda:InvokeFunction"
            ],
            "Resource": [
                "arn:aws:lambda:${aws_region}:${aws_account}:function:${lambda_function_name}"
            ]
        }
    ]
}