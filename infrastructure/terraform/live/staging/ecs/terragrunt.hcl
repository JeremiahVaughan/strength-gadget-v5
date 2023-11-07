locals {
  env_vars = read_terragrunt_config(find_in_parent_folders("env.hcl"))
  env_name = local.env_vars.locals.env
  aws_region = get_env("TF_VAR_aws_region")
}

include "root" {
  path = find_in_parent_folders()
}

include "env" {
  path = "${get_terragrunt_dir()}/../../_env/ecs.hcl"
}

dependency "artifacts" {
  config_path = "../../artifacts"
}

inputs = {
  env = local.env_name
  pub_ssh_key_path = get_env("PUB_SSH_KEY_PATH")
  ecr_url = dependency.artifacts.outputs.ecr_url
}
