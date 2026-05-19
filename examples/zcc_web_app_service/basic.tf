terraform {
  required_providers {
    zcc = {
      source  = "zscaler/zcc"
      version = "~> 0.1.0"
    }
  }
}

provider "zcc" {}

# app_name must match an existing ZCC web app service (bypass app) in your tenant.
resource "zcc_web_app_service" "example" {
  app_name = "ExampleBypassApp"
}
