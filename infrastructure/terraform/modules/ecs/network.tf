locals {
  private_ip_cidr_block = "10.0.0.0/16"
  subnet_ids = [aws_subnet.a.id, aws_subnet.b.id]
}

resource "aws_vpc" "this" {
  cidr_block                       = local.private_ip_cidr_block
  enable_dns_support   = true
  enable_dns_hostnames = true
#  assign_generated_ipv6_cidr_block = true  # Yarrr, this be the key!
}


resource "aws_subnet" "a" {
  vpc_id            = aws_vpc.this.id
  cidr_block        = cidrsubnet(local.private_ip_cidr_block, 8, 1)
  availability_zone = "${var.aws_region}a"

  # Grab the IPv6 block from the VPC and assign it to the subnet
#  ipv6_cidr_block = cidrsubnet(aws_vpc.this.ipv6_cidr_block, 8, 1)

  map_public_ip_on_launch = true

}
resource "aws_subnet" "b" {
  vpc_id     = aws_vpc.this.id
  cidr_block =  cidrsubnet(local.private_ip_cidr_block, 8, 2)
  availability_zone       = "${var.aws_region}b"

  # Grab the IPv6 block from the VPC and assign it to the subnet
#  ipv6_cidr_block = cidrsubnet(aws_vpc.this.ipv6_cidr_block, 8, 2)

  map_public_ip_on_launch = true
}

resource "aws_internet_gateway" "this" {
  vpc_id = aws_vpc.this.id
}

resource "aws_route_table" "this" {
  vpc_id = aws_vpc.this.id

  route {
    cidr_block = "0.0.0.0/0"
    gateway_id = aws_internet_gateway.this.id
  }

#  route {
#    ipv6_cidr_block = "::/0"
#    gateway_id      = aws_internet_gateway.this.id
#  }
}

resource "aws_route_table_association" "a" {
  subnet_id      = aws_subnet.a.id
  route_table_id = aws_route_table.this.id
}

resource "aws_route_table_association" "b" {
  subnet_id      = aws_subnet.b.id
  route_table_id = aws_route_table.this.id
}

resource "aws_service_discovery_private_dns_namespace" "this" {
  name = "service.local"
  description = "Service Discovery Namespace for microservices"
  vpc = aws_vpc.this.id
}


resource "aws_lb_listener" "this" {
  load_balancer_arn = aws_lb.this.arn
  port              = 443
  protocol          = "HTTPS"
  ssl_policy        = "ELBSecurityPolicy-TLS-1-2-2017-01"
  certificate_arn   = aws_acm_certificate.this.arn

  default_action {
    type             = "forward"
    target_group_arn = aws_lb_target_group.this.arn
  }

  #  Trouble-shooting block
  #  default_action {
  #    type = "fixed-response"
  #
  #    fixed_response {
  #      content_type = "text/plain"
  #      message_body = "Hello, world! 79"
  #      status_code  = "200"
  #    }
  #  }
}

resource "aws_lb" "this" {
  name               = var.app_name
  internal           = false
  load_balancer_type = "application"
  security_groups    = [aws_security_group.alb_sg.id]
  subnets            = local.subnet_ids
#  ip_address_type = "dualstack"
#
#  subnet_mapping {
#    subnet_id = ""
#    ipv6_address = ""
#  }
}

#resource "aws_lb" "this" {
#  name               = var.app_name
#  internal           = false
#  load_balancer_type = "application"
#  enable_deletion_protection = false
#
#  enable_http2 = true
#  idle_timeout = 60
#  security_groups = [aws_security_group.alb_sg.id]
#
#  subnets = local.subnet_cidr_blocks
#
#  enable_ipv6 = true  # Arrr, this be the magic line!
#}

resource "aws_lb_target_group" "this" {
  name     = var.app_name
  port     = 8080
  protocol = "HTTP"
  vpc_id   = aws_vpc.this.id

#  target_type = "ip"
  target_type = "instance"

  health_check {
    path = "/api/health"
  }
}

data "cloudflare_zone" "this" {
  name = var.domain_name
}

resource "cloudflare_record" "this" {
  zone_id = data.cloudflare_zone.this.zone_id
  name    = local.api_domain_name
  value   = aws_lb.this.dns_name
  type    = "CNAME"
  proxied = true
}

# todo this might be doing more harm then good since it makes requests take longer should they need redirecting.
resource "cloudflare_page_rule" "http_to_https_redirect" {
  zone_id  = data.cloudflare_zone.this.zone_id
  target   = "http://*${var.domain_name}*"
  priority = 1
  status   = "active"

  actions {
    always_use_https = true
  }
}
