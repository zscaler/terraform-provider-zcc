terraform {
  required_providers {
    zcc = {
      source  = "zscaler/zcc"
      version = "~> 0.1.0"
    }
  }
}

provider "zcc" {}

data "zcc_admin_roles" "example" {
  role_name = "Super Admin"
}
