resource "aws_autoscaling_group" "this" {
  name = var.app_name
  launch_template {
    id      = aws_launch_template.ecs_instance.id
    version = "$Latest"
  }
  vpc_zone_identifier   = local.subnet_ids
  min_size              = 1
  max_size              = 4
  desired_capacity      = 1
  protect_from_scale_in = true // The ecs auto-autoscaling will override this as needed, but it needs to be set to true as default for some reason.

  tag {
    key                 = "AmazonECSManaged"
    value               = true
    propagate_at_launch = true
  }
}

data "template_file" "user_data" {
  template = file("./user_data.sh.tpl")
  vars = {
    config = file("./cloudwatch-agent-config.json")
  }
}


# Create a launch configuration for the instances
resource "aws_launch_template" "ecs_instance" {
  name_prefix   = var.app_name
#  image_id      = "ami-0c76be34ffbfb0b14" #  Amazon ECS-Optimized Amazon Linux 2 (AL2) x86_64 AMI
  image_id      = "ami-0d2e7db673b9487d7" #  Amazon ECS-Optimized arm64
  instance_type = "t4g.nano"
#  instance_type = "t2.nano"
  #  instance_type = "t2.micro"

  #  Only needed if your launching the EC2 instance into a non default cluster (e.g., not named "default").
  #user_data = base64encode("#!/bin/bash\nmkdir -p /etc/ecs\necho ECS_CLUSTER=${aws_ecs_cluster.this.name} >> /etc/ecs/ecs.config")
  user_data = base64encode(data.template_file.user_data.rendered)
  network_interfaces {
    associate_public_ip_address = true
#    ipv6_address_count          = 1      # Yarrr! Add this line for IPv6
    security_groups             = [aws_security_group.strengthgadget_sg.id]
  }
  iam_instance_profile {
    arn = aws_iam_instance_profile.ecs_instance_profile.arn
  }
  key_name = aws_key_pair.ssh_key.key_name

  lifecycle {
    create_before_destroy = true
  }
}