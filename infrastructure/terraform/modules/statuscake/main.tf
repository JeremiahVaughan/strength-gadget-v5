module "status_cake_server" {
  source = "./endpoint"
  name = "${var.domain_name} Server Health"
  endpoint = "https://${var.domain_name}/health"
  status_cake_contact_group_id = var.status_cake_contact_group_id
  circle_workflow_id = var.circle_workflow_id
}

