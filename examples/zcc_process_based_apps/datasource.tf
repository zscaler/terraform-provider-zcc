terraform {
  required_providers {
    zcc = {
      source  = "zscaler/zcc"
      version = "~> 0.1.0"
    }
  }
}

provider "zcc" {}

data "zcc_process_based_apps" "example" {
  name = "ExampleProcessApp"
}
