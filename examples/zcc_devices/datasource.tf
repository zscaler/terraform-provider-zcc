data "zcc_devices" "all" {}

data "zcc_devices" "filtered" {
  username = "jdoe"
  os_type  = "windows"
}
