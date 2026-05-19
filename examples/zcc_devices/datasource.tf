terraform {
  required_providers {
    zcc = {
      source  = "zscaler/zcc"
      version = "~> 0.1.0"
    }
  }
}

provider "zcc" {}

data "zcc_devices" "all" {}

data "zcc_devices" "filtered" {
  username = "jdoe"
  os_type  = "windows"
}
