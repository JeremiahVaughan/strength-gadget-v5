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
      image = "${var.ecr_url}:${local.app_version}"

      logConfiguration = {
        logDriver = "awslogs"

        options = {
          "awslogs-region"        = var.aws_region
          "awslogs-group"         = aws_cloudwatch_log_group.backend_service.name
          "awslogs-stream-prefix" = "ecs"
        }
      }
      environment = [
        { "name" : "REGISTRATION_EMAIL_FROM", "value" : var.registration_email_from },
        { "name" : "REGISTRATION_EMAIL_FROM_PASSWORD", "value" : var.registration_email_from_password },
        { "name" : "DATABASE_CONNECTION_STRING", "value" : var.database_connection_string },
        { "name" : "REDIS_CONNECTION_STRING", "value" : var.redis_connection_string },
        { "name" : "VERSION", "value" : local.app_version },
        { "name" : "DATABASE_ROOT_CA", "value" : var.database_root_ca },
        { "name" : "EMAIL_ROOT_CA", "value" : var.email_root_ca },
        { "name" : "TRUSTED_UI_ORIGIN", "value" : "https://${var.domain_name}" },
        { "name" : "REDIS_PASSWORD", "value" : var.redis_password },
        { "name" : "VERIFICATION_EXCESSIVE_RETRY_ATTEMPT_LOCKOUT_DURATION_IN_SECONDS", "value" : "86400" },
        { "name" : "ALLOWED_VERIFICATION_ATTEMPTS_WITH_THE_EXCESSIVE_RETRY_LOCKOUT_WINDOW", "value" : "5" },
        { "name" : "VERIFICATION_CODE_VALIDITY_WINDOW_IN_MIN", "value" : "30" },
        {
          "name" : "WINDOW_LENGTH_IN_SECONDS_FOR_THE_NUMBER_OF_ALLOWED_VERIFICATION_EMAILS_BEFORE_LOCKOUT",
          "value" : "3600"
        },
        {
          "name" : "WINDOW_LENGTH_IN_SECONDS_FOR_THE_NUMBER_OF_ALLOWED_LOGIN_ATTEMPTS_BEFORE_LOCKOUT", "value" : "3600"
        },
        { "name" : "ALLOWED_LOGIN_ATTEMPTS_BEFORE_TRIGGERING_LOCKOUT", "value" : "7" },
        { "name" : "REDIS_USER_PRIVATE_KEY", "value" : var.redis_user_private_key },
        { "name" : "REDIS_USER_CRT", "value" : var.redis_user_crt },
        { "name" : "REDIS_CA_PEM_PART_ONE", "value" : var.redis_ca_pem_part_one },
        { "name" : "REDIS_CA_PEM_PART_TWO", "value" : var.redis_ca_pem_part_two },
        { "name" : "REDIS_CA_PEM_PART_THREE", "value" : var.redis_ca_pem_part_three },
        { "name" : "REDIS_CA_PEM_PART_FOUR", "value" : var.redis_ca_pem_part_four },
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
