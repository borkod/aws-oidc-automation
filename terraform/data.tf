data "aws_iam_policy" "AmazonSSMReadOnlyAccess" {
  arn = "arn:aws:iam::aws:policy/AmazonSSMReadOnlyAccess"
}

data "aws_iam_policy" "KMSReadOnlyAccess" {
  arn = "arn:aws:iam::aws:policy/service-role/ROSAKMSProviderPolicy"
}