terraform {
  source = "../../modules/artifacts"
}

include "root" {
  path = find_in_parent_folders()
}

