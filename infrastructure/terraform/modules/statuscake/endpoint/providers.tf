terraform {
  required_providers {
    statuscake = {
      source = "StatusCakeDev/statuscake"
      version = "~> 2.2.2" # This version must stay in sync with the root terragrunt.hcl file
    }
  }
}