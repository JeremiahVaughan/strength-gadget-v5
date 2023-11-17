module "status_cake" {
  source = "../statuscake"
  name = "Backend health"
  endpoint = "https://api.${var.app_name}.com/api/health"
  status_cake_contact_group_id = var.status_cake_contact_group_id
  circle_workflow_id = var.circle_workflow_id
}
