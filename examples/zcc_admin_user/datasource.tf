terraform {
  required_providers {
    zcc = {
      source  = "zscaler/zcc"
      version = "~> 0.1.0"
    }
  }
}
provider "zcc" {}

data "zcc_admin_user" "example" {
  user_name = "admin@example.com"
}
