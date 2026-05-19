terraform {
  required_providers {
    zcc = {
      source  = "zscaler/zcc"
      version = "~> 0.1.0"
    }
  }
}
provider "zcc" {}

resource "zcc_forwarding_profile" "example" {
  name = "example-forwarding-profile"
}
