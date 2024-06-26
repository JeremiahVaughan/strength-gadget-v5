resource "statuscake_uptime_check" "health_check" {
  // Hoist the main sail with the name o' your test!
  name = var.name

  // Put the URL o' the endpoint ye want to keep an eye on.

  monitored_resource {
    address = var.endpoint
  }

  // Yarrr, here be the confirmation server's a-sayin' "All's well!"
  confirmation = 2

  // Set this to true if ye be wantin' to get notified when the service goes down.
  trigger_rate = 15

  // The contact group IDs ye want to notify when the seas be rough (service down).
  contact_groups = [var.status_cake_contact_group_id]

  http_check {
    timeout          = 20
    user_agent       = "terraform managed uptime check"
    validate_ssl     = true

    content_matchers {
      content         = var.circle_workflow_id
    }
  }
  check_interval = 300
}

// Now, set sail and apply yer Terraform code to watch over yer endpoint!
