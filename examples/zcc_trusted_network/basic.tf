terraform {
  required_providers {
    zcc = {
      source  = "zscaler/zcc"
      version = "~> 0.1.0"
    }
  }
}

provider "zcc" {}

resource "zcc_trusted_network" "example" {
  network_name   = "example-trusted-network"
  active         = true
  condition_type = 0 # Must match GET response; API may return 0 or 1
  trusted_subnets = "192.0.2.0/24"
}
