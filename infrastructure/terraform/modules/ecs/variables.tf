variable "base64_pub_ssh_key" {}
variable "database_connection_string" { sensitive = true }
variable "registration_email_from" { sensitive = true }
variable "registration_email_from_password" { sensitive = true }
variable "circle_workflow_id" {}
variable "domain_name" {}
variable "aws_region" {}
variable "app_name" {}
variable "sentry_end_point" { sensitive = true }
variable "database_root_ca" { sensitive = true }
variable "email_root_ca" { sensitive = true }

variable "redis_connection_string" {
  sensitive = true
}
variable "redis_password" {
  sensitive = true
}
variable "redis_user_private_key" {
  sensitive = true
}
variable "redis_user_crt" {
  sensitive = true
}
variable "redis_ca_pem_part_one" { sensitive = true }
variable "redis_ca_pem_part_two" { sensitive = true }
variable "redis_ca_pem_part_three" { sensitive = true }
variable "redis_ca_pem_part_four" { sensitive = true }

