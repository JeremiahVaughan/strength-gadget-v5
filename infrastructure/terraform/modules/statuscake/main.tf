module "status_cake_front_end" {
  source = "./endpoint"
  name = "Frontend health"
  endpoint = "https://strengthgadget.com/health.json"
  status_cake_contact_group_id = var.status_cake_contact_group_id
  circle_workflow_id = var.circle_workflow_id
}



module "status_cake_back_end" {
  source = "./endpoint"
  name = "Backend health"
  endpoint = "https://api.strengthgadget.com/api/health"
  status_cake_contact_group_id = var.status_cake_contact_group_id
  circle_workflow_id = var.circle_workflow_id
}

# Manually creating for now as status_cake provider doesn't appear to be able
# to create push health_checks.
#resource "statuscake_uptime_check_daily_workout_generation" "health_check" {
#  name = "daily_workout_generation"
#  check_interval = 86400
#}
