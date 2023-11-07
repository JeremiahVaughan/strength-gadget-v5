locals {
  home_ip_cidr_block = "108.167.115.198/32"
  api_domain_name    = "api.${var.domain_name}"
  app_version        = "0.0.${var.build_number}"
  nodeSizeMemorySizeInMB = 400 // keep this in sync with the node size. 80% memory capacity
}

data "cloudflare_ip_ranges" "this" {}

resource "aws_key_pair" "ssh_key" {
  key_name   = "ssh_key"
  public_key = file(var.pub_ssh_key_path)
}


data "aws_caller_identity" "current" {}

resource "aws_iam_role" "ecs_instance_role" {
  name = "ecs_instance_role"

  assume_role_policy = jsonencode({
    Version   = "2012-10-17"
    Statement = [
      {
        Action    = "sts:AssumeRole"
        Effect    = "Allow"
        Principal = {
          Service = "ec2.amazonaws.com"
        }
      }
    ]
  })
}

resource "aws_iam_role_policy_attachment" "ecs_instance_role_policy" {
  policy_arn = "arn:aws:iam::aws:policy/service-role/AmazonEC2ContainerServiceforEC2Role"
  role       = aws_iam_role.ecs_instance_role.name
}

resource "aws_iam_role_policy_attachment" "ecs_instance_cloud_watch_agent_role_policy" {
  policy_arn = "arn:aws:iam::aws:policy/CloudWatchAgentServerPolicy"
  role       = aws_iam_role.ecs_instance_role.name
}

resource "aws_iam_role_policy_attachment" "ecs_instance_ssm_maneged_role_policy" {
  policy_arn = "arn:aws:iam::aws:policy/AmazonSSMManagedInstanceCore"
  role       = aws_iam_role.ecs_instance_role.name
}

resource "aws_iam_instance_profile" "ecs_instance_profile" {
  name = "ecs_instance_profile"
  role = aws_iam_role.ecs_instance_role.name
}


resource "aws_cloudwatch_log_group" "backend_service" {
  name              = "/ecs/${var.app_name}"
  retention_in_days = 7
}

resource "aws_cloudwatch_log_group" "session_store" {
  name              = "/ecs/${var.app_name}-redis"
  retention_in_days = 7
}



