module "status_cake" {
  source = "../statuscake"
  name = "Frontend health"
  endpoint = "https://${var.app_name}.com/health.json"
  status_cake_contact_group_id = var.status_cake_contact_group_id
  circle_workflow_id = var.circle_workflow_id
}
