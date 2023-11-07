resource "aws_appautoscaling_policy" "cpu_scaling_policy" {
  name               = "${var.service_name}-cpu-scaler"
  policy_type        = "TargetTrackingScaling"
  resource_id        = var.resource_id
  scalable_dimension = var.scalable_dimension
  service_namespace  = var.service_namespace

  target_tracking_scaling_policy_configuration {
    target_value       = 80.0
    scale_in_cooldown  = 300
    scale_out_cooldown = 300
    predefined_metric_specification {
      predefined_metric_type = "ECSServiceAverageCPUUtilization"
    }
  }
}

resource "aws_appautoscaling_policy" "memory_scaling_policy" {
  name               = "${var.service_name}-memory-scaler"
  policy_type        = "TargetTrackingScaling"
  resource_id        = var.resource_id
  scalable_dimension = var.scalable_dimension
  service_namespace  = var.service_namespace

  target_tracking_scaling_policy_configuration {
    target_value       = 80.0
    scale_in_cooldown  = 300
    scale_out_cooldown = 300
    predefined_metric_specification {
      predefined_metric_type = "ECSServiceAverageMemoryUtilization"
    }
  }
}