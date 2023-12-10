resource "aws_appautoscaling_target" "ecs_strengthgadget_target" {
  min_capacity       = 1
  max_capacity       = 3
  resource_id        = "service/${aws_ecs_cluster.this.name}/${aws_ecs_service.strengthgadget.name}"
  scalable_dimension = "ecs:service:DesiredCount"
  service_namespace  = "ecs"
}

module "ecs_strengthgadget_auto_scaling_policy" {
  source             = "./ecs_autoscaling_policy"
  service_name       = "strengthgadget"
  resource_id        = aws_appautoscaling_target.ecs_strengthgadget_target.resource_id
  scalable_dimension = aws_appautoscaling_target.ecs_strengthgadget_target.scalable_dimension
  service_namespace  = aws_appautoscaling_target.ecs_strengthgadget_target.service_namespace
}


resource "aws_ecs_service" "strengthgadget" {
  name                               = var.app_name
  cluster                            = aws_ecs_cluster.this.id
  task_definition                    = aws_ecs_task_definition.strengthgadget.arn
  desired_count                      = 1
  deployment_maximum_percent         = 200
  deployment_minimum_healthy_percent = 100

  capacity_provider_strategy {
    capacity_provider = aws_ecs_capacity_provider.this.name
    base              = 1
    weight            = 1
  }

  #  May need to comment out this block for troubleshooting
  load_balancer {
    target_group_arn = aws_lb_target_group.this.arn
    container_name   = var.app_name
    container_port   = 8080
  }

  deployment_controller {
    type = "ECS"
  }

  deployment_circuit_breaker {
    enable   = true
    rollback = true
  }

#  network_configuration {
#    assign_public_ip = false # public IP is only available for fargate tasks
#    subnets          = local.subnet_ids
#    security_groups  = [aws_security_group.strengthgadget_sg.id]
#  }

  depends_on = [
    aws_autoscaling_group.this,
  ]
}


resource "aws_ecs_task_definition" "strengthgadget" {
  family                   = var.app_name
#  network_mode             = "awsvpc"
  network_mode             = "bridge"
  requires_compatibilities = ["EC2"]

  container_definitions = jsonencode([
    {
      name  = var.app_name
      image = "piegarden/strengthgadget:${local.app_version}"

      logConfiguration = {
        logDriver = "awslogs"

        options = {
          "awslogs-region"        = var.aws_region
          "awslogs-group"         = aws_cloudwatch_log_group.backend_service.name
          "awslogs-stream-prefix" = "ecs"
        }
      }
      environment = [
        { "name" : "TF_VAR_environment", "value" : var.environment },
        { "name" : "TF_VAR_registration_email_from", "value" : var.registration_email_from },
        { "name" : "TF_VAR_registration_email_from_password", "value" : var.registration_email_from_password },
        { "name" : "TF_VAR_database_connection_string", "value" : var.database_connection_string },
        { "name" : "TF_VAR_redis_connection_string", "value" : var.redis_connection_string },
        { "name" : "TF_VAR_version", "value" : local.app_version },
        { "name" : "TF_VAR_database_root_ca", "value" : var.database_root_ca },
        { "name" : "TF_VAR_email_root_ca", "value" : var.email_root_ca },
        { "name" : "TF_VAR_trusted_ui_origin", "value" : "https://${var.domain_name}" },
        { "name" : "TF_VAR_redis_password", "value" : var.redis_password },
        { "name" : "TF_VAR_verification_excessive_retry_attempt_lockout_duration_in_seconds", "value" : "86400" },
        { "name" : "TF_VAR_allowed_verification_attempts_with_the_excessive_retry_lockout_window", "value" : "5" },
        { "name" : "TF_VAR_verification_code_validity_window_in_min", "value" : "30" },
        { "name" : "TF_VAR_window_length_in_seconds_for_the_number_of_allowed_verification_emails_before_lockout", "value" : "3600" },
        { "name" : "TF_VAR_window_length_in_seconds_for_the_number_of_allowed_login_attempts_before_lockout", "value" : "3600" },
        { "name" : "TF_VAR_allowed_login_attempts_before_triggering_lockout", "value" : "7" },
        { "name" : "TF_VAR_redis_user_private_key", "value" : var.redis_user_private_key },
        { "name" : "TF_VAR_redis_user_crt", "value" : var.redis_user_crt },
        { "name" : "TF_VAR_redis_ca_pem_part_one", "value" : var.redis_ca_pem_part_one },
        { "name" : "TF_VAR_redis_ca_pem_part_two", "value" : var.redis_ca_pem_part_two },
        { "name" : "TF_VAR_redis_ca_pem_part_three", "value" : var.redis_ca_pem_part_three },
        { "name" : "TF_VAR_redis_ca_pem_part_four", "value" : var.redis_ca_pem_part_four },
        { "name" : "TF_VAR_sentry_end_point", "value" : var.sentry_end_point },
      ],
      essential         = true
      memoryReservation = local.nodeSizeMemorySizeInMB
      portMappings      = [
        {
          containerPort = 8080
          hostPort      = 0 // zero means dynamic port
          protocol      = "tcp"
        }
      ]
    },
  ])
}
