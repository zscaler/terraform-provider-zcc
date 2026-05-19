terraform {
  required_providers {
    zcc = {
      source  = "zscaler/zcc"
      version = "~> 0.1.0"
    }
  }
}

provider "zcc" {}

data "zcc_predefined_ip_apps" "example" {
  name = "ExamplePredefinedApp"
}
