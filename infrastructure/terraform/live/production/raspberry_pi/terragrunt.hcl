locals {
  env_vars = read_terragrunt_config(find_in_parent_folders("env.hcl"))
  env_name = local.env_vars.locals.env
}

include "root" {
  path = find_in_parent_folders()
}

include "env" {
  path = "${get_terragrunt_dir()}/../../_env/raspberry_pi.hcl"
}


inputs = {
  environment = local.env_name
  domain_name = "strengthgadget.com"
  static_ip = "173.197.226.162"
}
