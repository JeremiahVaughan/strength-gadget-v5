locals {
  env_vars = read_terragrunt_config(find_in_parent_folders("env.hcl"))
  env_name = local.env_vars.locals.env
  aws_region = get_env("TF_VAR_aws_region")
}

include "root" {
  path = find_in_parent_folders()
}

include "env" {
  path = "${get_terragrunt_dir()}/../../_env/cloudfront.hcl"
}

inputs = {
  env = local.env_name
  aws_region = local.aws_region
}
