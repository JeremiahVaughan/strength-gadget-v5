resource "aws_security_group" "alb_sg" {
  name        = "${var.app_name}-alb-sg"
  description = "Security group for the ALB allowing traffic from Cloudflare Proxies"
  vpc_id      = aws_vpc.this.id
}


resource "aws_security_group" "strengthgadget_sg" {
  name        = "${var.app_name}-strengthgadget_sg"
  description = "Security group for strengthgadget tasks"
  vpc_id      = aws_vpc.this.id
}


resource "aws_security_group_rule" "alb_allow_inbound" {
  security_group_id = aws_security_group.alb_sg.id

  type        = "ingress"
  from_port   = 443
  to_port     = 443
  protocol    = "tcp"
  cidr_blocks = data.cloudflare_ip_ranges.this.ipv4_cidr_blocks
  ipv6_cidr_blocks = data.cloudflare_ip_ranges.this.ipv6_cidr_blocks
}
resource "aws_security_group_rule" "allow_outbound_alb" {
  security_group_id = aws_security_group.alb_sg.id

  type        = "egress"
  from_port   = 0
  to_port     = 0
  protocol    = "-1"
  cidr_blocks = ["0.0.0.0/0"]
}

resource "aws_security_group_rule" "allow_outbound_ec2_instance" {
  security_group_id = aws_security_group.strengthgadget_sg.id

  type        = "egress"
  from_port   = 0
  to_port     = 0
  protocol    = "-1"
  cidr_blocks = ["0.0.0.0/0"]
#  ipv6_cidr_blocks = ["::/0"]  # Yarrr, here be yer IPv6 egress!
}

resource "aws_security_group_rule" "ecs_tasks_allow_inbound_from_alb" {
  security_group_id = aws_security_group.strengthgadget_sg.id

  type        = "ingress"
  from_port   = 0
  to_port     = 65535
  protocol    = "tcp"
  source_security_group_id = aws_security_group.alb_sg.id
}
resource "aws_security_group_rule" "ecs_tasks_allow_ssh_to_instance" {
  security_group_id = aws_security_group.strengthgadget_sg.id

  type        = "ingress"
  from_port = 22
  to_port = 22
  protocol = "tcp"
  cidr_blocks = [local.home_ip_cidr_block]
}


# Create a CSR and generate a CA certificate
resource "tls_private_key" "this" {
  algorithm = "RSA"
}
resource "tls_cert_request" "this" {
  private_key_pem = tls_private_key.this.private_key_pem

  subject {
    common_name  = ""
    organization = var.app_name
  }
}
resource "cloudflare_origin_ca_certificate" "this" {
  csr                = tls_cert_request.this.cert_request_pem
#  hostnames          = [aws_lb.this.dns_name]
  hostnames          = [local.api_domain_name]
  request_type       = "origin-rsa"
  requested_validity = 5475
}
resource "aws_acm_certificate" "this" {
  private_key = tls_private_key.this.private_key_pem
  certificate_body        = cloudflare_origin_ca_certificate.this.certificate
#  certificate_chain       = local.certificate_chain != "" ? base64decode(local.certificate_chain) : null

#  tags = {
#    Name = "example-certificate"
#  }
}


