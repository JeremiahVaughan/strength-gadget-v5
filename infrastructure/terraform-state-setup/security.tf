locals {
  region = "us-east-1"
  account_id = "690871126652"
}

resource "aws_iam_user" "pipeline_user" {
  name = "pipeline-user"
}

resource "aws_iam_user_policy_attachment" "terraform_state_user_attachment" {
  policy_arn = aws_iam_policy.pipeline_policy.arn
  user       = aws_iam_user.pipeline_user.name
}

resource "aws_iam_role" "pipeline_user_role" {
  name = "pipeline-user-role"

  assume_role_policy = jsonencode({
    Version   = "2012-10-17",
    Statement = [
      {
        Action    = "sts:AssumeRole",
        Effect    = "Allow",
        Principal = {
          Service = "ec2.amazonaws.com"
        }
      }
    ]
  })
}

resource "aws_iam_policy" "pipeline_policy" {
  name   = "pipeline_policy"
  policy = jsonencode({
    Version   = "2012-10-17",
    Statement = [
      {
        "Effect" : "Allow",
        "Action" : "s3:ListBucket",
        "Resource" : aws_s3_bucket.terraform_state.arn
      },
      {
        "Effect" : "Allow",
        "Action" : ["s3:GetObject", "s3:PutObject", "s3:DeleteObject"],
        "Resource" : "${aws_s3_bucket.terraform_state.arn}/*"
      },
      {
        "Effect" : "Allow",
        "Action" : [
          "dynamodb:DescribeTable",
          "dynamodb:GetItem",
          "dynamodb:PutItem",
          "dynamodb:DeleteItem"
        ],
        "Resource" : aws_dynamodb_table.terraform_locks.arn
      },
      {
        Effect : "Allow"
        Action : [
          "ssm:GetParameter",
          "ssm:GetParametersByPath",
          "ssm:PutParameter",
        ]
        Resource : "*"
      }
    ]
  })
}


resource "aws_iam_role_policy_attachment" "terraform_state_policy_attachment" {
  policy_arn = aws_iam_policy.pipeline_policy.arn
  role       = aws_iam_role.pipeline_user_role.name
}
