resource "aws_ecr_repository" "this" {
  name = "${var.app_name}-artifacts"
}

resource "aws_ssm_parameter" "ecr_repo_url" {
  name  = "ecr-repo-url"
  type  = "String"
  value = aws_ecr_repository.this.repository_url
}

output "ecr_url" {
  value = aws_ecr_repository.this.repository_url
}
